package terraformrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformSnakeCaseLocalNameRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "dash in name",
			Content: `
locals {
	dash-name = "invalid"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformSnakeCaseLocalNameRule(),
					Message: "`dash-name` local name is not snake_case",
					Range: hcl.Range{
						Filename: "locals.tf",
						Start:    hcl.Pos{Line: 3, Column: 2},
						End:      hcl.Pos{Line: 3, Column: 23},
					},
				},
			},
		},
		{
			Name: "capital letter in name",
			Content: `
locals {
	camelCased = "invalid"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformSnakeCaseLocalNameRule(),
					Message: "`camelCased` local name is not snake_case",
					Range: hcl.Range{
						Filename: "locals.tf",
						Start:    hcl.Pos{Line: 3, Column: 2},
						End:      hcl.Pos{Line: 3, Column: 24},
					},
				},
			},
		},
		{
			Name: "valid snake_case name",
			Content: `
locals {
	snake_case = "valid"
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformSnakeCaseLocalNameRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"locals.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
