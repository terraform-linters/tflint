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
				{
					KeyName: aws.String("hogehoge"),
				},
				{
					KeyName: aws.String("fugafuga"),
				},
			},
			Issues: []*issue.Issue{
				{
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
				{
					KeyName: aws.String("foo"),
				},
				{
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
		err := TestDetectByCreatorName(
			"CreateAwsInstanceInvalidKeyNameDetector",
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
