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
func (r *TerraformDeprecatedInterpolationRule) Severity() string {
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
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	for _, resource := range runner.TFConfig.Module.ManagedResources {
		r.checkForDeprecatedInterpolationsInBody(runner, resource.Config)
	}
	for _, provider := range runner.TFConfig.Module.ProviderConfigs {
		r.checkForDeprecatedInterpolationsInBody(runner, provider.Config)
	}

	return nil
}

func (r *TerraformDeprecatedInterpolationRule) checkForDeprecatedInterpolationsInBody(runner *tflint.Runner, body hcl.Body) {
	nativeBody, ok := body.(*hclsyntax.Body)
	if !ok {
		return
	}

	for _, attr := range nativeBody.Attributes {
		r.checkForDeprecatedInterpolationsInExpr(runner, attr.Expr)
	}

	for _, block := range nativeBody.Blocks {
		r.checkForDeprecatedInterpolationsInBody(runner, block.Body)
	}

	return
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

	return
}
