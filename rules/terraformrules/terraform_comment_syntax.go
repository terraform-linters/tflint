package terraformrules

import (
	"log"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint/tflint"
)

// TerraformCommentSyntaxRule checks whether comments use the preferred syntax
type TerraformCommentSyntaxRule struct{}

// NewTerraformCommentSyntaxRule returns a new rule
func NewTerraformCommentSyntaxRule() *TerraformCommentSyntaxRule {
	return &TerraformCommentSyntaxRule{}
}

// Name returns the rule name
func (r *TerraformCommentSyntaxRule) Name() string {
	return "terraform_comment_syntax"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformCommentSyntaxRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformCommentSyntaxRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformCommentSyntaxRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether single line comments is used
func (r *TerraformCommentSyntaxRule) Check(runner *tflint.Runner) error {
	if !runner.TFConfig.Path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	for name, file := range runner.Files() {
		if err := r.checkComments(runner, name, file); err != nil {
			return err
		}
	}

	return nil
}

func (r *TerraformCommentSyntaxRule) checkComments(runner *tflint.Runner, filename string, file *hcl.File) error {
	if strings.HasSuffix(filename, ".json") {
		return nil
	}

	tokens, diags := hclsyntax.LexConfig(file.Bytes, filename, hcl.InitialPos)
	if diags.HasErrors() {
		return diags
	}

	for _, token := range tokens {
		if token.Type != hclsyntax.TokenComment {
			continue
		}

		if strings.HasPrefix(string(token.Bytes), "//") {
			runner.EmitIssue(
				r,
				"Single line comments should begin with #",
				token.Range,
			)
		}
	}

	return nil
}
