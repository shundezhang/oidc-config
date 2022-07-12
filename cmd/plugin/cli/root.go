package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	KubernetesConfigFlags *genericclioptions.ConfigFlags
	awsProfile            = "aws-profile"
	kubeConfigPath        = "kubeconfig"
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
	rootCmd.PersistentFlags().String(kubeConfigPath, "", "Path to kubeconfig")
	rootCmd.PersistentFlags().String(awsProfile, "default", "AWS profile name in .aws/config")
}

func initConfig() {
	viper.AutomaticEnv()
}
