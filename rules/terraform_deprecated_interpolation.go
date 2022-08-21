package rules

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
)

// TerraformDeprecatedInterpolationRule warns of deprecated interpolation in Terraform v0.11 or earlier.
type TerraformDeprecatedInterpolationRule struct {
	tflint.DefaultRule
}

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
	return project.ReferenceLink(r.Name())
}

// Check emits issues on the deprecated interpolation syntax.
// This logic is equivalent to the warning logic implemented in Terraform.
// See https://github.com/hashicorp/terraform/pull/23348
func (r *TerraformDeprecatedInterpolationRule) Check(runner tflint.Runner) error {
	path, err := runner.GetModulePath()
	if err != nil {
		return err
	}
	if !path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	diags := runner.WalkExpressions(tflint.ExprWalkFunc(func(expr hcl.Expression) hcl.Diagnostics {
		return r.checkForDeprecatedInterpolationsInExpr(runner, expr)
	}))
	if diags.HasErrors() {
		return diags
	}
	return nil
}

func (r *TerraformDeprecatedInterpolationRule) checkForDeprecatedInterpolationsInExpr(runner tflint.Runner, expr hcl.Expression) hcl.Diagnostics {
	if _, ok := expr.(*hclsyntax.TemplateWrapExpr); !ok {
		return nil
	}

	err := runner.EmitIssue(
		r,
		"Interpolation-only expressions are deprecated in Terraform v0.12.14",
		expr.Range(),
	)
	if err != nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "failed to call EmitIssue()",
				Detail:   err.Error(),
			},
		}
	}
	return nil
}
