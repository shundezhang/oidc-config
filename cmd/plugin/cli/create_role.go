package cli

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/shundezhang/oidc-config/pkg/aws"
	"github.com/shundezhang/oidc-config/pkg/k8s"
	"github.com/shundezhang/oidc-config/pkg/logger"
	"github.com/spf13/cobra"
)

const (
	saName          = "sa-name"
	saNameSpace     = "sa-namespace"
	roleName        = "role-name"
	policyName      = "policy-name"
	createSAFlag    = "create-sa"
	allowAllSAsFlag = "allow-all-sas"
)

var createRoleCmd = &cobra.Command{
	Use:   "create-role",
	Short: "create a role for sa",
	Long:  `create a role for sa`,
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.NewLogger()
		// configPath, err := cmd.Flags().GetString(kubeConfigPath)
		// if err != nil {
		// 	return
		// }
		role, err := cmd.Flags().GetString(roleName)
		if err != nil {
			log.Error(err)
			return
		}
		configPath, err := cmd.Flags().GetString(kubeConfigPath)
		if err != nil {
			log.Error(err)
			return
		}
		profile, err := cmd.Flags().GetString(awsProfile)
		if err != nil {
			log.Error(err)
			return
		}
		ns, err := cmd.Flags().GetString(saNameSpace)
		if err != nil {
			log.Error(err)
			return
		}
		sa, err := cmd.Flags().GetString(saName)
		if err != nil {
			log.Error(err)
			return
		}
		policy, err := cmd.Flags().GetString(policyName)
		if err != nil {
			log.Error(err)
			return
		}
		createSA, err := cmd.Flags().GetBool(createSAFlag)
		if err != nil {
			log.Error(err)
			return
		}
		allowAllSAs, err := cmd.Flags().GetBool(allowAllSAsFlag)
		if err != nil {
			log.Error(err)
			return
		}
		c, err := k8s.GetKubernetesConfig(configPath)
		if err != nil {
			log.Error(err)
			return
		}
		config, err := k8s.GetURL(c.Host+"/.well-known/openid-configuration", c.BearerToken, c.CAData)
		if err != nil {
			log.Error(err)
			return
		}
		var objmap map[string]interface{}
		if err := json.Unmarshal([]byte(config), &objmap); err != nil {
			log.Error(err)
		}
		u, err := url.Parse(fmt.Sprintf("%v", objmap["issuer"]))
		if err != nil {
			log.Error(err)
			return
		}
		arn, err := aws.GetOIDCProviderARN(profile, u.Hostname()+u.Path)
		if err != nil {
			log.Error(err)
			return
		}
		allowedSA := sa
		if allowAllSAs {
			allowedSA = "*"
		}
		policyArn, err := aws.GetPolicyARN(profile, policy)
		if err != nil {
			log.Error(err)
			return
		}
		roleArn, err := aws.CreateRole(profile, role, policyArn, arn, ns, allowedSA)
		if err != nil {
			log.Error(err)
			return
		}
		if createSA {
			err := k8s.CreateSA(configPath, sa, ns, roleArn)
			if err != nil {
				log.Error(err)
				return
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(createRoleCmd)
	createRoleCmd.Flags().StringP(roleName, "r", "", "role name")
	createRoleCmd.Flags().StringP(policyName, "p", "", "policy name")
	createRoleCmd.Flags().String(saName, "my-sa", "sa name")
	createRoleCmd.Flags().String(saNameSpace, "default", "sa namespace")
	createRoleCmd.Flags().Bool(createSAFlag, false, "Create SA for this role")
	createRoleCmd.Flags().Bool(allowAllSAsFlag, true, "Allow all SAs in the namespace to use this role, otherwise only the created SA can use.")
	createRoleCmd.MarkFlagRequired(roleName)
	createRoleCmd.MarkFlagRequired(policyName)
}
