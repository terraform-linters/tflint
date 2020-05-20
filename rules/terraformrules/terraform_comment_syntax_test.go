package terraformrules

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformCommentSyntaxRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		JSON     bool
		Expected tflint.Issues
	}{
		{
			Name:     "hash comment",
			Content:  `# foo`,
			Expected: tflint.Issues{},
		},
		{
			Name: "multi-line comment",
			Content: `
/*
	This comment spans multiple lines
*/			
`,
			Expected: tflint.Issues{},
		},
		{
			Name:    "double-slash comment",
			Content: `// foo`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformCommentSyntaxRule(),
					Message: "Single line comments should begin with #",
					Range: hcl.Range{
						Filename: "variables.tf",
						Start: hcl.Pos{
							Line:   1,
							Column: 1,
						},
						End: hcl.Pos{
							Line:   1,
							Column: 7,
						},
					},
				},
			},
		},
		{
			Name: "end-of-line hash comment",
			Content: `
variable "foo" {
	type = string # a string
}
`,
			Expected: tflint.Issues{},
		},
		{
			Name:     "JSON",
			Content:  `{"variable": {"foo": {"type": "string"}}}`,
			JSON:     true,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformCommentSyntaxRule()

	for _, tc := range cases {
		filename := "variables.tf"
		if tc.JSON {
			filename += ".json"
		}

		runner := tflint.TestRunner(t, map[string]string{filename: tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
