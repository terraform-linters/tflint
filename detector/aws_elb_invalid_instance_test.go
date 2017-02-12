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

func TestDetectAwsELBInvalidInstance(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		Response []*ec2.Instance
		Issues   []*issue.Issue
	}{
		{
			Name: "Instance is invalid",
			Src: `
resource "aws_elb" "balancer" {
    instances = [
        "i-1234abcd",
        "i-abcd1234",
    ]
}`,
			Response: []*ec2.Instance{
				&ec2.Instance{
					InstanceId: aws.String("i-12345678"),
				},
				&ec2.Instance{
					InstanceId: aws.String("i-abcdefgh"),
				},
			},
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "ERROR",
					Message: "\"i-1234abcd\" is invalid instance.",
					Line:    4,
					File:    "test.tf",
				},
				&issue.Issue{
					Type:    "ERROR",
					Message: "\"i-abcd1234\" is invalid instance.",
					Line:    5,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "Instance is valid",
			Src: `
resource "aws_elb" "balancer" {
    instances = [
        "i-1234abcd",
        "i-abcd1234",
    ]
}`,
			Response: []*ec2.Instance{
				&ec2.Instance{
					InstanceId: aws.String("i-1234abcd"),
				},
				&ec2.Instance{
					InstanceId: aws.String("i-abcd1234"),
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
		ec2mock.EXPECT().DescribeInstances(&ec2.DescribeInstancesInput{}).Return(&ec2.DescribeInstancesOutput{
			Reservations: []*ec2.Reservation{
				&ec2.Reservation{
					Instances: tc.Response,
				},
			},
		}, nil)
		awsClient.Ec2 = ec2mock

		var issues = []*issue.Issue{}
		TestDetectByCreatorName(
			"CreateAwsELBInvalidInstanceDetector",
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
