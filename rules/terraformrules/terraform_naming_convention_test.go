package terraformrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformNamingConventionRule(t *testing.T) {
	test_data_defaultConfig(t)
}

func test_data_defaultConfig(t *testing.T) {
	rule := &TerraformNamingConventionRule{}

	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "data: default config - Invalid snake_case with dash",
			Content: `
data "aws_eip" "dash-name" {
}`,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: "data name `dash-name` must match the following format: snake_case",
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 27},
					},
				},
			},
		},
		{
			Name: "data: default config - Invalid snake_case with camelCase",
			Content: `
data "aws_eip" "camelCased" {
}`,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: "data name `camelCased` must match the following format: snake_case",
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 28},
					},
				},
			},
		},
		{
			Name: "data: default config - Invalid snake_case with Mixed_Snake_Case",
			Content: `
data "aws_eip" "Foo_Bar" {
}`,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: "data name `Foo_Bar` must match the following format: snake_case",
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 25},
					},
				},
			},
		},
		{
			Name: "data: default config - Valid snake_case",
			Content: `
data "aws_eip" "foo_bar" {
}`,
			Expected: tflint.Issues{},
		},
	}

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"tests.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
