package rules

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
)

// TerraformCommentSyntaxRule checks whether comments use the preferred syntax
type TerraformCommentSyntaxRule struct {
	tflint.DefaultRule
}

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
	return true
}

// Severity returns the rule severity
func (r *TerraformCommentSyntaxRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformCommentSyntaxRule) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check checks whether single line comments is used
func (r *TerraformCommentSyntaxRule) Check(runner tflint.Runner) error {
	path, err := runner.GetModulePath()
	if err != nil {
		return err
	}
	if !path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	files, err := runner.GetFiles()
	if err != nil {
		return err
	}
	for name, file := range files {
		if err := r.checkComments(runner, name, file); err != nil {
			return err
		}
	}

	return nil
}

func (r *TerraformCommentSyntaxRule) checkComments(runner tflint.Runner, filename string, file *hcl.File) error {
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
			if err := runner.EmitIssue(
				r,
				"Single line comments should begin with #",
				token.Range,
			); err != nil {
				return err
			}
		}
	}

	return nil
}
