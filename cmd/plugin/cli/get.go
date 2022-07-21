package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"net/url"

	"github.com/shundezhang/oidc-config/pkg/aws"
	"github.com/shundezhang/oidc-config/pkg/k8s"
	"github.com/shundezhang/oidc-config/pkg/logger"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	outputFormat           = "output"
	uploadFlag             = "upload-to-s3"
	createOidcProviderFlag = "create-oidc-provider"
)

type Oidc struct {
	configUrl     string
	configContent string
	jwksUrl       string
	jwksContent   string
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get oidc config content, upload to s3 and create oidc provider",
	Long:  `get oidc config content, upload to s3 and create oidc provider`,
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.NewLogger()
		// log.Info("")
		configPath, err := cmd.Flags().GetString(kubeConfigPath)
		if err != nil {
			log.Error(err)
			return
		}
		output, err := cmd.Flags().GetString(outputFormat)
		if err != nil {
			log.Error(err)
			return
		}
		upload, err := cmd.Flags().GetBool(uploadFlag)
		if err != nil {
			log.Error(err)
			return
		}
		create, err := cmd.Flags().GetBool(createOidcProviderFlag)
		if err != nil {
			log.Error(err)
			return
		}
		profile, err := cmd.Flags().GetString(awsProfile)
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
		jwks, err := k8s.GetURL(c.Host+"/openid/v1/jwks", c.BearerToken, c.CAData)
		if err != nil {
			log.Error(err)
			return
		}
		if output == "" {
			fmt.Println(objmap["issuer"])
			fmt.Println(string(config))
			fmt.Println(objmap["jwks_uri"])
			fmt.Println(string(jwks))
		} else {
			outmap := make(map[string]interface{})
			outmap["configURL"] = fmt.Sprintf("%v/.well-known/openid-configuration", objmap["issuer"])
			outmap["configContent"] = string(config)
			outmap["jwksURL"] = objmap["jwks_uri"]
			outmap["jwksContent"] = string(jwks)
			if output == "json" {
				b, err := json.MarshalIndent(outmap, "", "  ")
				if err != nil {
					log.Error(err)
					return
				}
				fmt.Println(string(b))
			} else if output == "yaml" {
				out, err := yaml.Marshal(outmap)
				if err != nil {
					log.Error(err)
				}

				fmt.Println(string(out))
			} else {
				log.Info("output format %s not supported.", output)
			}
		}
		if upload {
			u, err := url.Parse(fmt.Sprintf("%v", objmap["issuer"]))
			if err != nil {
				log.Error(err)
				return
			}
			if !strings.HasSuffix(u.Hostname(), "s3.amazonaws.com") {
				fmt.Println("URL is not an S3 URL")
				return
			}
			bucket := strings.Split(u.Hostname(), ".")[0]
			err1 := aws.UploadToS3(profile, bucket, u.Path+"/.well-known/openid-configuration", string(config), u.Path+"/openid/v1/jwks", string(jwks))
			if err1 != nil {
				log.Error(err1)
				return
			}
		}
		if create {
			err1 := aws.CreateOIDCProvider(profile, fmt.Sprintf("%v", objmap["issuer"]))
			if err1 != nil {
				log.Error(err1)
				return
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().StringP(outputFormat, "o", "", "output format: default, yaml or json")
	getCmd.Flags().Bool(uploadFlag, false, "Upload config and jwks to s3 bucket")
	getCmd.Flags().Bool(createOidcProviderFlag, false, "Create OIDC provider in IAM")
}
