package detector

import (
	"reflect"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
)

func TestDetectAwsRouteNotSpecifiedTarget(t *testing.T) {
	cases := []struct {
		Name   string
		Src    string
		Issues []*issue.Issue
	}{
		{
			Name: "route target is not specified",
			Src: `
resource "aws_route" "foo" {
    route_table_id = "rtb-1234abcd"
}`,
			Issues: []*issue.Issue{
				{
					Detector: "aws_route_not_specified_target",
					Type:     "ERROR",
					Message:  "route target is not specified, each route must contain either a gateway_id, egress_only_gateway_id a nat_gateway_id, an instance_id or a vpc_peering_connection_id or a network_interface_id.",
					Line:     2,
					File:     "test.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_route_not_specified_target.md",
				},
			},
		},
		{
			Name: "gateway_id is specified",
			Src: `
resource "aws_route" "foo" {
    route_table_id = "rtb-1234abcd"
    gateway_id = "igw-1234abcd"
}`,
			Issues: []*issue.Issue{},
		},
		{
			Name: "egress_only_gateway_id is specified",
			Src: `
resource "aws_route" "foo" {
    route_table_id = "rtb-1234abcd"
    egress_only_gateway_id = "eigw-1234abcd"
}`,
			Issues: []*issue.Issue{},
		},
		{
			Name: "nat_gateway_id is specified",
			Src: `
resource "aws_route" "foo" {
    route_table_id = "rtb-1234abcd"
    nat_gateway_id = "nat-1234abcd"
}`,
			Issues: []*issue.Issue{},
		},
		{
			Name: "instance_id is specified",
			Src: `
resource "aws_route" "foo" {
    route_table_id = "rtb-1234abcd"
    instance_id = "i-1234abcd"
}`,
			Issues: []*issue.Issue{},
		},
		{
			Name: "vpc_peering_connection_id is specified",
			Src: `
resource "aws_route" "foo" {
    route_table_id = "rtb-1234abcd"
    vpc_peering_connection_id = "pcx-1234abcd"
}`,
			Issues: []*issue.Issue{},
		},
		{
			Name: "network_interface_id is specified",
			Src: `
resource "aws_route" "foo" {
    route_table_id = "rtb-1234abcd"
    network_interface_id = "eni-1234abcd"
}`,
			Issues: []*issue.Issue{},
		},
	}

	for _, tc := range cases {
		var issues = []*issue.Issue{}
		err := TestDetectByCreatorName(
			"CreateAwsRouteNotSpecifiedTargetDetector",
			tc.Src,
			"",
			config.Init(),
			config.Init().NewAwsClient(),
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
