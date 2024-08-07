package clients

import (
	"github.com/litmuschaos/admission-controller/pkg/log"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type ClientSets struct {
	KubeConfig    *rest.Config
	KubeClient    kubernetes.Interface
	DynamicClient dynamic.Interface
}

func GenerateClientSetFromKubeConfig(kubeConfig string) (ClientSets, error) {

	config, err := getKubeConfig(kubeConfig)
	if err != nil {
		return ClientSets{}, err
	}
	kubeClient, err := generateK8sClientSet(config)
	if err != nil {
		return ClientSets{}, err
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return ClientSets{}, err
	}
	return ClientSets{
		KubeConfig:    config,
		KubeClient:    kubeClient,
		DynamicClient: dynamicClient,
	}, nil
}

// getKubeConfig setup the config for access cluster resource
func getKubeConfig(kubeConfig string) (*rest.Config, error) {
	// It uses in-cluster config, if kubeconfig path is not specified
	config, err := buildConfigFromFlags("", kubeConfig)
	return config, err
}

// generateK8sClientSet will generation k8s client
func generateK8sClientSet(config *rest.Config) (*kubernetes.Clientset, error) {
	k8sClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to generate kubernetes clientSet, err: %v: ", err)
	}
	return k8sClientSet, nil
}

// buildConfigFromFlags is a helper function that builds configs from a master
// url or a kubeconfig filepath, if nothing is provided it falls back to inClusterConfig
func buildConfigFromFlags(masterUrl, kubeconfigPath string) (*rest.Config, error) {
	if kubeconfigPath == "" && masterUrl == "" {
		kubeconfig, err := rest.InClusterConfig()
		if err == nil {
			return kubeconfig, nil
		}
		log.Logger.Warnf("Neither --kubeconfig nor --master was specified.  Using the inClusterConfig. Error creating inClusterConfig: %v", err)
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: masterUrl}}).ClientConfig()
}
