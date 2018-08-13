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
	"github.com/wata727/tflint/config"
)

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

// NewAwsClient returns new AwsClient with configured session
func NewAwsClient(cfg *config.Config) *AwsClient {
	s := newAwsSession(cfg)

	return &AwsClient{
		IAM:         iam.New(s),
		EC2:         ec2.New(s),
		RDS:         rds.New(s),
		ElastiCache: elasticache.New(s),
		ELB:         elb.New(s),
		ELBV2:       elbv2.New(s),
		ECS:         ecs.New(s),
	}
}

// newAwsSession returns a session necessary for initialization of the AWS SDK
func newAwsSession(cfg *config.Config) *session.Session {
	s := session.New()

	if cfg.HasAwsRegion() {
		log.Printf("[INFO] Set AWS region: %s", cfg.AwsCredentials["region"])
		s = session.New(&aws.Config{
			Region: aws.String(cfg.AwsCredentials["region"]),
		})
	}
	if cfg.HasAwsSharedCredentials() {
		log.Printf("[INFO] Set AWS shared credentials")
		path, err := homedir.Expand("~/.aws/credentials")
		if err != nil {
			// Maybe this is bug
			panic(err)
		}
		s = session.New(&aws.Config{
			Credentials: credentials.NewSharedCredentials(path, cfg.AwsCredentials["profile"]),
			Region:      aws.String(cfg.AwsCredentials["region"]),
		})
	}
	if cfg.HasAwsStaticCredentials() {
		log.Printf("[INFO] Set AWS static credentials")
		s = session.New(&aws.Config{
			Credentials: credentials.NewStaticCredentials(cfg.AwsCredentials["access_key"], cfg.AwsCredentials["secret_key"], ""),
			Region:      aws.String(cfg.AwsCredentials["region"]),
		})
	}

	return s
}
