package terraformrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformNamingThisRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		JSON     bool
		Expected tflint.Issues
	}{
		{
			Name: "single resource with wrong name",
			Content: `
resource "test_type" "wrong_name" {}
`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformNamingThisRule(),
					Message: "Found only one resource of type `test_type`, therefore the resource name should be `this` but was `wrong_name`",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 34},
					},
				},
			},
		},
		{
			Name: "single resource with correct name",
			Content: `
resource "test_type" "this" {}
`,
			Expected: tflint.Issues{},
		},
		{
			Name: "multiple resources of same type with correct name",
			Content: `
resource "test_type" "a" {}

resource "test_type" "b" {}
`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformNamingThisRule()

	for _, tc := range cases {
		filename := "main.tf"
		if tc.JSON {
			filename += ".json"
		}

		t.Run(tc.Name, func(t *testing.T) {
			runner := tflint.TestRunner(t, map[string]string{filename: tc.Content})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			tflint.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}
