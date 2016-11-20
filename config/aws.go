package config

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/wata727/tflint/logger"
)

type AwsClient struct {
	Iam iamiface.IAMAPI
}

func (c *Config) NewAwsClient() *AwsClient {
	client := &AwsClient{}
	s := c.NewAwsSession()

	client.Iam = iam.New(s)

	return client
}

func (c *Config) NewAwsSession() *session.Session {
	var l = logger.Init(c.Debug)

	s := session.New()
	if c.HasAwsRegion() {
		l.Info("set AWS region")
		s = session.New(&aws.Config{
			Region: aws.String(c.AwsCredentials["region"]),
		})
	}
	if c.HasAwsCredentials() {
		l.Info("set AWS credentials")
		s = session.New(&aws.Config{
			Credentials: credentials.NewStaticCredentials(c.AwsCredentials["access_key"], c.AwsCredentials["secret_key"], ""),
			Region:      aws.String(c.AwsCredentials["region"]),
		})
	}

	return s
}
