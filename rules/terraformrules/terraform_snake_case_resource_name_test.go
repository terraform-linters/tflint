package terraformrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformSnakeCaseResourceNameRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "dash in name",
			Content: `
resource "aws_eip" "dash-name" {
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformSnakeCaseResourceNameRule(),
					Message: "`dash-name` resource name is not snake_case",
					Range: hcl.Range{
						Filename: "resources.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 31},
					},
				},
			},
		},
		{
			Name: "capital letter in name",
			Content: `
resource "aws_eip" "camelCased" {
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformSnakeCaseResourceNameRule(),
					Message: "`camelCased` resource name is not snake_case",
					Range: hcl.Range{
						Filename: "resources.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 32},
					},
				},
			},
		},
		{
			Name: "valid snake_case name",
			Content: `
resource "aws_eip" "snake_case" {
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformSnakeCaseResourceNameRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resources.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
