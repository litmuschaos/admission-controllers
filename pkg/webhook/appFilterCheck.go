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

func (wh *webhook) ValidateChaosExperimentInApplicationNamespaces(chaosEngine *v1alpha1.ChaosEngine) error {
	for _, experiment := range chaosEngine.Spec.Experiments {
		if err := wh.checkExperimentInNamespace(experiment.Name, chaosEngine.Spec.Appinfo.Appns); err != nil {
			return err
		}
	}
	return nil
}

func (wh *webhook) ValidateChaosExperimentsConfigMaps(chaosEngine *v1alpha1.ChaosEngine) error {
	configMapErrors := make([]string, 0)
	for _, experiment := range chaosEngine.Spec.Experiments {
		for _, expectedConfigMap := range experiment.Spec.Components.ConfigMaps {
			_, err := wh.kubeClient.CoreV1().ConfigMaps(chaosEngine.Spec.Appinfo.Appns).Get(expectedConfigMap.Name, metav1.GetOptions{})
			if err != nil {
				configMapErrors = append(configMapErrors,
					fmt.Sprintf("Unable to find ConfigMap %s needed for ChaosExperiment %s, please check the following error: %v", expectedConfigMap.Name, experiment.Name, err))
			}
		}
	}
	if len(configMapErrors) == 0 {
		return nil
	}
	return fmt.Errorf(strings.Join(configMapErrors, "\n"))
}

func (wh *webhook) ValidateChaosExperimentsSecrets(chaosEngine *v1alpha1.ChaosEngine) error {
	secretsErrors := make([]string, 0)
	for _, experiment := range chaosEngine.Spec.Experiments {
		for _, expectedSecret := range experiment.Spec.Components.Secrets {
			_, err := wh.kubeClient.CoreV1().Secrets(chaosEngine.Spec.Appinfo.Appns).Get(expectedSecret.Name, metav1.GetOptions{})
			if err != nil {
				secretsErrors = append(secretsErrors,
					fmt.Sprintf("Unable to find Secret %s needed for ChaosExperiment %s, please check the following error: %v", expectedSecret.Name, experiment.Name, err))
			}
		}
	}
	if len(secretsErrors) == 0 {
		return nil
	}
	return fmt.Errorf(strings.Join(secretsErrors, "\n"))
}

func (wh *webhook) ValidateChaosTarget(chaosEngine *v1alpha1.ChaosEngine) error {
	switch resourceType := strings.ToLower(chaosEngine.Spec.Appinfo.AppKind); resourceType {
	case "deployment", "deployments":
		return validateDeployment(chaosEngine.Spec.Appinfo, wh.kubeClient)
	case "statefulset", "statefulsets":
		return validateStatefulSet(chaosEngine.Spec.Appinfo, wh.kubeClient)
	case "daemonset", "daemonsets":
		return validateDaemonSet(chaosEngine.Spec.Appinfo, wh.kubeClient)
	default:
		return fmt.Errorf("Unable to validate resourceType: %v, unsupported resource", resourceType)
	}
}

func validateDeployment(appInfo v1alpha1.ApplicationParams, kubeClient kubernetes.Interface) error {
	deployments, err := kubeClient.AppsV1().Deployments(appInfo.Appns).List(metav1.ListOptions{
		LabelSelector: appInfo.Applabel,
	})
	if err != nil {
		return fmt.Errorf("unable to list deployments with matching labels, please check the following error: %v", err)
	}
	if len(deployments.Items) == 0 {
		return fmt.Errorf("unable to find deployment specified in ChaosEngine")
	}

	for _, deployment := range deployments.Items {
		if err := validatePodTemplateSpec(appInfo, deployment.Spec.Template); err != nil {
			return fmt.Errorf("unable to find labels in pod template of deployment provided")
		}
	}

	return nil

}

func validateStatefulSet(appInfo v1alpha1.ApplicationParams, kubeClient kubernetes.Interface) error {
	statefulsets, err := kubeClient.AppsV1().StatefulSets(appInfo.Appns).List(metav1.ListOptions{
		LabelSelector: appInfo.Applabel,
	})
	if err != nil {
		return fmt.Errorf("unable to list statefulsets with matching labels, please check the following error: %v", err)
	}
	if len(statefulsets.Items) == 0 {
		return fmt.Errorf("unable to find statefulset specified in ChaosEngine")
	}

	for _, statefulset := range statefulsets.Items {
		if err := validatePodTemplateSpec(appInfo, statefulset.Spec.Template); err != nil {
			return fmt.Errorf("unable to find labels in pod template of statefulset provided")
		}
	}
	return nil

}

func validateDaemonSet(appInfo v1alpha1.ApplicationParams, kubeClient kubernetes.Interface) error {
	daemonsets, err := kubeClient.AppsV1().DaemonSets(appInfo.Appns).List(metav1.ListOptions{
		LabelSelector: appInfo.Applabel,
	})
	if err != nil {
		return fmt.Errorf("unable to list daemonsets with matching labels, please check the following error: %v", err)
	}
	if len(daemonsets.Items) == 0 {
		return fmt.Errorf("unable to find daemonset specified in ChaosEngine")
	}

	for _, daemonset := range daemonsets.Items {
		if err := validatePodTemplateSpec(appInfo, daemonset.Spec.Template); err != nil {
			return fmt.Errorf("unable to find labels in pod template of daemonset provided")
		}
	}

	return nil

}

func validatePodTemplateSpec(appInfo v1alpha1.ApplicationParams, podTemplateSpec corev1.PodTemplateSpec) error {
	appLabel := strings.Split(appInfo.Applabel, "=")
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

func (wh *webhook) checkExperimentInNamespace(experimentName, namespace string) error {
	_, err := wh.litmusClient.LitmuschaosV1alpha1().ChaosExperiments(namespace).Get(experimentName, metav1.GetOptions{})
	return err
}
