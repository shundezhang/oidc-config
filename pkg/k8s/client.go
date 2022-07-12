package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getKubernetesConfigInCluster() (*rest.Config, error) {
	var config *rest.Config
	config, err := rest.InClusterConfig()
	if err != nil {
		return getKubernetesLocalConfig()
	}
	return config, nil
}

func getKubernetesLocalConfig() (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	clientCfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	return clientCfg.ClientConfig()
}

func GetKubernetesConfig(kubePath string) (*rest.Config, error) {
	var (
		config *rest.Config
		err    error
	)

	// if kubeconfig path is not provided, try to auto detect
	if kubePath == "" {
		config, err = getKubernetesConfigInCluster()
		if err != nil {
			return nil, err
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubePath)
		if err != nil {
			return nil, err
		}
	}

	return config, err
}

func GetKubernetesClient(kubePath string) (kubernetes.Interface, error) {
	config, err := GetKubernetesConfig(kubePath)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)

	if err != nil {
		return nil, err
	}
	return client, nil
}
