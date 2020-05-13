// Copyright © 2018-2019 The LitmusChaos Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package webhook

import (
	"encoding/json"
	"testing"

	"k8s.io/api/admission/v1beta1"
	"k8s.io/client-go/kubernetes/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAdmissionRequired(t *testing.T) {
	cases := []struct {
		Name, Namespace string
		want            bool
	}{
		{"default-policy", "test-namespace", true},
		{"default-policy", "default", true},
		{"no-policy-in-kube-system", "kube-system", false},
		{"no-policy-in-kube-public", "kube-public", false},
	}

	for _, c := range cases {
		meta := &metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: c.Namespace,
		}

		if got := validationRequired(ignoredNamespaces, meta); got != c.want {
			t.Errorf("admissionRequired(%v)  got %v want %v", meta.Name, got, c.want)
		}
	}
}

func TestValidate(t *testing.T) {
	//fakeKubeClient := fake.NewSimpleClientset()
	//fakeLitmusClient := litmusFakeClientset.NewSimpleClientset()
	wh := webhook{
		kubeClient: *fake.NewSimpleClientset(),
		//litmusClient: *fakeLitmusClient,
	}
	cases := map[string]struct {
		testAdmissionRev *v1beta1.AdmissionReview
		expectedResponse bool
	}{
		"ChaosEngine Create request": {
			testAdmissionRev: &v1beta1.AdmissionReview{
				Request: &v1beta1.AdmissionRequest{
					Kind: metav1.GroupVersionKind{
						Kind:    "ChaosEngine",
						Group:   "litmuschaos.io",
						Version: "v1alpha1",
					},
					Operation: v1beta1.Create,
				},
			},
			expectedResponse: true,
		},
	}
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			resp := wh.validate(test.testAdmissionRev)
			if resp.Allowed != test.expectedResponse {
				t.Errorf("validate request failed got: '%v' expected: '%v'", resp.Allowed, test.expectedResponse)
			}
		})
	}
}

func serialize(v interface{}) []byte {
	bytes, _ := json.Marshal(v)
	return bytes
}

// func TestValidateChaosEngineCreateUpdate(t *testing.T) {
// 	wh := webhook{
// 		kubeClient:
// 	}
// 	cases := map[string]struct {
// 		fakeChaosEngine  v1alpha1.ChaosEngine
// 		expectedResponse bool
// 	}{
// 		"Empty ChaosEngine Create request": {
// 			fakeChaosEngine:  v1alpha1.ChaosEngine{},
// 			expectedResponse: true,
// 		},
// 		"Valid ChaosEngine Create Request": {
// 			fakeChaosEngine: v1alpha1.ChaosEngine{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "fakeChaosEngine",
// 					Namespace: "fakeNamespace",
// 				},
// 				Spec: v1alpha1.ChaosEngineSpec{
// 					Appinfo: v1alpha1.ApplicationParams{
// 						AppKind:  "deployment",
// 						Appns:    "",
// 						Applabel: "",
// 					},
// 				},
// 			},
// 			// fakePVC: corev1.PersistentVolumeClaim{
// 			// 	ObjectMeta: metav1.ObjectMeta{
// 			// 		Annotations: fakepvcAnnotation,
// 			// 	},
// 			// 	Spec: corev1.PersistentVolumeClaimSpec{
// 			// 		VolumeName: "pvc-1",
// 			// 	},
// 			// },
// 			expectedResponse: true,
// 		},
// 	}
// 	for _, test := range cases {
// 		webhookReq := &v1beta1.AdmissionRequest{
// 			Operation: v1beta1.Create,
// 			Object: runtime.RawExtension{
// 				Raw: serialize(test.fakeChaosEngine),
// 			},
// 		}
// 		resp := wh.validateChaosEngineCreateUpdate(webhookReq)
// 		if resp.Allowed != test.expectedResponse {
// 			t.Errorf("validate request failed got: '%v' expected: '%v'", resp.Allowed, test.expectedResponse)
// 		}
// 	}
// }
