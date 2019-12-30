package terraformrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformDashInOutputNameRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "dash in output name",
			Content: `
output "dash-name" {
	value = aws_alb.main.dns_name
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformDashInOutputNameRule(),
					Message: "`dash-name` output name has a dash",
					Range: hcl.Range{
						Filename: "outputs.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 19},
					},
				},
			},
		},
		{
			Name: "no dash in output name",
			Content: `
output "no_dash_name" {
	value = aws_alb.main.dns_name
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformDashInOutputNameRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"outputs.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
