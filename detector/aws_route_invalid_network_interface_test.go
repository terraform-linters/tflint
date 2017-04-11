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

func TestDetectAwsRouteInvalidNetworkInterface(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		Response []*ec2.NetworkInterface
		Issues   []*issue.Issue
	}{
		{
			Name: "network interface id is invalid",
			Src: `
resource "aws_route" "foo" {
    network_interface_id = "eni-1234abcd"
}`,
			Response: []*ec2.NetworkInterface{
				{
					NetworkInterfaceId: aws.String("eni-5678abcd"),
				},
				{
					NetworkInterfaceId: aws.String("eni-abcd1234"),
				},
			},
			Issues: []*issue.Issue{
				{
					Type:    "ERROR",
					Message: "\"eni-1234abcd\" is invalid network interface ID.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "network interfaec id is valid",
			Src: `
resource "aws_route" "foo" {
    network_interface_id = "eni-1234abcd"
}`,
			Response: []*ec2.NetworkInterface{
				{
					NetworkInterfaceId: aws.String("eni-1234abcd"),
				},
				{
					NetworkInterfaceId: aws.String("eni-abcd1234"),
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
		ec2mock.EXPECT().DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{}).Return(&ec2.DescribeNetworkInterfacesOutput{
			NetworkInterfaces: tc.Response,
		}, nil)
		awsClient.Ec2 = ec2mock

		var issues = []*issue.Issue{}
		err := TestDetectByCreatorName(
			"CreateAwsRouteInvalidNetworkInterfaceDetector",
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
