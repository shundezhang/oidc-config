package cli

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"net/http"
	"net/url"
	"time"

	"github.com/shundezhang/oidc-config/pkg/logger"
	"github.com/shundezhang/oidc-config/pkg/s3"
	"github.com/spf13/cobra"
)

const (
	kubeConfigPath = "kubeconfig"
	outputFormat   = "output"
	uploadFlag     = "upload-to-s3"
)

type Oidc struct {
	configUrl     string
	configContent string
	jwksUrl       string
	jwksContent   string
}

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "get oidc content",
	Long:  `get oidc content in configure`,
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.NewLogger()
		log.Info("")
		configPath, err := cmd.Flags().GetString(kubeConfigPath)
		if err != nil {
			return
		}
		output, err := cmd.Flags().GetString(outputFormat)
		if err != nil {
			return
		}
		upload, err := cmd.Flags().GetBool(uploadFlag)
		if err != nil {
			return
		}
		c, err := getKubernetesConfig(configPath)
		if err != nil {
			log.Error(err)
			return
		}
		config, err := get(c.Host+"/.well-known/openid-configuration", c.BearerToken, c.CAData)
		if err != nil {
			log.Error(err)
			return
		}
		var objmap map[string]interface{}
		if err := json.Unmarshal([]byte(config), &objmap); err != nil {
			log.Error(err)
		}
		jwks, err := get(c.Host+"/openid/v1/jwks", c.BearerToken, c.CAData)
		if err != nil {
			log.Error(err)
			return
		}
		if output == "" {
			fmt.Println(objmap["issuer"])
			fmt.Println(string(config))
			fmt.Println(objmap["jwks_uri"])
			fmt.Println(string(jwks))
		} else if output == "json" {
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
			s3.UploadToS3(bucket, u.Path+"/.well-known/openid-configuration", string(config), u.Path+"/openid/v1/jwks", jwks)
		}
	},
}

func get(url string, token string, ca []byte) ([]byte, error) {
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(ca)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Got error %s", err.Error())
	}
	req.Header.Set("Authorization", "Bearer "+token)
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Got error %s", err.Error())
	}
	data, _ := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	return data, nil
}

func init() {
	rootCmd.AddCommand(viewCmd)
	viewCmd.Flags().String(kubeConfigPath, "", "Path to kubeconfig")
	viewCmd.Flags().StringP(outputFormat, "o", "", "output format: default, yaml or json")
	viewCmd.Flags().Bool(uploadFlag, false, "Upload config and jwks to s3 bucket")
}
