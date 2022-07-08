package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/shundezhang/oidc-config/pkg/logger"
)

func UploadToS3(profile, bucket, configPath, configContent, jwksPath, jwksContent string) error {
	log := logger.NewLogger()
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile:           profile,
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create S3 service client
	svc := s3.New(sess)
	result, err := svc.ListBuckets(nil)
	if err != nil {
		fmt.Printf("Unable to list buckets, %v", err)
		return err
	}

	if !bucketExists(bucket, result.Buckets) {
		log.Info("bucket %s not found, creating it...", bucket)
		acl := "public-read"
		newBucket := &s3.CreateBucketInput{
			Bucket: aws.String(bucket),
			ACL:    &acl,
		}
		result, err := svc.CreateBucket(newBucket)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case s3.ErrCodeBucketAlreadyExists:
					fmt.Println(s3.ErrCodeBucketAlreadyExists, aerr.Error())
				case s3.ErrCodeBucketAlreadyOwnedByYou:
					fmt.Println(s3.ErrCodeBucketAlreadyOwnedByYou, aerr.Error())
				default:
					fmt.Println(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				fmt.Println(err.Error())
			}
			return err
		}
		fmt.Println(result)
	} else {
		log.Info("bucket %s exists.", bucket)
	}
	log.Info("Put config %s to bucket %s...", configPath, bucket)
	putPublicObject(svc, bucket, configPath, configContent)
	log.Info("Put jwks %s to bucket %s...", jwksPath, bucket)
	putPublicObject(svc, bucket, jwksPath, jwksContent)
	return nil
}

func putPublicObject(svc *s3.S3, bucket string, key string, content string) {
	input := &s3.PutObjectInput{
		ACL:    aws.String("public-read"),
		Body:   aws.ReadSeekCloser(strings.NewReader(content)),
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	result, err := svc.PutObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(result)
}
func bucketExists(bucket string, buckets []*s3.Bucket) bool {
	for _, b := range buckets {
		if aws.StringValue(b.Name) == bucket {
			return true
		}
	}
	return false
}
