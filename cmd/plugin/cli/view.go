package cli

import (
	"fmt"
	"io/ioutil"
	"log"

	"net/http"
	"time"

	"github.com/spf13/cobra"
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
		c, err := getKubernetesConfig(configPath)
		if err != nil {
			log.Fatal(err)
			return
		}
		config, err := get(c.Host+"/.well-known/openid-configuration", c.BearerToken)
		if err != nil {
			log.Fatal(err)
			return
		}
		fmt.Println(config)
	},
}

func get(url, token string) ([]byte, error) {
	client := &http.Client{
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
}
