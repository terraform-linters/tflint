package client

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/rds"
)

// DescribeSecurityGroups is a wrapper of DescribeSecurityGroups
func (c *AwsClient) DescribeSecurityGroups() (map[string]bool, error) {
	ret := map[string]bool{}
	resp, err := c.EC2.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})
	if err != nil {
		return ret, err
	}
	for _, sg := range resp.SecurityGroups {
		ret[*sg.GroupId] = true
	}
	return ret, err
}

// DescribeSubnets is a wrapper of DescribeSubnets
func (c *AwsClient) DescribeSubnets() (map[string]bool, error) {
	ret := map[string]bool{}
	resp, err := c.EC2.DescribeSubnets(&ec2.DescribeSubnetsInput{})
	if err != nil {
		return ret, err
	}
	for _, subnet := range resp.Subnets {
		ret[*subnet.SubnetId] = true
	}
	return ret, err
}

// DescribeDBSubnetGroups is a wrapper of DescribeDBSubnetGroups
func (c *AwsClient) DescribeDBSubnetGroups() (map[string]bool, error) {
	ret := map[string]bool{}
	resp, err := c.RDS.DescribeDBSubnetGroups(&rds.DescribeDBSubnetGroupsInput{})
	if err != nil {
		return ret, err
	}
	for _, subnetGroup := range resp.DBSubnetGroups {
		ret[*subnetGroup.DBSubnetGroupName] = true
	}
	return ret, err
}

// DescribeOptionGroups is a wrapper of DescribeOptionGroups
func (c *AwsClient) DescribeOptionGroups() (map[string]bool, error) {
	ret := map[string]bool{}
	resp, err := c.RDS.DescribeOptionGroups(&rds.DescribeOptionGroupsInput{})
	if err != nil {
		return ret, err
	}
	for _, optionGroup := range resp.OptionGroupsList {
		ret[*optionGroup.OptionGroupName] = true
	}
	return ret, err
}

// DescribeDBParameterGroups is a wrapper of DescribeDBParameterGroups
func (c *AwsClient) DescribeDBParameterGroups() (map[string]bool, error) {
	ret := map[string]bool{}
	resp, err := c.RDS.DescribeDBParameterGroups(&rds.DescribeDBParameterGroupsInput{})
	if err != nil {
		return ret, err
	}
	for _, parameterGroup := range resp.DBParameterGroups {
		ret[*parameterGroup.DBParameterGroupName] = true
	}
	return ret, err
}

// DescribeCacheParameterGroups is a wrapper of DescribeCacheParameterGroups
func (c *AwsClient) DescribeCacheParameterGroups() (map[string]bool, error) {
	ret := map[string]bool{}
	resp, err := c.ElastiCache.DescribeCacheParameterGroups(&elasticache.DescribeCacheParameterGroupsInput{})
	if err != nil {
		return ret, err
	}
	for _, parameterGroup := range resp.CacheParameterGroups {
		ret[*parameterGroup.CacheParameterGroupName] = true
	}
	return ret, err
}

// DescribeCacheSubnetGroups is a wrapper of DescribeCacheSubnetGroups
func (c *AwsClient) DescribeCacheSubnetGroups() (map[string]bool, error) {
	ret := map[string]bool{}
	resp, err := c.ElastiCache.DescribeCacheSubnetGroups(&elasticache.DescribeCacheSubnetGroupsInput{})
	if err != nil {
		return ret, err
	}
	for _, subnetGroup := range resp.CacheSubnetGroups {
		ret[*subnetGroup.CacheSubnetGroupName] = true
	}
	return ret, err
}

// DescribeInstances is a wrapper of DescribeInstances
func (c *AwsClient) DescribeInstances() (map[string]bool, error) {
	ret := map[string]bool{}
	resp, err := c.EC2.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err != nil {
		return ret, err
	}
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			ret[*instance.InstanceId] = true
		}
	}
	return ret, err
}

// ListInstanceProfiles is a wrapper of ListInstanceProfiles
func (c *AwsClient) ListInstanceProfiles() (map[string]bool, error) {
	ret := map[string]bool{}
	resp, err := c.IAM.ListInstanceProfiles(&iam.ListInstanceProfilesInput{})
	if err != nil {
		return ret, err
	}
	for _, iamProfile := range resp.InstanceProfiles {
		ret[*iamProfile.InstanceProfileName] = true
	}
	return ret, err
}

// DescribeKeyPairs is a wrapper of DescribeKeyPairs
func (c *AwsClient) DescribeKeyPairs() (map[string]bool, error) {
	ret := map[string]bool{}
	resp, err := c.EC2.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{})
	if err != nil {
		return ret, err
	}
	for _, keyPair := range resp.KeyPairs {
		ret[*keyPair.KeyName] = true
	}
	return ret, err
}

// DescribeEgressOnlyInternetGateways is wrapper of DescribeEgressOnlyInternetGateways
func (c *AwsClient) DescribeEgressOnlyInternetGateways() (map[string]bool, error) {
	ret := map[string]bool{}
	resp, err := c.EC2.DescribeEgressOnlyInternetGateways(&ec2.DescribeEgressOnlyInternetGatewaysInput{})
	if err != nil {
		return ret, err
	}
	for _, egateway := range resp.EgressOnlyInternetGateways {
		ret[*egateway.EgressOnlyInternetGatewayId] = true
	}
	return ret, err
}

// DescribeInternetGateways is a wrapper of DescribeInternetGateways
func (c *AwsClient) DescribeInternetGateways() (map[string]bool, error) {
	ret := map[string]bool{}
	resp, err := c.EC2.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{})
	if err != nil {
		return ret, err
	}
	for _, gateway := range resp.InternetGateways {
		ret[*gateway.InternetGatewayId] = true
	}
	return ret, err
}

// DescribeNatGateways is a wrapper of DescribeNatGateways
func (c *AwsClient) DescribeNatGateways() (map[string]bool, error) {
	ret := map[string]bool{}
	resp, err := c.EC2.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{})
	if err != nil {
		return ret, err
	}
	for _, ngateway := range resp.NatGateways {
		ret[*ngateway.NatGatewayId] = true
	}
	return ret, err
}

// DescribeNetworkInterfaces is a wrapper of DescribeNetworkInterfaces
func (c *AwsClient) DescribeNetworkInterfaces() (map[string]bool, error) {
	ret := map[string]bool{}
	resp, err := c.EC2.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{})
	if err != nil {
		return ret, err
	}
	for _, networkInterface := range resp.NetworkInterfaces {
		ret[*networkInterface.NetworkInterfaceId] = true
	}
	return ret, err
}

// DescribeRouteTables is a wrapper of DescribeRouteTables
func (c *AwsClient) DescribeRouteTables() (map[string]bool, error) {
	ret := map[string]bool{}
	resp, err := c.EC2.DescribeRouteTables(&ec2.DescribeRouteTablesInput{})
	if err != nil {
		return ret, err
	}
	for _, routeTable := range resp.RouteTables {
		ret[*routeTable.RouteTableId] = true
	}
	return ret, err
}

// DescribeVpcPeeringConnections is a wrapper of DescribeVpcPeeringConnections
func (c *AwsClient) DescribeVpcPeeringConnections() (map[string]bool, error) {
	ret := map[string]bool{}
	resp, err := c.EC2.DescribeVpcPeeringConnections(&ec2.DescribeVpcPeeringConnectionsInput{})
	if err != nil {
		return ret, err
	}
	for _, vpcPeeringConnection := range resp.VpcPeeringConnections {
		ret[*vpcPeeringConnection.VpcPeeringConnectionId] = true
	}
	return ret, err
}
