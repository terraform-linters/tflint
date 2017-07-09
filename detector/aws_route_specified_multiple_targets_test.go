package detector

import (
	"reflect"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
)

func TestDetectAwsRouteSpecifiedMultipleTargets(t *testing.T) {
	cases := []struct {
		Name   string
		Src    string
		Issues []*issue.Issue
	}{
		{
			Name: "multiple route targets are specified",
			Src: `
resource "aws_route" "foo" {
    route_table_id = "rtb-1234abcd"
    gateway_id = "igw-1234abcd"
    egress_only_gateway_id = "eigw-1234abcd"
}`,
			Issues: []*issue.Issue{
				{
					Detector: "aws_route_specified_multiple_targets",
					Type:     "ERROR",
					Message:  "More than one routing target specified. It must be one.",
					Line:     2,
					File:     "test.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_route_specified_multiple_targets.md",
				},
			},
		},
		{
			Name: "single a route target is specified",
			Src: `
resource "aws_route" "foo" {
    route_table_id = "rtb-1234abcd"
    gateway_id = "igw-1234abcd"
}`,
			Issues: []*issue.Issue{},
		},
	}

	for _, tc := range cases {
		var issues = []*issue.Issue{}
		err := TestDetectByCreatorName(
			"CreateAwsRouteSpecifiedMultipleTargetsDetector",
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
