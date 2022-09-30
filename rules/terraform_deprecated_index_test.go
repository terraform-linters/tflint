package rules

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_TerraformDeprecatedIndexRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name: "deprecated dot index style",
			Content: `
locals {
  list = ["a"]
  value = list.0
}
`,
			Expected: helper.Issues{
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
			Name: "deprecated dot splat index style",
			Content: `
locals {
  maplist = [{a = "b"}]
  values = maplist.*.a
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformDeprecatedIndexRule(),
					Message: "List items should be accessed using square brackets",
					Range: hcl.Range{
						Filename: "config.tf",
						Start: hcl.Pos{
							Line:   4,
							Column: 12,
						},
						End: hcl.Pos{
							Line:   4,
							Column: 23,
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
			Expected: helper.Issues{},
		},
		{
			Name: "fractional number",
			Content: `
locals {
  value = 1.5
}
`,
			Expected: helper.Issues{},
		},
	}

	rule := NewTerraformDeprecatedIndexRule()

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"config.tf": tc.Content})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}
