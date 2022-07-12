package aws

import (
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/prometheus/common/log"
	"github.com/shundezhang/oidc-config/pkg/logger"
)

var trustedPolicyTemplate = `{
	"Version": "2012-10-17",
	"Statement": [
	  {
		"Effect": "Allow",
		"Principal": {
			"Federated": "%s"
		},
		"Action": "sts:AssumeRoleWithWebIdentity",
		"Condition": {
			"StringLike": {
				"%s:sub": "system:serviceaccount:%s:%s"
			}
		}
	  }
	]
  }`

func CreateOIDCProvider(profile, providerUrl string) error {
	log := logger.NewLogger()
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile:           profile,
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := iam.New(sess)
	input := &iam.CreateOpenIDConnectProviderInput{
		ClientIDList: []*string{
			aws.String("sts.amazonaws.com"),
		},
		ThumbprintList: []*string{
			aws.String(getThumbPrint(providerUrl)),
		},
		Url: aws.String(providerUrl),
	}

	result, err := svc.CreateOpenIDConnectProvider(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeInvalidInputException:
				fmt.Println(iam.ErrCodeInvalidInputException, aerr.Error())
			case iam.ErrCodeEntityAlreadyExistsException:
				fmt.Println(iam.ErrCodeEntityAlreadyExistsException, aerr.Error())
			case iam.ErrCodeLimitExceededException:
				fmt.Println(iam.ErrCodeLimitExceededException, aerr.Error())
			case iam.ErrCodeConcurrentModificationException:
				fmt.Println(iam.ErrCodeConcurrentModificationException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Error(err)
		}
		return err
	}

	fmt.Println(result)
	return nil
}

func getThumbPrint(httpsUrl string) string {
	u, err := url.Parse(httpsUrl)
	if err != nil {
		log.Error(err)
		return ""
	}
	add := u.Hostname()
	if u.Port() != "" {
		add = add + ":" + u.Port()
	} else if u.Port() == "" && u.Scheme == "https" {
		add = add + ":443"
	}
	fmt.Printf("getting thumbprint from %s\n", add)
	conn, err := tls.Dial("tcp", add, &tls.Config{})
	if err != nil {
		panic("failed to connect: " + err.Error())
	}

	// Get the ConnectionState struct as that's the one which gives us x509.Certificate struct
	cert := conn.ConnectionState().PeerCertificates[0]
	// fmt.Printf("%s\n", cert.Issuer)
	fingerprint := sha1.Sum(cert.Raw)

	var buf bytes.Buffer
	for _, f := range fingerprint {
		// if i > 0 {
		// 	fmt.Fprintf(&buf, ":")
		// }
		fmt.Fprintf(&buf, "%02X", f)
	}
	fmt.Printf("%x", fingerprint)
	fmt.Printf("Fingerprint for %s: %s", httpsUrl, buf.String())

	defer conn.Close()
	return buf.String()
}

func CreateRole(profile, roleName, policyArn, oidcProviderArn, saNamespace, sa string) (string, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile:           profile,
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := iam.New(sess)
	input := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(fmt.Sprintf(trustedPolicyTemplate, oidcProviderArn, strings.Split(oidcProviderArn, "oidc-provider/")[1], saNamespace, sa)),
		RoleName:                 aws.String(roleName),
	}

	result, err := svc.CreateRole(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeLimitExceededException:
				fmt.Println(iam.ErrCodeLimitExceededException, aerr.Error())
			case iam.ErrCodeInvalidInputException:
				fmt.Println(iam.ErrCodeInvalidInputException, aerr.Error())
			case iam.ErrCodeEntityAlreadyExistsException:
				fmt.Println(iam.ErrCodeEntityAlreadyExistsException, aerr.Error())
			case iam.ErrCodeMalformedPolicyDocumentException:
				fmt.Println(iam.ErrCodeMalformedPolicyDocumentException, aerr.Error())
			case iam.ErrCodeConcurrentModificationException:
				fmt.Println(iam.ErrCodeConcurrentModificationException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return "", err
	}
	fmt.Println(result)

	inputP := &iam.AttachRolePolicyInput{
		PolicyArn: aws.String(policyArn),
		RoleName:  aws.String(roleName),
	}

	resultP, errP := svc.AttachRolePolicy(inputP)
	if errP != nil {
		if aerr, ok := errP.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				fmt.Println(iam.ErrCodeNoSuchEntityException, aerr.Error())
			case iam.ErrCodeLimitExceededException:
				fmt.Println(iam.ErrCodeLimitExceededException, aerr.Error())
			case iam.ErrCodeInvalidInputException:
				fmt.Println(iam.ErrCodeInvalidInputException, aerr.Error())
			case iam.ErrCodeUnmodifiableEntityException:
				fmt.Println(iam.ErrCodeUnmodifiableEntityException, aerr.Error())
			case iam.ErrCodePolicyNotAttachableException:
				fmt.Println(iam.ErrCodePolicyNotAttachableException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return "", errP
	}

	fmt.Println(resultP)

	return *result.Role.Arn, nil
}

func GetPolicyARN(profile, policy string) (string, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile:           profile,
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := iam.New(sess)
	policies, err := svc.ListPolicies(&iam.ListPoliciesInput{})
	if err != nil {
		return "", err
	}
	for _, p := range policies.Policies {
		if *p.PolicyName == policy {
			return *p.Arn, nil
		}
	}
	return "", errors.New("Policy " + policy + " not found.")
}

func GetOIDCProviderARN(profile, issuer string) (string, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile:           profile,
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := iam.New(sess)
	providers, err := svc.ListOpenIDConnectProviders(&iam.ListOpenIDConnectProvidersInput{})
	if err != nil {
		return "", err
	}
	for _, provider := range providers.OpenIDConnectProviderList {
		if strings.HasSuffix(*provider.Arn, issuer) {
			return *provider.Arn, nil
		}
	}
	return "", errors.New("OIDC Provider " + issuer + " not found.")
}
