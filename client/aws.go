package client

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
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
	homedir "github.com/mitchellh/go-homedir"
)

//go:generate mockgen -source ../vendor/github.com/aws/aws-sdk-go/service/ec2/ec2iface/interface.go -destination aws_ec2_mock.go -package client
//go:generate mockgen -source ../vendor/github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface/interface.go -destination aws_elasticache_mock.go -package client
//go:generate mockgen -source ../vendor/github.com/aws/aws-sdk-go/service/elb/elbiface/interface.go -destination aws_elb_mock.go -package client
//go:generate mockgen -source ../vendor/github.com/aws/aws-sdk-go/service/elbv2/elbv2iface/interface.go -destination aws_elbv2_mock.go -package client
//go:generate mockgen -source ../vendor/github.com/aws/aws-sdk-go/service/iam/iamiface/interface.go -destination aws_iam_mock.go -package client
//go:generate mockgen -source ../vendor/github.com/aws/aws-sdk-go/service/rds/rdsiface/interface.go -destination aws_rds_mock.go -package client
//go:generate mockgen -source ../vendor/github.com/aws/aws-sdk-go/service/ecs/ecsiface/interface.go -destination aws_ecs_mock.go -package client

// AwsClient is a wrapper of the AWS SDK client
// It has interfaces for each services to make testing easier
type AwsClient struct {
	IAM         iamiface.IAMAPI
	EC2         ec2iface.EC2API
	RDS         rdsiface.RDSAPI
	ElastiCache elasticacheiface.ElastiCacheAPI
	ELB         elbiface.ELBAPI
	ELBV2       elbv2iface.ELBV2API
	ECS         ecsiface.ECSAPI
}

// AwsCredentials is credentials for AWS used in deep check mode
type AwsCredentials struct {
	AccessKey string
	SecretKey string
	Profile   string
	CredsFile string
	Region    string
}

// NewAwsClient returns new AwsClient with configured session
func NewAwsClient(creds AwsCredentials) (*AwsClient, error) {
	log.Print("[INFO] Initialize AWS Client")

	s, err := newAwsSession(creds)
	if err != nil {
		return nil, err
	}

	return &AwsClient{
		IAM:         iam.New(s),
		EC2:         ec2.New(s),
		RDS:         rds.New(s),
		ElastiCache: elasticache.New(s),
		ELB:         elb.New(s),
		ELBV2:       elbv2.New(s),
		ECS:         ecs.New(s),
	}, nil
}

// newAwsSession returns a session necessary for initialization of the AWS SDK
func newAwsSession(creds AwsCredentials) (*session.Session, error) {
	s := session.New()

	if creds.Region != "" {
		log.Printf("[INFO] Set AWS region: %s", creds.Region)
		s = session.New(&aws.Config{
			Region: aws.String(creds.Region),
		})
	}
	if creds.Profile != "" && creds.Region != "" {
		var credsFile string
		var err error
		if creds.CredsFile != "" {
			credsFile, err = homedir.Expand(creds.CredsFile)
			if err != nil {
				return nil, err
			}
		} else {
			credsFile, err = homedir.Expand("~/.aws/credentials")
			if err != nil {
				// Maybe this is bug
				panic(err)
			}
		}
		log.Printf("[INFO] Set AWS shared credentials: %s", credsFile)
		s = session.New(&aws.Config{
			Credentials: credentials.NewSharedCredentials(credsFile, creds.Profile),
			Region:      aws.String(creds.Region),
		})
	}
	if creds.AccessKey != "" && creds.SecretKey != "" && creds.Region != "" {
		log.Printf("[INFO] Set AWS static credentials")
		s = session.New(&aws.Config{
			Credentials: credentials.NewStaticCredentials(creds.AccessKey, creds.SecretKey, ""),
			Region:      aws.String(creds.Region),
		})
	}

	return s, nil
}
