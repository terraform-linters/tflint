package terraformrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformDashInModuleNameRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "dash in resource name",
			Content: `
module "some-module" {
	source = ""
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformDashInModuleNameRule(),
					Message: "`some-module` module name has a dash",
					Range: hcl.Range{
						Filename: "resources.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 21},
					},
				},
			},
		},
		{
			Name: "no dash in resource name",
			Content: `
module "some_module" {
	source = ""
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformDashInModuleNameRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resources.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
