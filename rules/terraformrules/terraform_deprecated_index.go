package terraformrules

import (
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint/tflint"
)

// TerraformDeprecatedIndexRule warns about usage of the legacy dot syntax for indexes (foo.0)
type TerraformDeprecatedIndexRule struct{}

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
	return tflint.ReferenceLink(r.Name())
}

// Check walks all expressions and emit issues if deprecated index syntax is found
func (r *TerraformDeprecatedIndexRule) Check(runner *tflint.Runner) error {
	if !runner.TFConfig.Path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkExpressions(func(expr hcl.Expression) error {
		for _, variable := range expr.Variables() {
			for _, traversal := range variable.SimpleSplit().Rel {
				if traversal, ok := traversal.(hcl.TraverseIndex); ok {
					filename := traversal.SrcRange.Filename
					bytes := traversal.SrcRange.SliceBytes(runner.File(filename).Bytes)

					tokens, diags := hclsyntax.LexExpression(bytes, filename, traversal.SrcRange.Start)
					if diags.HasErrors() {
						return diags
					}

					if tokens[0].Type == hclsyntax.TokenDot {
						runner.EmitIssue(
							r,
							"List items should be accessed using square brackets",
							expr.Range(),
						)
					}
				}
			}
		}

		return nil
	})
}
