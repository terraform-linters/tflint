package detector

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/mock"
)

func TestDetectAwsInstanceInvalidSubnet(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		Response []*ec2.Subnet
		Issues   []*issue.Issue
	}{
		{
			Name: "Subnet ID is invalid",
			Src: `
resource "aws_instance" "web" {
    subnet_id = "subnet-1234abcd"
}`,
			Response: []*ec2.Subnet{
				&ec2.Subnet{
					SubnetId: aws.String("subnet-12345678"),
				},
				&ec2.Subnet{
					SubnetId: aws.String("subnet-abcdefgh"),
				},
			},
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "ERROR",
					Message: "\"subnet-1234abcd\" is invalid subnet ID.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "key name is valid",
			Src: `
resource "aws_instance" "web" {
    subnet_id = "subnet-1234abcd"
}`,
			Response: []*ec2.Subnet{
				&ec2.Subnet{
					SubnetId: aws.String("subnet-1234abcd"),
				},
				&ec2.Subnet{
					SubnetId: aws.String("subnet-abcd1234"),
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
		ec2mock.EXPECT().DescribeSubnets(&ec2.DescribeSubnetsInput{}).Return(&ec2.DescribeSubnetsOutput{
			Subnets: tc.Response,
		}, nil)
		awsClient.Ec2 = ec2mock

		var issues = []*issue.Issue{}
		TestDetectByCreatorName(
			"CreateAwsInstanceInvalidSubnetDetector",
			tc.Src,
			"",
			c,
			awsClient,
			&issues,
		)

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", issues, tc.Issues, tc.Name)
		}
	}
}
