package terraformrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformSnakeCaseVariableNameRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "dash in name",
			Content: `
variable "dash-name" {
	description = "Invalid name"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformSnakeCaseVariableNameRule(),
					Message: "`dash-name` variable name is not snake_case",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 21},
					},
				},
			},
		},
		{
			Name: "capital letter in name",
			Content: `
variable "camelCased" {
	description = "Invalid name"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformSnakeCaseVariableNameRule(),
					Message: "`camelCased` variable name is not snake_case",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 22},
					},
				},
			},
		},
		{
			Name: "valid snake_case name",
			Content: `
variable "snake_case" {
	description = "Valid name"
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformSnakeCaseVariableNameRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"variables.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
