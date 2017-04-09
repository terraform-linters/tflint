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

func TestDetectAwsRouteInvalidVpcPeeringConnection(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		Response []*ec2.VpcPeeringConnection
		Issues   []*issue.Issue
	}{
		{
			Name: "VPC peering connection id is invalid",
			Src: `
resource "aws_route" "foo" {
    vpc_peering_connection_id = "pcx-1234abcd"
}`,
			Response: []*ec2.VpcPeeringConnection{
				{
					VpcPeeringConnectionId: aws.String("pcx-5678abcd"),
				},
				{
					VpcPeeringConnectionId: aws.String("pcx-abcd1234"),
				},
			},
			Issues: []*issue.Issue{
				{
					Type:    "ERROR",
					Message: "\"pcx-1234abcd\" is invalid VPC peering connection ID.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "VPC peering connection id is valid",
			Src: `
resource "aws_route" "foo" {
    vpc_peering_connection_id = "pcx-1234abcd"
}`,
			Response: []*ec2.VpcPeeringConnection{
				{
					VpcPeeringConnectionId: aws.String("pcx-1234abcd"),
				},
				{
					VpcPeeringConnectionId: aws.String("pcx-abcd1234"),
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
		ec2mock.EXPECT().DescribeVpcPeeringConnections(&ec2.DescribeVpcPeeringConnectionsInput{}).Return(&ec2.DescribeVpcPeeringConnectionsOutput{
			VpcPeeringConnections: tc.Response,
		}, nil)
		awsClient.Ec2 = ec2mock

		var issues = []*issue.Issue{}
		err := TestDetectByCreatorName(
			"CreateAwsRouteInvalidVpcPeeringConnectionDetector",
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
