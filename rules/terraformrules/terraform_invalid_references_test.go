package terraformrules

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/wata727/tflint/tflint"
)

func Test_TerraformInvalidReferencesRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "no braces",
			Content: `
resource "null_resource" "generate_inventory" {
  triggers = {
    template_rendered = "$data.template_file.inventory.rendered"
  }
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformInvalidReferencesRule(),
					Message: "reference to ${data.template_file} is missing braces",
					Range: hcl.Range{
						Filename: "template_references.tf",
						Start:    hcl.Pos{Line: 4, Column: 26},
						End:      hcl.Pos{Line: 4, Column: 64},
					},
				},
			},
		},
		{
			Name: "with braces",
			Content: `
resource "null_resource" "generate_inventory" {
  triggers = {
    template_rendered = "${data.template_file.inventory.rendered}"
  }
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformInvalidReferencesRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"template_references.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
