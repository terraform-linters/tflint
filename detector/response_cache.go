package detector

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/rds"
)

type ResponseCache struct {
	DescribeImagesOutput               *ec2.DescribeImagesOutput
	DescribeKeyPairsOutput             *ec2.DescribeKeyPairsOutput
	DescribeSubnetsOutput              *ec2.DescribeSubnetsOutput
	DescribeSecurityGroupsOutput       *ec2.DescribeSecurityGroupsOutput
	DescribeVpcsOutput                 *ec2.DescribeVpcsOutput
	ListInstanceProfilesOutput         *iam.ListInstanceProfilesOutput
	DescribeInstancesOutput            *ec2.DescribeInstancesOutput
	DescribeDBSubnetGroupsOutput       *rds.DescribeDBSubnetGroupsOutput
	DescribeDBParameterGroupsOutput    *rds.DescribeDBParameterGroupsOutput
	DescribeOptionGroupsOutput         *rds.DescribeOptionGroupsOutput
	DescribeDBInstancesOutput          *rds.DescribeDBInstancesOutput
	DescribeCacheParameterGroupsOutput *elasticache.DescribeCacheParameterGroupsOutput
	DescribeCacheSubnetGroupsOutput    *elasticache.DescribeCacheSubnetGroupsOutput
	DescribeCacheClustersOutput        *elasticache.DescribeCacheClustersOutput
	DescribeLoadBalancersOutput        *elbv2.DescribeLoadBalancersOutput
	DescribeClassicLoadBalancersOutput *elb.DescribeLoadBalancersOutput
}
