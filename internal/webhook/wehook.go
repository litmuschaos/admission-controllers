package webhook

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/litmuschaos/admission-controller/pkg/clients"
	"github.com/litmuschaos/admission-controller/pkg/utils"
	"k8s.io/api/admissionregistration/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

const (
	controllerLabel = "litmuschaos.io/component-name=admission-controller"
)

func ManageDependencies(clients clients.ClientSets) (*tls.Certificate, error) {
	// get controller name
	deployName, err := getControllerReferences(clients)
	if err != nil {
		return nil, err
	}

	var serviceName = fmt.Sprintf("%s-service", deployName)

	// create certs
	rootCA, certSecret, err := createCerts(serviceName)
	if err != nil {
		return nil, err
	}

	// create validation webhook
	if err := createValidatingWebhookConfiguration(serviceName, EncodeCertPEM(rootCA), clients); err != nil {
		return nil, err
	}

	return certSecret, nil
}

func createValidatingWebhookConfiguration(serviceName string, signingCert []byte, clients clients.ClientSets) error {
	webhook, err := clients.KubeClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(context.Background(), fmt.Sprintf("pod-%s-%s-validator", utils.WebHookFilters.TargetServiceAccount, utils.WebHookFilters.ChaosNamespace), metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return createValidator(serviceName, signingCert, clients)
		}
		return err
	}

	if len(webhook.Webhooks) != 0 {
		webhook.Webhooks[0].ClientConfig.CABundle = signingCert
	}
	_, err = clients.KubeClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Update(context.Background(), webhook, metav1.UpdateOptions{})
	return err
}

func createValidator(serviceName string, signingCert []byte, clients clients.ClientSets) error {
	var (
		sideEffect       = v1.SideEffectClassNone
		validatePodsPath = "/validate/pods"
		timeout          = int32(5)
		failPolicy       = v1.Ignore
	)

	webhookHandler := v1.ValidatingWebhook{
		Name: fmt.Sprintf("pod-validator.%s.svc", utils.WebHookFilters.TargetServiceAccount),
		NamespaceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"kubernetes.io/metadata.name": utils.WebHookFilters.ChaosNamespace,
			},
		},
		AdmissionReviewVersions: []string{"v1"},
		SideEffects:             &sideEffect,
		FailurePolicy:           &failPolicy,
		Rules: []v1.RuleWithOperations{{
			Operations: []v1.OperationType{
				v1.Create,
			},
			Rule: v1.Rule{
				APIGroups:   []string{""},
				APIVersions: []string{"v1"},
				Resources:   []string{"pods"},
			},
		},
		},
		ClientConfig: v1.WebhookClientConfig{
			Service: &v1.ServiceReference{
				Namespace: utils.WebHookFilters.ChaosNamespace,
				Name:      serviceName,
				Path:      &validatePodsPath,
			},
			CABundle: signingCert,
		},
		TimeoutSeconds: &timeout,
	}

	validator := &v1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("pod-%s-%s-validator", utils.WebHookFilters.TargetServiceAccount, utils.WebHookFilters.ChaosNamespace),
			Labels: map[string]string{
				"app":                           "admission-controller",
				"litmuschaos.io/component-name": "admission-controller",
			},
		},
		Webhooks: []v1.ValidatingWebhook{webhookHandler},
	}

	_, err := clients.KubeClient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Create(context.Background(), validator, metav1.CreateOptions{})
	return err
}

func createCerts(serviceName string) (*x509.Certificate, *tls.Certificate, error) {
	// Create a signing certificate
	caKeyPair, err := NewCA(fmt.Sprintf("%s-ca", serviceName))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create root-ca: %v", err)
	}

	// Create app certs signed through the certificate created above
	apiServerKeyPair, err := NewServerKeyPair(
		caKeyPair,
		strings.Join([]string{serviceName, utils.WebHookFilters.ChaosNamespace, "svc"}, "."),
		serviceName,
		utils.WebHookFilters.ChaosNamespace,
		"cluster.local",
		[]string{},
		[]string{},
	)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to create server key pair: %v", err)
	}

	sCert, err := tls.X509KeyPair(EncodeCertPEM(apiServerKeyPair.Cert), EncodePrivateKeyPEM(apiServerKeyPair.Key))
	if err != nil {
		return nil, nil, err
	}

	return caKeyPair.Cert, &sCert, nil
}

func getControllerReferences(clients clients.ClientSets) (string, error) {
	deploy, err := clients.KubeClient.AppsV1().Deployments(utils.WebHookFilters.ChaosNamespace).List(context.Background(), metav1.ListOptions{LabelSelector: controllerLabel})
	if err != nil {
		return "", fmt.Errorf("failed to list admission deployment: %s", err.Error())
	}
	for _, d := range deploy.Items {
		if len(d.Name) != 0 {
			return d.Name, nil

		}
	}
	return "", fmt.Errorf("failed to get deployment name")
}
