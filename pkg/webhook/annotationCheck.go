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
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
)

func (wh *webhook) ValidateChaosTarget(chaosEngine *v1alpha1.ChaosEngine) (bool, error) {
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

func validatePodTemplateSpec(engine *v1alpha1.ChaosEngine, podTemplateSpec corev1.PodTemplateSpec) error {
	appLabel := strings.Split(engine.Spec.Appinfo.Applabel, "=")
	labelFound := checkLabelInMap(appLabel, podTemplateSpec.Labels)
	if !labelFound {
		return fmt.Errorf("Unable to validate appLabel provided in ChaosEngine in PodTemplateSpec")
	}
	return nil
}

func checkLabelInMap(toCheck []string, labels map[string]string) bool {
	for key, value := range labels {
		if key == toCheck[0] {
			return value == toCheck[1]
		}
	}
	return false
}
