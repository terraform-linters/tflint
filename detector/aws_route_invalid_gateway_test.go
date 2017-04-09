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

func TestDetectAwsRouteInvalidGateway(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		Response []*ec2.InternetGateway
		Issues   []*issue.Issue
	}{
		{
			Name: "gateway id is invalid",
			Src: `
resource "aws_route" "foo" {
    gateway_id = "eigw-1234abcd"
}`,
			Response: []*ec2.InternetGateway{
				{
					InternetGatewayId: aws.String("igw-1234abcd"),
				},
				{
					InternetGatewayId: aws.String("igw-abcd1234"),
				},
			},
			Issues: []*issue.Issue{
				{
					Type:    "ERROR",
					Message: "\"eigw-1234abcd\" is invalid internet gateway ID.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "gateway id is valid",
			Src: `
resource "aws_route" "foo" {
    gateway_id = "igw-1234abcd"
}`,
			Response: []*ec2.InternetGateway{
				{
					InternetGatewayId: aws.String("igw-1234abcd"),
				},
				{
					InternetGatewayId: aws.String("igw-abcd1234"),
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
		ec2mock.EXPECT().DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{}).Return(&ec2.DescribeInternetGatewaysOutput{
			InternetGateways: tc.Response,
		}, nil)
		awsClient.Ec2 = ec2mock

		var issues = []*issue.Issue{}
		err := TestDetectByCreatorName(
			"CreateAwsRouteInvalidGatewayDetector",
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
