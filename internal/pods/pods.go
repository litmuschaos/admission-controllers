package pods

import (
	"encoding/json"
	"fmt"
	"github.com/litmuschaos/admission-controller/internal/hook"
	"github.com/litmuschaos/admission-controller/pkg/clients"
	"github.com/litmuschaos/admission-controller/pkg/utils"
	"k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

// NewValidationHook creates a new instance of pods validation hook
func NewValidationHook(clients clients.ClientSets) hook.Hook {
	return hook.Hook{
		Create: validateCreate(clients),
	}
}

func validateCreate(clients clients.ClientSets) hook.AdmitFunc {
	return func(r *v1.AdmissionRequest) (*hook.Result, error) {
		pod, err := parsePod(r.Object.Raw)
		if err != nil {
			return &hook.Result{Msg: err.Error()}, nil
		}

		if !targetFilter(pod) {
			return &hook.Result{Allowed: true}, nil
		}

		if originFromTerminal(r.UserInfo.Username) {
			return &hook.Result{
				Allowed: false,
				Msg:     fmt.Sprintf("request is not initiated by a serviceAccount"),
			}, nil
		}

		// check for destination images
		for _, c := range pod.Spec.Containers {
			if !validateDestinationImage(c.Image) {
				return &hook.Result{
					Allowed: false,
					Msg:     fmt.Sprintf("%v image of %v pod doesn't met allowed image criteria", c.Image, pod.Name),
				}, nil
			}
		}

		// check for origin allowed service account
		allowed, msg := validateOriginServiceAccount(r.UserInfo.Username)
		if !allowed {
			return &hook.Result{
				Allowed: false,
				Msg:     fmt.Sprintf("origin serviceAccount doesn't met allowed serviceAccount criteria %v", msg),
			}, nil
		}

		// check for origin allowed images
		allowed, msg = validateOriginImage(r.UserInfo.Username, pod, r.UserInfo.Extra, clients)
		if !allowed {
			return &hook.Result{
				Allowed: false,
				Msg:     fmt.Sprintf("origin image doesn't met allowed image criteria: %v", msg),
			}, nil
		}

		return &hook.Result{Allowed: true}, nil
	}
}

func parsePod(object []byte) (*corev1.Pod, error) {
	var pod corev1.Pod
	if err := json.Unmarshal(object, &pod); err != nil {
		return nil, err
	}

	return &pod, nil
}

func targetFilter(pod *corev1.Pod) bool {
	// check service account
	if pod.Spec.ServiceAccountName == utils.WebHookFilters.TargetServiceAccount {
		return true
	}
	return false
}

func validateDestinationImage(destinationImage string) bool {
	if utils.WebHookFilters.AllowedTargetImages.AllowedAll {
		return true
	}

	for _, image := range utils.WebHookFilters.AllowedTargetImages.AllowedList {
		if utils.MatchRegex(image, destinationImage) {
			return true
		}
	}
	return false
}
