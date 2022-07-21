package k8s

import (
	"context"
	"fmt"

	"github.com/shundezhang/oidc-config/pkg/logger"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateSA(kubePath, saName, saNameSpace, roleArn string) error {
	log := logger.NewLogger()
	k, err := GetKubernetesClient(kubePath)
	if err != nil {
		return err
	}
	saClient := k.CoreV1().ServiceAccounts(saNameSpace)
	sa := &apiv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: saName,
			Annotations: map[string]string{
				"eks.amazonaws.com/role-arn":               roleArn,
				"eks.amazonaws.com/audience":               "sts.amazonaws.com",
				"eks.amazonaws.com/sts-regional-endpoints": "true",
				"eks.amazonaws.com/token-expiration":       "86400",
			},
		},
	}
	result, err := saClient.Create(context.TODO(), sa, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	fmt.Println(result)
	log.Info("Created service account %s/%s", saName, saNameSpace)
	return nil
}
