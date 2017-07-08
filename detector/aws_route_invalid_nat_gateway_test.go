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

func TestDetectAwsRouteInvalidNatGateway(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		Response []*ec2.NatGateway
		Issues   []*issue.Issue
	}{
		{
			Name: "NAT gateway id is invalid",
			Src: `
resource "aws_route" "foo" {
    nat_gateway_id = "nat-1234abcd"
}`,
			Response: []*ec2.NatGateway{
				{
					NatGatewayId: aws.String("nat-5678abcd"),
				},
				{
					NatGatewayId: aws.String("nat-abcd1234"),
				},
			},
			Issues: []*issue.Issue{
				{
					Detector: "aws_route_invalid_nat_gateway",
					Type:     "ERROR",
					Message:  "\"nat-1234abcd\" is invalid NAT gateway ID.",
					Line:     3,
					File:     "test.tf",
				},
			},
		},
		{
			Name: "NAT gateway id is valid",
			Src: `
resource "aws_route" "foo" {
    nat_gateway_id = "nat-1234abcd"
}`,
			Response: []*ec2.NatGateway{
				{
					NatGatewayId: aws.String("nat-1234abcd"),
				},
				{
					NatGatewayId: aws.String("nat-abcd1234"),
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
		ec2mock.EXPECT().DescribeNatGateways(&ec2.DescribeNatGatewaysInput{}).Return(&ec2.DescribeNatGatewaysOutput{
			NatGateways: tc.Response,
		}, nil)
		awsClient.Ec2 = ec2mock

		var issues = []*issue.Issue{}
		err := TestDetectByCreatorName(
			"CreateAwsRouteInvalidNatGatewayDetector",
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
