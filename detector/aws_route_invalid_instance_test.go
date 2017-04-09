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

func TestDetectAwsRouteInvalidInstance(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		Response []*ec2.Instance
		Issues   []*issue.Issue
	}{
		{
			Name: "instance id is invalid",
			Src: `
resource "aws_route" "foo" {
    instance_id = "i-1234abcd"
}`,
			Response: []*ec2.Instance{
				{
					InstanceId: aws.String("i-5678abcd"),
				},
				{
					InstanceId: aws.String("i-abcd1234"),
				},
			},
			Issues: []*issue.Issue{
				{
					Type:    "ERROR",
					Message: "\"i-1234abcd\" is invalid instance ID.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "instance id is valid",
			Src: `
resource "aws_route" "foo" {
    instance_id = "i-1234abcd"
}`,
			Response: []*ec2.Instance{
				{
					InstanceId: aws.String("i-1234abcd"),
				},
				{
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
				{
					Instances: tc.Response,
				},
			},
		}, nil)
		awsClient.Ec2 = ec2mock

		var issues = []*issue.Issue{}
		err := TestDetectByCreatorName(
			"CreateAwsRouteInvalidInstanceDetector",
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
