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

func TestDetectAwsRouteInvalidRouteTable(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		Response []*ec2.RouteTable
		Issues   []*issue.Issue
	}{
		{
			Name: "route table id is invalid",
			Src: `
resource "aws_route" "foo" {
    route_table_id = "rtb-nat-gw-a"
}`,
			Response: []*ec2.RouteTable{
				{
					RouteTableId: aws.String("rtb-1234abcd"),
				},
				{
					RouteTableId: aws.String("rtb-abcd1234"),
				},
			},
			Issues: []*issue.Issue{
				{
					Type:    "ERROR",
					Message: "\"rtb-nat-gw-a\" is invalid route table ID.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "route table id is valid",
			Src: `
resource "aws_route" "foo" {
    route_table_id = "rtb-1234abcd"
}`,
			Response: []*ec2.RouteTable{
				{
					RouteTableId: aws.String("rtb-1234abcd"),
				},
				{
					RouteTableId: aws.String("rtb-abcd1234"),
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
		ec2mock.EXPECT().DescribeRouteTables(&ec2.DescribeRouteTablesInput{}).Return(&ec2.DescribeRouteTablesOutput{
			RouteTables: tc.Response,
		}, nil)
		awsClient.Ec2 = ec2mock

		var issues = []*issue.Issue{}
		err := TestDetectByCreatorName(
			"CreateAwsRouteInvalidRouteTableDetector",
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
