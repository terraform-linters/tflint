package awsrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_AwsDBInstanceDefaultParameterGroup(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "default.mysql5.6 is default parameter group",
			Content: `
resource "aws_db_instance" "db" {
    parameter_group_name = "default.mysql5.6"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsDBInstanceDefaultParameterGroupRule(),
					Message: "\"default.mysql5.6\" is default parameter group. You cannot edit it.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 28},
						End:      hcl.Pos{Line: 3, Column: 46},
					},
				},
			},
		},
		{
			Name: "application5.6 is not default parameter group",
			Content: `
resource "aws_db_instance" "db" {
    parameter_group_name = "application5.6"
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewAwsDBInstanceDefaultParameterGroupRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resource.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
