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

func TestDetectAwsELBInvalidSubnet(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		Response []*ec2.Subnet
		Issues   []*issue.Issue
	}{
		{
			Name: "Subnet ID is invalid",
			Src: `
resource "aws_elb" "balancer" {
    subnets = [
        "subnet-1234abcd",
        "subnet-abcd1234",
    ]
}`,
			Response: []*ec2.Subnet{
				{
					SubnetId: aws.String("subnet-12345678"),
				},
				{
					SubnetId: aws.String("subnet-abcdefgh"),
				},
			},
			Issues: []*issue.Issue{
				{
					Type:    "ERROR",
					Message: "\"subnet-1234abcd\" is invalid subnet ID.",
					Line:    4,
					File:    "test.tf",
				},
				{
					Type:    "ERROR",
					Message: "\"subnet-abcd1234\" is invalid subnet ID.",
					Line:    5,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "Subnet ID is valid",
			Src: `
resource "aws_elb" "balancer" {
    subnets = [
        "subnet-1234abcd",
        "subnet-abcd1234",
    ]
}`,
			Response: []*ec2.Subnet{
				{
					SubnetId: aws.String("subnet-1234abcd"),
				},
				{
					SubnetId: aws.String("subnet-abcd1234"),
				},
			},
			Issues: []*issue.Issue{},
		},
		{
			Name: "use list variable",
			Src: `
variable "subnets" {
    default = ["subnet-1234abcd", "subnet-abcd1234"]
}

resource "aws_elb" "balancer" {
    subnets = "${var.subnets}"
}`,
			Response: []*ec2.Subnet{
				{
					SubnetId: aws.String("subnet-12345678"),
				},
				{
					SubnetId: aws.String("subnet-abcdefgh"),
				},
			},
			Issues: []*issue.Issue{
				{
					Type:    "ERROR",
					Message: "\"subnet-1234abcd\" is invalid subnet ID.",
					Line:    7,
					File:    "test.tf",
				},
				{
					Type:    "ERROR",
					Message: "\"subnet-abcd1234\" is invalid subnet ID.",
					Line:    7,
					File:    "test.tf",
				},
			},
		},
	}

	for _, tc := range cases {
		c := config.Init()
		c.DeepCheck = true

		awsClient := c.NewAwsClient()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ec2mock := mock.NewMockEC2API(ctrl)
		ec2mock.EXPECT().DescribeSubnets(&ec2.DescribeSubnetsInput{}).Return(&ec2.DescribeSubnetsOutput{
			Subnets: tc.Response,
		}, nil)
		awsClient.Ec2 = ec2mock

		var issues = []*issue.Issue{}
		err := TestDetectByCreatorName(
			"CreateAwsELBInvalidSubnetDetector",
			tc.Src,
			"",
			c,
			awsClient,
			&issues,
		)
		if err != nil {
			t.Fatalf("\nERROR: %s", err)
		}

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(issues), pp.Sprint(tc.Issues), tc.Name)
		}
	}
}
