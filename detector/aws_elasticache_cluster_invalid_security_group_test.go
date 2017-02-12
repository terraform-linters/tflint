package detector

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"
	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/mock"
)

func TestDetectAwsElastiCacheClusterInvalidSecurityGroup(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		Response []*ec2.SecurityGroup
		Issues   []*issue.Issue
	}{
		{
			Name: "security group is invalid",
			Src: `
resource "aws_elasticache_cluster" "redis" {
    security_group_ids = [
        "sg-1234abcd",
        "sg-abcd1234",
    ]
}`,
			Response: []*ec2.SecurityGroup{
				&ec2.SecurityGroup{
					GroupId: aws.String("sg-12345678"),
				},
				&ec2.SecurityGroup{
					GroupId: aws.String("sg-abcdefgh"),
				},
			},
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "ERROR",
					Message: "\"sg-1234abcd\" is invalid security group.",
					Line:    4,
					File:    "test.tf",
				},
				&issue.Issue{
					Type:    "ERROR",
					Message: "\"sg-abcd1234\" is invalid security group.",
					Line:    5,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "security group is valid",
			Src: `
resource "aws_elasticache_cluster" "redis" {
    security_group_ids = [
        "sg-1234abcd",
        "sg-abcd1234",
    ]
}`,
			Response: []*ec2.SecurityGroup{
				&ec2.SecurityGroup{
					GroupId: aws.String("sg-1234abcd"),
				},
				&ec2.SecurityGroup{
					GroupId: aws.String("sg-abcd1234"),
				},
			},
			Issues: []*issue.Issue{},
		},
	}

	for _, tc := range cases {
		c := config.Init()
		c.DeepCheck = true

		awsClient := c.NewAwsClient()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ec2mock := mock.NewMockEC2API(ctrl)
		ec2mock.EXPECT().DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{}).Return(&ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: tc.Response,
		}, nil)
		awsClient.Ec2 = ec2mock

		var issues = []*issue.Issue{}
		TestDetectByCreatorName(
			"CreateAwsElastiCacheClusterInvalidSecurityGroupDetector",
			tc.Src,
			"",
			c,
			awsClient,
			&issues,
		)

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(issues), pp.Sprint(tc.Issues), tc.Name)
		}
	}
}
