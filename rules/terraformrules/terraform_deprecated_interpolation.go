package terraformrules

import (
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint/tflint"
)

// TerraformDeprecatedInterpolationRule warns of deprecated interpolation in Terraform v0.11 or earlier.
type TerraformDeprecatedInterpolationRule struct{}

// NewTerraformDeprecatedInterpolationRule return a new rule
func NewTerraformDeprecatedInterpolationRule() *TerraformDeprecatedInterpolationRule {
	return &TerraformDeprecatedInterpolationRule{}
}

// Name returns the rule name
func (r *TerraformDeprecatedInterpolationRule) Name() string {
	return "terraform_deprecated_interpolation"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformDeprecatedInterpolationRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *TerraformDeprecatedInterpolationRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformDeprecatedInterpolationRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check emits issues on the deprecated interpolation syntax.
// This logic is equivalent to the warning logic implemented in Terraform.
// See https://github.com/hashicorp/terraform/pull/23348
func (r *TerraformDeprecatedInterpolationRule) Check(runner *tflint.Runner) error {
	if !runner.TFConfig.Path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkExpressions(func(expr hcl.Expression) error {
		r.checkForDeprecatedInterpolationsInExpr(runner, expr)
		return nil
	})
}

func (r *TerraformDeprecatedInterpolationRule) checkForDeprecatedInterpolationsInExpr(runner *tflint.Runner, expr hcl.Expression) {
	if _, ok := expr.(*hclsyntax.TemplateWrapExpr); !ok {
		return
	}

	runner.EmitIssue(
		r,
		"Interpolation-only expressions are deprecated in Terraform v0.12.14",
		expr.Range(),
	)
}
