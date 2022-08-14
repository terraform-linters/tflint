package rules

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
)

// TerraformDeprecatedIndexRule warns about usage of the legacy dot syntax for indexes (foo.0)
type TerraformDeprecatedIndexRule struct {
	tflint.DefaultRule
}

// NewTerraformDeprecatedIndexRule return a new rule
func NewTerraformDeprecatedIndexRule() *TerraformDeprecatedIndexRule {
	return &TerraformDeprecatedIndexRule{}
}

// Name returns the rule name
func (r *TerraformDeprecatedIndexRule) Name() string {
	return "terraform_deprecated_index"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformDeprecatedIndexRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformDeprecatedIndexRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformDeprecatedIndexRule) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check walks all expressions and emit issues if deprecated index syntax is found
func (r *TerraformDeprecatedIndexRule) Check(runner tflint.Runner) error {
	path, err := runner.GetModulePath()
	if err != nil {
		return err
	}
	if !path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	return WalkExpressions(runner, func(expr hcl.Expression) error {
		for _, variable := range expr.Variables() {
			for _, traversal := range variable.SimpleSplit().Rel {
				if traversal, ok := traversal.(hcl.TraverseIndex); ok {
					filename := traversal.SrcRange.Filename
					file, err := runner.GetFile(filename)
					if err != nil {
						return err
					}
					bytes := traversal.SrcRange.SliceBytes(file.Bytes)

					tokens, diags := hclsyntax.LexExpression(bytes, filename, traversal.SrcRange.Start)
					if diags.HasErrors() {
						return diags
					}

					if tokens[0].Type == hclsyntax.TokenDot {
						if err := runner.EmitIssue(
							r,
							"List items should be accessed using square brackets",
							expr.Range(),
						); err != nil {
							return err
						}
					}
				}
			}
		}

		return nil
	})
}
