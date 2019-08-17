package client

import (
	"errors"
	"log"
	"strings"

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
	awsbase "github.com/hashicorp/aws-sdk-go-base"
	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/terraform/configs"
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

type awsProvider struct {
	AccessKey     string `hcl:"access_key,optional"`
	SecretKey     string `hcl:"secret_key,optional"`
	Profile       string `hcl:"profile,optional"`
	CredsFilename string `hcl:"shared_credentials_file,optional"`
	Region        string `hcl:"region,optional"`

	Remain hcl.Body `hcl:",remain"`
}

// NewAwsClient returns new AwsClient with configured session
func NewAwsClient(provider *configs.Provider, creds AwsCredentials) (*AwsClient, error) {
	log.Print("[INFO] Initialize AWS Client")

	config, err := getBaseConfig(provider, creds)
	if err != nil {
		return nil, err
	}

	s, err := awsbase.GetSession(config)
	if err != nil {
		return nil, formatBaseConfigError(err)
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

func getBaseConfig(provider *configs.Provider, creds AwsCredentials) (*awsbase.Config, error) {
	base := &awsbase.Config{}

	if provider != nil {
		pc, err := decodeProviderConfig(provider)
		if err != nil {
			return base, err
		}

		base = &awsbase.Config{
			AccessKey: pc.AccessKey,
			Profile:   pc.Profile,
			Region:    pc.Region,
			SecretKey: pc.SecretKey,
		}

		path, err := homedir.Expand(pc.CredsFilename)
		if err != nil {
			return base, err
		}
		base.CredsFilename = path
	}

	if creds.AccessKey != "" {
		base.AccessKey = creds.AccessKey
	}
	if creds.CredsFile != "" {
		path, err := homedir.Expand(creds.CredsFile)
		if err != nil {
			return base, err
		}
		base.CredsFilename = path
	}
	if creds.Profile != "" {
		base.Profile = creds.Profile
	}
	if creds.Region != "" {
		base.Region = creds.Region
	}
	if creds.SecretKey != "" {
		base.SecretKey = creds.SecretKey
	}

	return base, nil
}

func decodeProviderConfig(provider *configs.Provider) (awsProvider, error) {
	var config awsProvider
	diags := gohcl.DecodeBody(provider.Config, nil, &config)
	if diags.HasErrors() {
		return config, diags
	}
	return config, nil
}

// @see https://github.com/hashicorp/aws-sdk-go-base/blob/v0.3.0/session.go#L87
func formatBaseConfigError(err error) error {
	if strings.Contains(err.Error(), "No valid credential sources found for AWS Provider") {
		return errors.New("No valid credential sources found")
	}
	return err
}
