package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	KubernetesConfigFlags *genericclioptions.ConfigFlags
)

var rootCmd = &cobra.Command{
	Use:   "oidc-config",
	Short: "config oidc stuff",
	Long:  "config oidc stuff",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.AutomaticEnv()
}

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

func getKubernetesConfig(kubePath string) (*rest.Config, error) {
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

func getKubernetesClient(kubePath string) (kubernetes.Interface, error) {
	config, err := getKubernetesConfig(kubePath)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)

	if err != nil {
		return nil, err
	}
	return client, nil
}
