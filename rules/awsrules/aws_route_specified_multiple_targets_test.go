package awsrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/wata727/tflint/tflint"
)

func Test_AwsRouteSpecifiedMultipleTargets(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "multiple route targets are specified",
			Content: `
resource "aws_route" "foo" {
    route_table_id = "rtb-1234abcd"
    gateway_id = "igw-1234abcd"
    egress_only_gateway_id = "eigw-1234abcd"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsRouteSpecifiedMultipleTargetsRule(),
					Message: "More than one routing target specified. It must be one.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 27},
					},
				},
			},
		},
		{
			Name: "single a route target is specified",
			Content: `
resource "aws_route" "foo" {
    route_table_id = "rtb-1234abcd"
    gateway_id = "igw-1234abcd"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "multiple targes found, but the second one is null",
			Content: `
variable "egress_only_gateway_id" {
    type    = string
	default = null
}

resource "aws_route" "foo" {
    route_table_id = "rtb-1234abcd"
    gateway_id = "igw-1234abcd"
    egress_only_gateway_id = var.egress_only_gateway_id
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewAwsRouteSpecifiedMultipleTargetsRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resource.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
