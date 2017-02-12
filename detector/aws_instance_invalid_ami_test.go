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

func TestDetectAwsInstanceInvalidAMI(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		Response []*ec2.Image
		Issues   []*issue.Issue
	}{
		{
			Name: "AMI is invalid",
			Src: `
resource "aws_instance" "web" {
    ami = "ami-1234abcd"
}`,
			Response: []*ec2.Image{
				&ec2.Image{
					ImageId: aws.String("ami-0c11b26d"),
				},
				&ec2.Image{
					ImageId: aws.String("ami-9ad76sd1"),
				},
			},
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "ERROR",
					Message: "\"ami-1234abcd\" is invalid AMI.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "AMI is valid",
			Src: `
resource "aws_instance" "web" {
    ami = "ami-0c11b26d"
}`,
			Response: []*ec2.Image{
				&ec2.Image{
					ImageId: aws.String("ami-0c11b26d"),
				},
				&ec2.Image{
					ImageId: aws.String("ami-9ad76sd1"),
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
		ec2mock.EXPECT().DescribeImages(&ec2.DescribeImagesInput{}).Return(&ec2.DescribeImagesOutput{
			Images: tc.Response,
		}, nil)
		awsClient.Ec2 = ec2mock

		var issues = []*issue.Issue{}
		err := TestDetectByCreatorName(
			"CreateAwsInstanceInvalidAMIDetector",
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
