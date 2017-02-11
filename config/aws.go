package config

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/wata727/tflint/logger"
)

type AwsClient struct {
	Iam         iamiface.IAMAPI
	Ec2         ec2iface.EC2API
	Rds         rdsiface.RDSAPI
	Elasticache elasticacheiface.ElastiCacheAPI
	Elb         elbiface.ELBAPI
	Elbv2       elbv2iface.ELBV2API
}

func (c *Config) NewAwsClient() *AwsClient {
	client := &AwsClient{}
	s := c.NewAwsSession()

	client.Iam = iam.New(s)
	client.Ec2 = ec2.New(s)
	client.Rds = rds.New(s)
	client.Elasticache = elasticache.New(s)
	client.Elb = elb.New(s)
	client.Elbv2 = elbv2.New(s)

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
