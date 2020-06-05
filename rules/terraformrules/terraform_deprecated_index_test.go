package terraformrules

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformDeprecatedIndexRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "deprecated dot index style",
			Content: `
locals {
  list = ["a"]
  value = list.0
}
`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformDeprecatedIndexRule(),
					Message: "List items should be accessed using square brackets",
					Range: hcl.Range{
						Filename: "config.tf",
						Start: hcl.Pos{
							Line:   4,
							Column: 11,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 17,
						},
					},
				},
			},
		},
		{
			Name: "attribute access",
			Content: `
locals {
  map = {a = "b"}
  value = map.a
}
`,
			Expected: tflint.Issues{},
		},
		{
			Name: "fractional number",
			Content: `
locals {
  value = 1.5
}
`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformDeprecatedIndexRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"config.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
