package terraformrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformSnakeCaseOutputNameRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "dash in name",
			Content: `
output "dash-name" {
	value = aws_alb.main.dns_name
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformSnakeCaseOutputNameRule(),
					Message: "`dash-name` output name is not snake_case",
					Range: hcl.Range{
						Filename: "outputs.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 19},
					},
				},
			},
		},
		{
			Name: "capital letter in name",
			Content: `
output "camelCased" {
	value = aws_alb.main.dns_name
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformSnakeCaseOutputNameRule(),
					Message: "`camelCased` output name is not snake_case",
					Range: hcl.Range{
						Filename: "outputs.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 20},
					},
				},
			},
		},
		{
			Name: "valid snake_case name",
			Content: `
output "snake_case" {
	value = aws_alb.main.dns_name
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformSnakeCaseOutputNameRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"outputs.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
