package cli

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	kubeConfigPath = "kubeconfig"
	saName         = "saname"
	saNameSpace    = "sanamespace"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "get oidc content",
	Long:  `get oidc content in configure`,
	Run: func(cmd *cobra.Command, args []string) {
		configPath, err := cmd.Flags().GetString(kubeConfigPath)
		if err != nil {
			return
		}
		serviceAccountName, err := cmd.Flags().GetString(saName)
		if err != nil {
			log.Fatal("unable to get service account name")
			return
		}
		serviceAccountNameSpace, err := cmd.Flags().GetString(saNameSpace)
		if err != nil {
			log.Fatal("unable to get service account namespace")
			return
		}
		k, err := getKubernetesClient(configPath)
		if err != nil {
			log.Fatal(err)
		}
		sa, err := k.CoreV1().ServiceAccounts(serviceAccountNameSpace).Get(serviceAccountName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			log.Fatal(err)
			return
		}
		if err != nil {
			fmt.Printf("unable to get service account %s/%s", serviceAccountName, serviceAccountNameSpace)
			return
		}
		for _, ref := range sa.Secrets {
			secret, err := k.CoreV1().Secrets(serviceAccountNameSpace).Get(ref.Name, metav1.GetOptions{})
			if errors.IsNotFound(err) {
				fmt.Printf("secret %s/%s not found", serviceAccountName, ref.Name)
				continue
			}
			if err != nil {
				fmt.Printf("unable to get secret %s/%s", serviceAccountName, ref.Name)
				return
			}
			if secret.Type != api.SecretTypeServiceAccountToken {
				fmt.Printf("secret %s/%s is not a service account token", serviceAccountName, ref.Name)
				continue
			}

			tokenData := secret.Data[api.ServiceAccountTokenKey]
			fmt.Printf("token: %s", string(tokenData[:]))
		}

	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
	viewCmd.Flags().String(kubeConfigPath, "", "Path to kubeconfig")
	viewCmd.Flags().String(saName, "default", "Service Account Name")
	viewCmd.Flags().String(saNameSpace, "default", "Service Account NameSpace")
}
