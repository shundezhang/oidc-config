package cli

import (
	"strings"

	"github.com/shundezhang/oidc-config/pkg/k8s"
	"github.com/shundezhang/oidc-config/pkg/logger"
	"github.com/spf13/cobra"
)

const (
	cmYAML     = "cm-yaml"
	webhookURL = "webhook-url"
)

var (
	webhookFiles = []string{"auth.yaml", "deployment-base.yaml", "service.yaml", "mutatingwebhook.yaml"}
)
var deployWebhookCmd = &cobra.Command{
	Use:   "deploy-webhook",
	Short: "deploy eks-pod-identity-webhook",
	Long:  `deploy eks-pod-identity-webhook`,
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.NewLogger()
		cm, err := cmd.Flags().GetString(cmYAML)
		if err != nil {
			log.Error(err)
			return
		}
		wh, err := cmd.Flags().GetString(webhookURL)
		if err != nil {
			log.Error(err)
			return
		}
		configPath, err := cmd.Flags().GetString(kubeConfigPath)
		if err != nil {
			log.Error(err)
			return
		}
		log.Info("Getting YAML from %s", cm)
		content, err := k8s.GetYAML(cm)
		if err != nil {
			log.Error(err)
			return
		}
		err = k8s.Apply(configPath, content)
		if err != nil {
			log.Error(err)
			return
		}
		for i := range webhookFiles {
			log.Info("Getting YAML from %s", wh+webhookFiles[i])
			content, err := k8s.GetYAML(wh + webhookFiles[i])
			if err != nil {
				log.Error(err)
				return
			}
			if webhookFiles[i] == "deployment-base.yaml" {
				content = []byte(strings.ReplaceAll(string(content), "IMAGE", "amazon/amazon-eks-pod-identity-webhook:latest"))
			}
			err = k8s.Apply(configPath, content)
			if err != nil {
				log.Error(err)
				return
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(deployWebhookCmd)
	deployWebhookCmd.Flags().String(cmYAML, "https://github.com/cert-manager/cert-manager/releases/download/v1.8.2/cert-manager.yaml", "cert manager yaml URL")
	deployWebhookCmd.Flags().String(webhookURL, "https://github.com/aws/amazon-eks-pod-identity-webhook/raw/v0.4.0/deploy/", "eks pod identity webhook URL")
}
