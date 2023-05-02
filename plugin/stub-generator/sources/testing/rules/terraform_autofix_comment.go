package rules

import (
	"runtime"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformAutofixComment checks whether ...
type TerraformAutofixComment struct {
	tflint.DefaultRule
}

// NewTerraformAutofixCommentRule returns a new rule
func NewTerraformAutofixCommentRule() *TerraformAutofixComment {
	return &TerraformAutofixComment{}
}

// Name returns the rule name
func (r *TerraformAutofixComment) Name() string {
	return "terraform_autofix_comment"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformAutofixComment) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *TerraformAutofixComment) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *TerraformAutofixComment) Link() string {
	return ""
}

// Check checks whether ...
func (r *TerraformAutofixComment) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for name, file := range files {
		if strings.HasSuffix(name, ".tf.json") {
			continue
		}

		tokens, diags := hclsyntax.LexConfig(file.Bytes, name, hcl.InitialPos)
		if diags.HasErrors() {
			return diags
		}

		for _, token := range tokens {
			if token.Type != hclsyntax.TokenComment {
				continue
			}

			if string(token.Bytes) == "// autofixed"+newLine() {
				if err := runner.EmitIssueWithFix(
					r,
					`Use "# autofixed" instead of "// autofixed"`,
					token.Range,
					func(f tflint.Fixer) error {
						return f.ReplaceText(
							f.RangeTo("// autofixed", name, token.Range.Start),
							"# autofixed",
						)
					},
				); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func newLine() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}
