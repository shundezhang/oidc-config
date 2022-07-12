package k8s

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func GetURL(url string, token string, ca []byte) ([]byte, error) {
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
