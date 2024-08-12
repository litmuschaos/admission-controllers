package pods

import (
	"context"
	"fmt"
	"github.com/litmuschaos/admission-controller/pkg/clients"
	"github.com/litmuschaos/admission-controller/pkg/utils"
	v1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

const (
	podNameKey                     = "authentication.kubernetes.io/pod-name"
	jobControllerServiceAccount    = "system:serviceaccount:kube-system:job-controller"
	controllerServiceAccountPrefix = "system:serviceaccount:kube-system"
)

func validateOriginServiceAccount(serviceAccount string) (bool, string) {
	if utils.WebHookFilters.AllowedOriginServiceAccount.AllowedAll {
		return true, ""
	}

	if serviceAccount == jobControllerServiceAccount {
		return true, ""
	}

	serviceAccountList := strings.Split(serviceAccount, ":")
	if len(serviceAccountList) != 4 {
		return false, fmt.Sprintf("%v serviceAccount is not in a valid format 'system:serviceaccount:<ns><name>'")
	}

	for _, v := range utils.WebHookFilters.AllowedOriginServiceAccount.AllowedList {
		if utils.MatchRegex(v, serviceAccountList[3]) {
			return true, ""
		}
	}

	return false, ""
}

func validateOriginImage(serviceAccount string, pod *corev1.Pod, extras map[string]v1.ExtraValue, clients clients.ClientSets) (bool, string) {
	if utils.WebHookFilters.AllowedOriginImages.AllowedAll {
		return true, ""
	}

	if strings.Contains(serviceAccount, controllerServiceAccountPrefix) {
		return true, ""
	}

	return validateOriginPodImage(pod.Namespace, extras, clients)
}

func validateOriginPodImage(namespace string, extras map[string]v1.ExtraValue, clients clients.ClientSets) (bool, string) {
	if extras == nil {
		return false, fmt.Sprintf("could not get the origin pod name")
	}

	podName, ok := extras[podNameKey]
	if !ok || len(podName) == 0 {
		return false, fmt.Sprintf("could not verify the origin pod")
	}

	pod, err := clients.KubeClient.CoreV1().Pods(namespace).Get(context.Background(), podName[0], metav1.GetOptions{})
	if err != nil {
		return false, fmt.Sprintf("could not get origin pod: %v", err)
	}

	for _, c := range pod.Spec.Containers {
		for _, v := range utils.WebHookFilters.AllowedOriginImages.AllowedList {
			if utils.MatchRegex(v, c.Image) {
				return true, ""
			}
		}
	}

	return false, ""
}

func originFromTerminal(serviceAccount string) bool {
	if strings.Contains(serviceAccount, "system:serviceaccount") {
		return false
	}
	return true
}
