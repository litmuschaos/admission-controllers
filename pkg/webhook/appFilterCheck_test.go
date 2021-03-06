/*
Copyright 2019 LitmusChaos Authors

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
	"testing"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	fakelitmus "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

var testNamespace = "test-ns"

func TestValidateChaosExperimentsConfigMaps(t *testing.T) {
	var tests = []struct {
		description   string
		k8sObjects    []runtime.Object
		chaosEngine   v1alpha1.ChaosEngine
		isErrExpected bool
	}{
		{
			description: "Validation fails when none of the specified ConfigMap is in the Cluster.",
			k8sObjects:  []runtime.Object{},
			chaosEngine: v1alpha1.ChaosEngine{
				Spec: v1alpha1.ChaosEngineSpec{
					Appinfo: v1alpha1.ApplicationParams{
						Appns: testNamespace,
					},
					Experiments: []v1alpha1.ExperimentList{
						{
							Name: "experiment1",
							Spec: v1alpha1.ExperimentAttributes{
								Rank: 0,
								Components: v1alpha1.ExperimentComponents{
									ConfigMaps: []v1alpha1.ConfigMap{
										{
											Name: "configmap-1",
										},
									},
								},
							},
						},
					},
				},
			},
			isErrExpected: true,
		},
		{
			description: "Validation fails when only some of the specified ConfigMaps are in the Cluster.",
			k8sObjects: []runtime.Object{&v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "configmap-1",
					Namespace: testNamespace,
				},
			}},
			chaosEngine: v1alpha1.ChaosEngine{
				Spec: v1alpha1.ChaosEngineSpec{
					Appinfo: v1alpha1.ApplicationParams{
						Appns: testNamespace,
					},
					Experiments: []v1alpha1.ExperimentList{
						{
							Name: "experiment1",
							Spec: v1alpha1.ExperimentAttributes{
								Rank: 0,
								Components: v1alpha1.ExperimentComponents{
									ConfigMaps: []v1alpha1.ConfigMap{
										{
											Name: "configmap-1",
										},
									},
								},
							},
						},
						{
							Name: "experiment2",
							Spec: v1alpha1.ExperimentAttributes{
								Rank: 0,
								Components: v1alpha1.ExperimentComponents{
									ConfigMaps: []v1alpha1.ConfigMap{
										{
											Name: "configmap-2",
										},
									},
								},
							},
						},
					},
				},
			},
			isErrExpected: true,
		},
		{
			description: "Validation is successfull when all of the specified ConfigMaps are in the Cluster.",
			k8sObjects: []runtime.Object{
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "configmap-1",
						Namespace: testNamespace,
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "configmap-2",
						Namespace: testNamespace,
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "configmap-3",
						Namespace: testNamespace,
					},
				},
			},
			chaosEngine: v1alpha1.ChaosEngine{
				Spec: v1alpha1.ChaosEngineSpec{
					Appinfo: v1alpha1.ApplicationParams{
						Appns: testNamespace,
					},
					Experiments: []v1alpha1.ExperimentList{
						{
							Name: "experiment1",
							Spec: v1alpha1.ExperimentAttributes{
								Rank: 0,
								Components: v1alpha1.ExperimentComponents{
									ConfigMaps: []v1alpha1.ConfigMap{
										{
											Name: "configmap-1",
										},
										{
											Name: "configmap-2",
										},
									},
								},
							},
						},
						{
							Name: "experiment2",
							Spec: v1alpha1.ExperimentAttributes{
								Rank: 0,
								Components: v1alpha1.ExperimentComponents{
									ConfigMaps: []v1alpha1.ConfigMap{
										{
											Name: "configmap-3",
										},
									},
								},
							},
						},
					},
				},
			},
			isErrExpected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			webhook := webhook{
				kubeClient: fake.NewSimpleClientset(test.k8sObjects...),
			}
			err := webhook.ValidateChaosExperimentsConfigMaps(&test.chaosEngine)
			if test.isErrExpected && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil.", test.description)
			}
			if !test.isErrExpected && err != nil {
				fmt.Println(err)
				t.Fatalf("Test %q failed: expected error to be nil", test.description)
			}
		})
	}
}

func TestValidateChaosExperimentInApplicationNamespaces(t *testing.T) {
	var tests = []struct {
		description   string
		k8sObjects    []runtime.Object
		litmusObjects []runtime.Object
		chaosEngine   v1alpha1.ChaosEngine
		isErrExpected bool
	}{
		{
			description:   "Validation fails when none of the specified ChaosExperiment is in the Applcation Namespace.",
			litmusObjects: []runtime.Object{},
			chaosEngine: v1alpha1.ChaosEngine{
				Spec: v1alpha1.ChaosEngineSpec{
					Appinfo: v1alpha1.ApplicationParams{
						Appns: testNamespace,
					},
					Experiments: []v1alpha1.ExperimentList{
						{
							Name: "experiment1",
						},
					},
				},
			},
			isErrExpected: true,
		},
		{
			description: "Validation fails when only some of the specified ChaosExperiment are in the Applcation Namespace.",
			litmusObjects: []runtime.Object{
				&v1alpha1.ChaosExperiment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "experiment1",
						Namespace: testNamespace,
					},
				},
			},
			chaosEngine: v1alpha1.ChaosEngine{
				Spec: v1alpha1.ChaosEngineSpec{
					Appinfo: v1alpha1.ApplicationParams{
						Appns: testNamespace,
					},
					Experiments: []v1alpha1.ExperimentList{
						{
							Name: "experiment1",
						},
						{
							Name: "experiment2",
						},
					},
				},
			},
			isErrExpected: true,
		},
		{
			description: "Validation is successful when all of the specified ChaosExperiment are in the Applcation Namespace.",
			litmusObjects: []runtime.Object{
				&v1alpha1.ChaosExperiment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "experiment1",
						Namespace: testNamespace,
					},
				},
				&v1alpha1.ChaosExperiment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "experiment2",
						Namespace: testNamespace,
					},
				},
			},
			chaosEngine: v1alpha1.ChaosEngine{
				Spec: v1alpha1.ChaosEngineSpec{
					Appinfo: v1alpha1.ApplicationParams{
						Appns: testNamespace,
					},
					Experiments: []v1alpha1.ExperimentList{
						{
							Name: "experiment1",
						},
						{
							Name: "experiment2",
						},
					},
				},
			},
			isErrExpected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			webhook := webhook{
				kubeClient:   fake.NewSimpleClientset(test.k8sObjects...),
				litmusClient: fakelitmus.NewSimpleClientset(test.litmusObjects...),
			}
			err := webhook.ValidateChaosExperimentInApplicationNamespaces(&test.chaosEngine)
			if test.isErrExpected && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil.", test.description)
			}
			if !test.isErrExpected && err != nil {
				fmt.Println(err)
				t.Fatalf("Test %q failed: expected error to be nil", test.description)
			}
		})
	}
}

func TestValidateChaosExperimentsSecrets(t *testing.T) {
	var tests = []struct {
		description   string
		k8sObjects    []runtime.Object
		chaosEngine   v1alpha1.ChaosEngine
		isErrExpected bool
	}{
		{
			description: "Validation fails when none of the specified Secrets is in the Cluster.",
			k8sObjects:  []runtime.Object{},
			chaosEngine: v1alpha1.ChaosEngine{
				Spec: v1alpha1.ChaosEngineSpec{
					Appinfo: v1alpha1.ApplicationParams{
						Appns: testNamespace,
					},
					Experiments: []v1alpha1.ExperimentList{
						{
							Name: "experiment1",
							Spec: v1alpha1.ExperimentAttributes{
								Rank: 0,
								Components: v1alpha1.ExperimentComponents{
									Secrets: []v1alpha1.Secret{
										{
											Name: "secret-1",
										},
									},
								},
							},
						},
					},
				},
			},
			isErrExpected: true,
		},
		{
			description: "Validation fails when only some of the specified Secrets are in the Cluster.",
			k8sObjects: []runtime.Object{
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret-1",
						Namespace: testNamespace,
					},
				},
			},
			chaosEngine: v1alpha1.ChaosEngine{
				Spec: v1alpha1.ChaosEngineSpec{
					Appinfo: v1alpha1.ApplicationParams{
						Appns: testNamespace,
					},
					Experiments: []v1alpha1.ExperimentList{
						{
							Name: "experiment1",
							Spec: v1alpha1.ExperimentAttributes{
								Rank: 0,
								Components: v1alpha1.ExperimentComponents{
									Secrets: []v1alpha1.Secret{
										{
											Name: "secret-1",
										},
									},
								},
							},
						},
						{
							Name: "experiment2",
							Spec: v1alpha1.ExperimentAttributes{
								Rank: 0,
								Components: v1alpha1.ExperimentComponents{
									Secrets: []v1alpha1.Secret{
										{
											Name: "configmap-2",
										},
									},
								},
							},
						},
					},
				},
			},
			isErrExpected: true,
		},
		{
			description: "Validation is successful when all of the specified Secrets are in the Cluster.",
			k8sObjects: []runtime.Object{
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret-1",
						Namespace: testNamespace,
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret-2",
						Namespace: testNamespace,
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret-3",
						Namespace: testNamespace,
					},
				},
			},
			chaosEngine: v1alpha1.ChaosEngine{
				Spec: v1alpha1.ChaosEngineSpec{
					Appinfo: v1alpha1.ApplicationParams{
						Appns: testNamespace,
					},
					Experiments: []v1alpha1.ExperimentList{
						{
							Name: "experiment1",
							Spec: v1alpha1.ExperimentAttributes{
								Rank: 0,
								Components: v1alpha1.ExperimentComponents{
									Secrets: []v1alpha1.Secret{
										{
											Name: "secret-1",
										},
										{
											Name: "secret-2",
										},
									},
								},
							},
						},
						{
							Name: "experiment2",
							Spec: v1alpha1.ExperimentAttributes{
								Rank: 0,
								Components: v1alpha1.ExperimentComponents{
									Secrets: []v1alpha1.Secret{
										{
											Name: "secret-3",
										},
									},
								},
							},
						},
					},
				},
			},
			isErrExpected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			webhook := webhook{
				kubeClient: fake.NewSimpleClientset(test.k8sObjects...),
			}
			err := webhook.ValidateChaosExperimentsSecrets(&test.chaosEngine)
			if test.isErrExpected && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil.", test.description)
			}
			if !test.isErrExpected && err != nil {
				fmt.Println(err)
				t.Fatalf("Test %q failed: expected error to be nil", test.description)
			}
		})
	}
}

func TestValidateApplicationNamespace(t *testing.T) {
	var tests = []struct {
		description   string
		k8sObjects    []runtime.Object
		chaosEngine   v1alpha1.ChaosEngine
		isErrExpected bool
	}{
		{
			description: "Validation fails when none of the specified Namespace is not in the Cluster.",
			k8sObjects:  []runtime.Object{},
			chaosEngine: v1alpha1.ChaosEngine{
				Spec: v1alpha1.ChaosEngineSpec{
					Appinfo: v1alpha1.ApplicationParams{
						Appns: testNamespace,
					},
				},
			},
			isErrExpected: true,
		},
		{
			description: "Validation is successfull when all of the specified Namespace is in the Cluster.",
			k8sObjects: []runtime.Object{
				&v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: testNamespace,
					},
				},
			},
			chaosEngine: v1alpha1.ChaosEngine{
				Spec: v1alpha1.ChaosEngineSpec{
					Appinfo: v1alpha1.ApplicationParams{
						Appns: testNamespace,
					},
				},
			},
			isErrExpected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			webhook := webhook{
				kubeClient: fake.NewSimpleClientset(test.k8sObjects...),
			}
			err := webhook.ValidateApplicationNamespace(&test.chaosEngine)
			if test.isErrExpected && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil.", test.description)
			}
			if !test.isErrExpected && err != nil {
				fmt.Println(err)
				t.Fatalf("Test %q failed: expected error to be nil", test.description)
			}
		})
	}
}
