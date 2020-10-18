package webhook

import (
	"fmt"
	"testing"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
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
			description: "Validation is successfull when all of the specified Secrets are in the Cluster.",
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
