/*
Copyright 2019 The LitmusChaos Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package webhook

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"k8s.io/api/admission/v1beta1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	litmuschaosv1alpha1 "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	// (https://github.com/kubernetes/kubernetes/issues/57982)
	defaulter = runtime.ObjectDefaulter(runtimeScheme)
)

// Skip validation in special namespaces, i.e. in kube-system and kube-public
// namespaces the validation will be skipped
var (
	ignoredNamespaces = []string{
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}
	snapshotAnnotation = "snapshot.alpha.kubernetes.io/snapshot"
)

// webhook implements a validating webhook.
type webhook struct {
	//  Server defines parameters for running an golang HTTP server.
	Server *http.Server

	// kubeClient is a standard kubernetes clientset
	kubeClient kubernetes.Clientset

	litmusClient litmuschaosv1alpha1.Clientset
}

// Parameters are server configures parameters
type Parameters struct {
	// Port is webhook server port
	Port int
	//CertFile is path to the x509 certificate for https
	CertFile string
	//KeyFile is path to the x509 private key matching `CertFile`
	KeyFile string
}

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1beta1.AddToScheme(runtimeScheme)
	// defaulting with webhooks:
	// https://github.com/kubernetes/kubernetes/issues/57982
	_ = v1.AddToScheme(runtimeScheme)
}

// New creates a new instance of a webhook. Prior to
// invoking this function, InitValidationServer function must be called to
// set up secret (for TLS certs) k8s resource. This function runs forever.
func New(p Parameters, kubeClient kubernetes.Clientset,
	litmusClient litmuschaosv1alpha1.Clientset) (
	*webhook, error) {

	admNamespace, err := getLitmusNamespace()
	if err != nil {
		return nil, err
	}

	// Fetch certificate secret information
	certSecret, err := GetSecret(admNamespace, validatorSecret, &kubeClient)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read secret(%s) object %v",
			validatorSecret,
			err,
		)
	}

	// extract cert information from the secret object
	certBytes, ok := certSecret.Data[appCrt]
	if !ok {
		return nil, fmt.Errorf(
			"%s value not found in %s secret",
			appCrt,
			validatorSecret,
		)
	}
	keyBytes, ok := certSecret.Data[appKey]
	if !ok {
		return nil, fmt.Errorf(
			"%s value not found in %s secret",
			appKey,
			validatorSecret,
		)
	}

	signingCertBytes, ok := certSecret.Data[rootCrt]
	if !ok {
		return nil, fmt.Errorf(
			"%s value not found in %s secret",
			rootCrt,
			validatorSecret,
		)
	}

	certPool := x509.NewCertPool()
	ok = certPool.AppendCertsFromPEM(signingCertBytes)
	if !ok {
		return nil, fmt.Errorf("failed to parse root certificate")
	}

	sCert, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		return nil, err
	}

	wh := &webhook{
		Server: &http.Server{
			Addr:      fmt.Sprintf(":%v", p.Port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{sCert}},
		},
		kubeClient:   kubeClient,
		litmusClient: litmusClient,
		//snapClientSet: snapClient,
	}
	return wh, nil
}

func admissionRequired(ignoredList []string, metadata *metav1.ObjectMeta) bool {
	// skip special kubernetes system namespaces
	for _, namespace := range ignoredList {
		if metadata.Namespace == namespace {
			klog.V(4).Infof("Skip validation for %v for it's in special namespace:%v", metadata.Name, metadata.Namespace)
			return false
		}
	}
	return true
}

func validationRequired(ignoredList []string, metadata *metav1.ObjectMeta) bool {
	required := admissionRequired(ignoredList, metadata)
	klog.V(4).Infof("Validation policy for %v/%v: required:%v", metadata.Namespace, metadata.Name, required)
	return required
}

func (wh *webhook) validateChaosEngineCreateUpdate(req *v1beta1.AdmissionRequest) *v1beta1.AdmissionResponse {
	response := &v1beta1.AdmissionResponse{}
	response.Allowed = true
	var chaosEngine v1alpha1.ChaosEngine
	err := json.Unmarshal(req.Object.Raw, &chaosEngine)
	if err != nil {
		klog.Errorf("Could not unmarshal raw object: %v, %v", err, req.Object.Raw)
		response.Allowed = false
		response.Result = &metav1.Status{
			Status:  metav1.StatusFailure,
			Code:    http.StatusBadRequest,
			Reason:  metav1.StatusReasonBadRequest,
			Message: err.Error(),
		}
		return response
	}

	validationStatus, err := wh.validateChaosTarget(&chaosEngine)
	if validationStatus {
		klog.V(2).Infof("Validation Successful for ChaosEngine: %v", chaosEngine.Name)
		response.Allowed = true
		return response
	}

	klog.V(2).Infof("Validation Failed for ChaosEngine: %v", chaosEngine.Name)
	response.Allowed = false
	response.Result = &metav1.Status{
		Status:  metav1.StatusFailure,
		Code:    http.StatusBadRequest,
		Reason:  metav1.StatusReasonBadRequest,
		Message: err.Error(),
	}

	return response
}

// validate validates the chaosengine create, update request
func (wh *webhook) validate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	var (
		resourceName string
	)
	klog.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v (%v) UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, resourceName, req.UID, req.Operation, req.UserInfo)
	response := &v1beta1.AdmissionResponse{}
	response.Allowed = true
	switch req.Kind.Kind {

	case "ChaosEngine":
		klog.V(0).Infof("Starting to validate, admission webhook request for type: %s", req.Kind.Kind)
		return wh.validateChaosEngine(ar)

	default:
		return response
	}

}

func (wh *webhook) validateChaosEngine(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	response := &v1beta1.AdmissionResponse{}
	response.Allowed = true

	if req.Operation == v1beta1.Create || req.Operation == v1beta1.Update {
		return wh.validateChaosEngineCreateUpdate(req)
	}
	return response
}

// Serve method for webhook server, handles http requests for webhooks
func (wh *webhook) Serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		klog.Error("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("Content-Type=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		klog.Errorf("Can't decode body: %v", err)
		admissionResponse = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		if r.URL.Path == "/validate" {
			admissionResponse = wh.validate(&ar)
		}
	}

	admissionReview := v1beta1.AdmissionReview{}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		klog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	klog.V(5).Infof("Ready to write reponse ...")
	if _, err := w.Write(resp); err != nil {
		klog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}

func (wh *webhook) validateChaosTarget(chaosEngine *v1alpha1.ChaosEngine) (bool, error) {
	switch resourceType := strings.ToLower(chaosEngine.Spec.Appinfo.AppKind); resourceType {
	case "deployment", "deployments":
		return validateDeployment(chaosEngine.Spec.Appinfo, wh.kubeClient)
	case "statefulset", "statefulsets":
		return validateStatefulSet(chaosEngine.Spec.Appinfo, wh.kubeClient)
	case "daemonset", "daemonsets":
		return validateDaemonSet(chaosEngine.Spec.Appinfo, wh.kubeClient)
	default:
		return false, fmt.Errorf("Unable to validate resourceType: %v, unsupported resource", resourceType)
	}
}

func validateDeployment(appInfo v1alpha1.ApplicationParams, kubeClient kubernetes.Clientset) (bool, error) {
	deployments, err := kubeClient.AppsV1().Deployments(appInfo.Appns).List(metav1.ListOptions{
		LabelSelector: appInfo.Applabel,
	})
	if err != nil {
		return false, fmt.Errorf("unable to list deployments, please provide a suitable RBAC with apiGroup 'apps', and resource 'deployments' and verb 'list' , or remove this deployment")
	}
	if len(deployments.Items) == 0 {
		return false, fmt.Errorf("unable to find deployment specified in ChaosEngine")
	}
	return true, nil

}

func validateStatefulSet(appInfo v1alpha1.ApplicationParams, kubeClient kubernetes.Clientset) (bool, error) {
	statefulsets, err := kubeClient.AppsV1().StatefulSets(appInfo.Appns).List(metav1.ListOptions{
		LabelSelector: appInfo.Applabel,
	})
	if err != nil {
		return false, fmt.Errorf("unable to list statefulsets, please provide a suitable RBAC with apiGroup 'apps', resource 'statefulsets' and verb 'list' , or remove this deployment")
	}
	if len(statefulsets.Items) == 0 {
		return false, fmt.Errorf("unable to find statefulset specified in ChaosEngine")
	}
	return true, nil

}

func validateDaemonSet(appInfo v1alpha1.ApplicationParams, kubeClient kubernetes.Clientset) (bool, error) {
	daemonsets, err := kubeClient.AppsV1().DaemonSets(appInfo.Appns).List(metav1.ListOptions{
		LabelSelector: appInfo.Applabel,
	})
	if err != nil {
		return false, fmt.Errorf("unable to fetch daemonsets, please provide a suitable RBAC with apiGroup 'apps', and resource 'daemonsets' and verb 'list', or remove this deployment")
	}
	if len(daemonsets.Items) == 0 {
		return false, fmt.Errorf("unable to find daemonset specified in ChaosEngine")
	}
	return true, nil

}
