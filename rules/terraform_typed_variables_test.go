package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_TerraformTypedVariablesRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		JSON     bool
		Expected helper.Issues
	}{
		{
			Name: "no type",
			Content: `
variable "no_type" {
  default = "default"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformTypedVariablesRule(),
					Message: "`no_type` variable has no type",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 19},
					},
				},
			},
		},
		{
			Name: "complex type",
			Content: `
variable "no_type2" {
  type = list(object({
    internal = number
    external = number
    protocol = string
  }))
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "with type",
			Content: `
variable "with_type" {
  type = string
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "type any",
			Content: `
variable "any" {
  type = any
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "json with type",
			JSON: true,
			Content: `
{
	"variable": {
		"with_type": {
			"type": "string"
		}
	}
}
`,
			Expected: helper.Issues{},
		},
		{
			Name: "json no type",
			JSON: true,
			Content: `
{
	"variable": {
		"no_type": {
			"default": "foo"
		}
	}
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformTypedVariablesRule(),
					Message: "`no_type` variable has no type",
					Range: hcl.Range{
						Filename: "variables.tf.json",
						Start:    hcl.Pos{Line: 4, Column: 16},
						End:      hcl.Pos{Line: 4, Column: 17},
					},
				},
			},
		},
	}

	rule := NewTerraformTypedVariablesRule()

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			filename := "variables.tf"
			if tc.JSON {
				filename += ".json"
			}

			runner := helper.TestRunner(t, map[string]string{filename: tc.Content})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}
