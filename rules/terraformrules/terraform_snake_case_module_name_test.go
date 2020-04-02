package terraformrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformSnakeCaseModuleNameRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "dash in name",
			Content: `
module "dash-name" {
	source = ""
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformSnakeCaseModuleNameRule(),
					Message: "`dash-name` module name is not snake_case",
					Range: hcl.Range{
						Filename: "resources.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 19},
					},
				},
			},
		},
		{
			Name: "capital letter in name",
			Content: `
module "camelCased" {
	source = ""
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformSnakeCaseModuleNameRule(),
					Message: "`camelCased` module name is not snake_case",
					Range: hcl.Range{
						Filename: "resources.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 20},
					},
				},
			},
		},
		{
			Name: "valid snake_case name",
			Content: `
module "snake_case" {
	source = ""
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformSnakeCaseModuleNameRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resources.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
