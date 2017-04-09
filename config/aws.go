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
	homedir "github.com/mitchellh/go-homedir"
	"github.com/wata727/tflint/logger"
)

type AwsClient struct {
	Iam         iamiface.IAMAPI
	Ec2         ec2iface.EC2API
	Rds         rdsiface.RDSAPI
	Elasticache elasticacheiface.ElastiCacheAPI
	Elb         elbiface.ELBAPI
	Elbv2       elbv2iface.ELBV2API
	Cache       *ResponseCache
}

func (c *Config) NewAwsClient() *AwsClient {
	client := &AwsClient{
		Cache: &ResponseCache{},
	}
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
	if c.HasAwsSharedCredentials() {
		l.Info("set AWS shared credentials")
		path, err := homedir.Expand("~/.aws/credentials")
		if err != nil {
			l.Error(err)
		}
		s = session.New(&aws.Config{
			Credentials: credentials.NewSharedCredentials(path, c.AwsCredentials["profile"]),
			Region:      aws.String(c.AwsCredentials["region"]),
		})
	}
	if c.HasAwsStaticCredentials() {
		l.Info("set AWS credentials")
		s = session.New(&aws.Config{
			Credentials: credentials.NewStaticCredentials(c.AwsCredentials["access_key"], c.AwsCredentials["secret_key"], ""),
			Region:      aws.String(c.AwsCredentials["region"]),
		})
	}

	return s
}

type ResponseCache struct {
	DescribeImagesOutput                     *ec2.DescribeImagesOutput
	DescribeKeyPairsOutput                   *ec2.DescribeKeyPairsOutput
	DescribeSubnetsOutput                    *ec2.DescribeSubnetsOutput
	DescribeSecurityGroupsOutput             *ec2.DescribeSecurityGroupsOutput
	DescribeVpcsOutput                       *ec2.DescribeVpcsOutput
	DescribeInstancesOutput                  *ec2.DescribeInstancesOutput
	DescribeAccountAttributesOutput          *ec2.DescribeAccountAttributesOutput
	DescribeRouteTablesOutput                *ec2.DescribeRouteTablesOutput
	DescribeInternetGatewaysOutput           *ec2.DescribeInternetGatewaysOutput
	DescribeEgressOnlyInternetGatewaysOutput *ec2.DescribeEgressOnlyInternetGatewaysOutput
	DescribeNatGatewaysOutput                *ec2.DescribeNatGatewaysOutput
	ListInstanceProfilesOutput               *iam.ListInstanceProfilesOutput
	DescribeDBSubnetGroupsOutput             *rds.DescribeDBSubnetGroupsOutput
	DescribeDBParameterGroupsOutput          *rds.DescribeDBParameterGroupsOutput
	DescribeOptionGroupsOutput               *rds.DescribeOptionGroupsOutput
	DescribeDBInstancesOutput                *rds.DescribeDBInstancesOutput
	DescribeCacheParameterGroupsOutput       *elasticache.DescribeCacheParameterGroupsOutput
	DescribeCacheSubnetGroupsOutput          *elasticache.DescribeCacheSubnetGroupsOutput
	DescribeCacheClustersOutput              *elasticache.DescribeCacheClustersOutput
	DescribeLoadBalancersOutput              *elbv2.DescribeLoadBalancersOutput
	DescribeClassicLoadBalancersOutput       *elb.DescribeLoadBalancersOutput
}

func (c *AwsClient) DescribeImages() (*ec2.DescribeImagesOutput, error) {
	if c.Cache.DescribeImagesOutput == nil {
		resp, err := c.Ec2.DescribeImages(&ec2.DescribeImagesInput{})
		if err != nil {
			return nil, err
		}
		c.Cache.DescribeImagesOutput = resp
	}
	return c.Cache.DescribeImagesOutput, nil
}

func (c *AwsClient) DescribeKeyPairs() (*ec2.DescribeKeyPairsOutput, error) {
	if c.Cache.DescribeKeyPairsOutput == nil {
		resp, err := c.Ec2.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{})
		if err != nil {
			return nil, err
		}
		c.Cache.DescribeKeyPairsOutput = resp
	}
	return c.Cache.DescribeKeyPairsOutput, nil
}

func (c *AwsClient) DescribeSubnets() (*ec2.DescribeSubnetsOutput, error) {
	if c.Cache.DescribeSubnetsOutput == nil {
		resp, err := c.Ec2.DescribeSubnets(&ec2.DescribeSubnetsInput{})
		if err != nil {
			return nil, err
		}
		c.Cache.DescribeSubnetsOutput = resp
	}
	return c.Cache.DescribeSubnetsOutput, nil
}

func (c *AwsClient) DescribeSecurityGroups() (*ec2.DescribeSecurityGroupsOutput, error) {
	if c.Cache.DescribeSecurityGroupsOutput == nil {
		resp, err := c.Ec2.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})
		if err != nil {
			return nil, err
		}
		c.Cache.DescribeSecurityGroupsOutput = resp
	}
	return c.Cache.DescribeSecurityGroupsOutput, nil
}

func (c *AwsClient) DescribeVpcs() (*ec2.DescribeVpcsOutput, error) {
	if c.Cache.DescribeVpcsOutput == nil {
		resp, err := c.Ec2.DescribeVpcs(&ec2.DescribeVpcsInput{})
		if err != nil {
			return nil, err
		}
		c.Cache.DescribeVpcsOutput = resp
	}
	return c.Cache.DescribeVpcsOutput, nil
}

func (c *AwsClient) DescribeInstances() (*ec2.DescribeInstancesOutput, error) {
	if c.Cache.DescribeInstancesOutput == nil {
		resp, err := c.Ec2.DescribeInstances(&ec2.DescribeInstancesInput{})
		if err != nil {
			return nil, err
		}
		c.Cache.DescribeInstancesOutput = resp
	}
	return c.Cache.DescribeInstancesOutput, nil
}

func (c AwsClient) DescribeAccountAttributes() (*ec2.DescribeAccountAttributesOutput, error) {
	if c.Cache.DescribeAccountAttributesOutput == nil {
		resp, err := c.Ec2.DescribeAccountAttributes(&ec2.DescribeAccountAttributesInput{})
		if err != nil {
			return nil, err
		}
		c.Cache.DescribeAccountAttributesOutput = resp
	}
	return c.Cache.DescribeAccountAttributesOutput, nil
}

func (c AwsClient) DescribeRouteTables() (*ec2.DescribeRouteTablesOutput, error) {
	if c.Cache.DescribeRouteTablesOutput == nil {
		resp, err := c.Ec2.DescribeRouteTables(&ec2.DescribeRouteTablesInput{})
		if err != nil {
			return nil, err
		}
		c.Cache.DescribeRouteTablesOutput = resp
	}
	return c.Cache.DescribeRouteTablesOutput, nil
}

func (c AwsClient) DescribeInternetGateways() (*ec2.DescribeInternetGatewaysOutput, error) {
	if c.Cache.DescribeInternetGatewaysOutput == nil {
		resp, err := c.Ec2.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{})
		if err != nil {
			return nil, err
		}
		c.Cache.DescribeInternetGatewaysOutput = resp
	}
	return c.Cache.DescribeInternetGatewaysOutput, nil
}

func (c AwsClient) DescribeEgressOnlyInternetGateways() (*ec2.DescribeEgressOnlyInternetGatewaysOutput, error) {
	if c.Cache.DescribeEgressOnlyInternetGatewaysOutput == nil {
		resp, err := c.Ec2.DescribeEgressOnlyInternetGateways(&ec2.DescribeEgressOnlyInternetGatewaysInput{})
		if err != nil {
			return nil, err
		}
		c.Cache.DescribeEgressOnlyInternetGatewaysOutput = resp
	}
	return c.Cache.DescribeEgressOnlyInternetGatewaysOutput, nil
}

func (c AwsClient) DescribeNatGateways() (*ec2.DescribeNatGatewaysOutput, error) {
	if c.Cache.DescribeNatGatewaysOutput == nil {
		resp, err := c.Ec2.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{})
		if err != nil {
			return nil, err
		}
		c.Cache.DescribeNatGatewaysOutput = resp
	}
	return c.Cache.DescribeNatGatewaysOutput, nil
}

func (c *AwsClient) ListInstanceProfiles() (*iam.ListInstanceProfilesOutput, error) {
	if c.Cache.ListInstanceProfilesOutput == nil {
		resp, err := c.Iam.ListInstanceProfiles(&iam.ListInstanceProfilesInput{})
		if err != nil {
			return nil, err
		}
		c.Cache.ListInstanceProfilesOutput = resp
	}
	return c.Cache.ListInstanceProfilesOutput, nil
}

func (c *AwsClient) DescribeDBSubnetGroups() (*rds.DescribeDBSubnetGroupsOutput, error) {
	if c.Cache.DescribeDBSubnetGroupsOutput == nil {
		resp, err := c.Rds.DescribeDBSubnetGroups(&rds.DescribeDBSubnetGroupsInput{})
		if err != nil {
			return nil, err
		}
		c.Cache.DescribeDBSubnetGroupsOutput = resp
	}
	return c.Cache.DescribeDBSubnetGroupsOutput, nil
}

func (c *AwsClient) DescribeDBParameterGroups() (*rds.DescribeDBParameterGroupsOutput, error) {
	if c.Cache.DescribeDBParameterGroupsOutput == nil {
		resp, err := c.Rds.DescribeDBParameterGroups(&rds.DescribeDBParameterGroupsInput{})
		if err != nil {
			return nil, err
		}
		c.Cache.DescribeDBParameterGroupsOutput = resp
	}
	return c.Cache.DescribeDBParameterGroupsOutput, nil
}

func (c *AwsClient) DescribeOptionGroups() (*rds.DescribeOptionGroupsOutput, error) {
	if c.Cache.DescribeOptionGroupsOutput == nil {
		resp, err := c.Rds.DescribeOptionGroups(&rds.DescribeOptionGroupsInput{})
		if err != nil {
			return nil, err
		}
		c.Cache.DescribeOptionGroupsOutput = resp
	}
	return c.Cache.DescribeOptionGroupsOutput, nil
}

func (c *AwsClient) DescribeDBInstances() (*rds.DescribeDBInstancesOutput, error) {
	if c.Cache.DescribeDBInstancesOutput == nil {
		resp, err := c.Rds.DescribeDBInstances(&rds.DescribeDBInstancesInput{})
		if err != nil {
			return nil, err
		}
		c.Cache.DescribeDBInstancesOutput = resp
	}
	return c.Cache.DescribeDBInstancesOutput, nil
}

func (c *AwsClient) DescribeCacheParameterGroups() (*elasticache.DescribeCacheParameterGroupsOutput, error) {
	if c.Cache.DescribeCacheParameterGroupsOutput == nil {
		resp, err := c.Elasticache.DescribeCacheParameterGroups(&elasticache.DescribeCacheParameterGroupsInput{})
		if err != nil {
			return nil, err
		}
		c.Cache.DescribeCacheParameterGroupsOutput = resp
	}
	return c.Cache.DescribeCacheParameterGroupsOutput, nil
}

func (c *AwsClient) DescribeCacheSubnetGroups() (*elasticache.DescribeCacheSubnetGroupsOutput, error) {
	if c.Cache.DescribeCacheSubnetGroupsOutput == nil {
		resp, err := c.Elasticache.DescribeCacheSubnetGroups(&elasticache.DescribeCacheSubnetGroupsInput{})
		if err != nil {
			return nil, err
		}
		c.Cache.DescribeCacheSubnetGroupsOutput = resp
	}
	return c.Cache.DescribeCacheSubnetGroupsOutput, nil
}

func (c *AwsClient) DescribeCacheClusters() (*elasticache.DescribeCacheClustersOutput, error) {
	if c.Cache.DescribeCacheClustersOutput == nil {
		resp, err := c.Elasticache.DescribeCacheClusters(&elasticache.DescribeCacheClustersInput{})
		if err != nil {
			return nil, err
		}
		c.Cache.DescribeCacheClustersOutput = resp
	}
	return c.Cache.DescribeCacheClustersOutput, nil
}

func (c *AwsClient) DescribeLoadBalancers() (*elbv2.DescribeLoadBalancersOutput, error) {
	if c.Cache.DescribeLoadBalancersOutput == nil {
		resp, err := c.Elbv2.DescribeLoadBalancers(&elbv2.DescribeLoadBalancersInput{})
		if err != nil {
			return nil, err
		}
		c.Cache.DescribeLoadBalancersOutput = resp
	}
	return c.Cache.DescribeLoadBalancersOutput, nil
}

func (c *AwsClient) DescribeClassicLoadBalancers() (*elb.DescribeLoadBalancersOutput, error) {
	if c.Cache.DescribeClassicLoadBalancersOutput == nil {
		resp, err := c.Elb.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})
		if err != nil {
			return nil, err
		}
		c.Cache.DescribeClassicLoadBalancersOutput = resp
	}
	return c.Cache.DescribeClassicLoadBalancersOutput, nil
}
