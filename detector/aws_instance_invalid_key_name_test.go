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

func TestDetectAwsInstanceInvalidKeyName(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		Response []*ec2.KeyPairInfo
		Issues   []*issue.Issue
	}{
		{
			Name: "Key name is invalid",
			Src: `
resource "aws_instance" "web" {
    key_name = "foo"
}`,
			Response: []*ec2.KeyPairInfo{
				&ec2.KeyPairInfo{
					KeyName: aws.String("hogehoge"),
				},
				&ec2.KeyPairInfo{
					KeyName: aws.String("fugafuga"),
				},
			},
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "ERROR",
					Message: "\"foo\" is invalid key name.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "key name is valid",
			Src: `
resource "aws_instance" "web" {
    key_name = "foo"
}`,
			Response: []*ec2.KeyPairInfo{
				&ec2.KeyPairInfo{
					KeyName: aws.String("foo"),
				},
				&ec2.KeyPairInfo{
					KeyName: aws.String("bar"),
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
		ec2mock.EXPECT().DescribeKeyPairs(&ec2.DescribeKeyPairsInput{}).Return(&ec2.DescribeKeyPairsOutput{
			KeyPairs: tc.Response,
		}, nil)
		awsClient.Ec2 = ec2mock

		var issues = []*issue.Issue{}
		TestDetectByCreatorName(
			"CreateAwsInstanceInvalidKeyNameDetector",
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
