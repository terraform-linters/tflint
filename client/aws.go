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
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform/configs/configschema"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/zclconf/go-cty/cty"
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
	AccessKey             string
	SecretKey             string
	Profile               string
	CredsFile             string
	AssumeRoleARN         string
	AssumeRoleExternalID  string
	AssumeRolePolicy      string
	AssumeRoleSessionName string
	Region                string
}

// AwsProviderBlockSchema is a schema of `aws` provider block
var AwsProviderBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "access_key"},
		{Name: "secret_key"},
		{Name: "profile"},
		{Name: "shared_credentials_file"},
		{Name: "region"},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "assume_role"},
	},
}

type providerResource interface {
	Get(key string) (string, bool, error)
	GetBlock(key string, schema *configschema.Block) (map[string]string, bool, error)
}

// NewAwsClient returns new AwsClient with configured session
func NewAwsClient(creds AwsCredentials) (*AwsClient, error) {
	log.Print("[INFO] Initialize AWS Client")

	config, err := getBaseConfig(creds)
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

// ConvertToCredentials converts to credentials from the given provider config
func ConvertToCredentials(providerConfig providerResource) (AwsCredentials, error) {
	ret := AwsCredentials{}

	accessKey, exists, err := providerConfig.Get("access_key")
	if err != nil {
		return ret, err
	}
	if exists {
		ret.AccessKey = accessKey
	}

	secretKey, exists, err := providerConfig.Get("secret_key")
	if err != nil {
		return ret, err
	}
	if exists {
		ret.SecretKey = secretKey
	}

	profile, exists, err := providerConfig.Get("profile")
	if err != nil {
		return ret, err
	}
	if exists {
		ret.Profile = profile
	}

	credsFile, exists, err := providerConfig.Get("shared_credentials_file")
	if err != nil {
		return ret, err
	}
	if exists {
		ret.CredsFile = credsFile
	}

	region, exists, err := providerConfig.Get("region")
	if err != nil {
		return ret, err
	}
	if exists {
		ret.Region = region
	}

	assumeRole, exists, err := providerConfig.GetBlock("assume_role", &configschema.Block{
		Attributes: map[string]*configschema.Attribute{
			"role_arn":     {Type: cty.String, Required: true},
			"session_name": {Type: cty.String},
			"external_id":  {Type: cty.String},
			"policy":       {Type: cty.String},
		},
	})
	if err != nil {
		return ret, err
	}
	if exists {
		ret.AssumeRoleARN = assumeRole["role_arn"]
		ret.AssumeRoleSessionName = assumeRole["session_name"]
		ret.AssumeRoleExternalID = assumeRole["external_id"]
		ret.AssumeRolePolicy = assumeRole["policy"]
	}

	return ret, nil
}

// Merge returns a merged credentials
func (c AwsCredentials) Merge(other AwsCredentials) AwsCredentials {
	if other.AccessKey != "" {
		c.AccessKey = other.AccessKey
	}
	if other.SecretKey != "" {
		c.SecretKey = other.SecretKey
	}
	if other.Profile != "" {
		c.Profile = other.Profile
	}
	if other.CredsFile != "" {
		c.CredsFile = other.CredsFile
	}
	if other.Region != "" {
		c.Region = other.Region
	}
	if other.AssumeRoleARN != "" {
		c.AssumeRoleARN = other.AssumeRoleARN
	}
	if other.AssumeRoleSessionName != "" {
		c.AssumeRoleSessionName = other.AssumeRoleSessionName
	}
	if other.AssumeRoleExternalID != "" {
		c.AssumeRoleExternalID = other.AssumeRoleExternalID
	}
	if other.AssumeRolePolicy != "" {
		c.AssumeRolePolicy = other.AssumeRolePolicy
	}
	return c
}

func getBaseConfig(creds AwsCredentials) (*awsbase.Config, error) {
	expandedCredsFile, err := homedir.Expand(creds.CredsFile)
	if err != nil {
		return nil, err
	}

	return &awsbase.Config{
		AccessKey:             creds.AccessKey,
		AssumeRoleARN:         creds.AssumeRoleARN,
		AssumeRoleExternalID:  creds.AssumeRoleExternalID,
		AssumeRolePolicy:      creds.AssumeRolePolicy,
		AssumeRoleSessionName: creds.AssumeRoleSessionName,
		SecretKey:             creds.SecretKey,
		Profile:               creds.Profile,
		CredsFilename:         expandedCredsFile,
		Region:                creds.Region,
	}, nil
}

// @see https://github.com/hashicorp/aws-sdk-go-base/blob/v0.3.0/session.go#L87
func formatBaseConfigError(err error) error {
	if strings.Contains(err.Error(), "No valid credential sources found for AWS Provider") {
		return errors.New("No valid credential sources found")
	}
	return err
}
