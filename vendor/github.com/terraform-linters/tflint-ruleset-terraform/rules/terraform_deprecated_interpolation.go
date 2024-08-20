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
// See https://github.com/hashicorp/terraform/blob/2ce03abe480c3f40d04bd0f289762721ea280848/configs/compat_shim.go#L144-L156
func (r *TerraformDeprecatedInterpolationRule) Check(runner tflint.Runner) error {
	path, err := runner.GetModulePath()
	if err != nil {
		return err
	}
	if !path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	diags := runner.WalkExpressions(&terraformDeprecatedInterpolationWalker{
		runner: runner,
		rule:   r,
		// create some capacity so that we can deal with simple expressions
		// without any further allocation during our walk.
		contextStack: make([]terraformDeprecatedInterpolationContext, 0, 16),
	})
	if diags.HasErrors() {
		return diags
	}
	return nil
}

type terraformDeprecatedInterpolationWalker struct {
	runner       tflint.Runner
	rule         *TerraformDeprecatedInterpolationRule
	contextStack []terraformDeprecatedInterpolationContext
}

var _ tflint.ExprWalker = (*terraformDeprecatedInterpolationWalker)(nil)

type terraformDeprecatedInterpolationContext int

const (
	terraformDeprecatedInterpolationContextNormal terraformDeprecatedInterpolationContext = 0
	terraformDeprecatedInterpolationContextObjKey terraformDeprecatedInterpolationContext = 1
)

func (w *terraformDeprecatedInterpolationWalker) Enter(expr hcl.Expression) hcl.Diagnostics {
	var err error

	context := terraformDeprecatedInterpolationContextNormal
	switch expr := expr.(type) {
	case *hclsyntax.ObjectConsKeyExpr:
		context = terraformDeprecatedInterpolationContextObjKey
	case *hclsyntax.TemplateWrapExpr:
		// hclsyntax.TemplateWrapExpr is a special node type used by HCL only
		// for the situation where a template is just a single interpolation,
		// so we don't need to do anything further to distinguish that
		// situation. ("normal" templates are *hclsyntax.TemplateExpr.)

		const message = "Interpolation-only expressions are deprecated in Terraform v0.12.14"
		switch w.currentContext() {
		case terraformDeprecatedInterpolationContextObjKey:
			// This case requires a different autofix strategy is needed
			// to avoid ambiguous attribute keys.
			err = w.runner.EmitIssueWithFix(
				w.rule,
				message,
				expr.Range(),
				func(f tflint.Fixer) error {
					return f.ReplaceText(expr.Range(), "(", f.TextAt(expr.Wrapped.Range()), ")")
				},
			)
		default:
			err = w.runner.EmitIssueWithFix(
				w.rule,
				message,
				expr.Range(),
				func(f tflint.Fixer) error {
					return f.ReplaceText(expr.Range(), f.TextAt(expr.Wrapped.Range()))
				},
			)
		}
	}

	// Note the context of the current node for when we potentially visit
	// child nodes.
	w.contextStack = append(w.contextStack, context)

	if err != nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "failed to call EmitIssueWithFix()",
				Detail:   err.Error(),
			},
		}
	}
	return nil
}

func (w *terraformDeprecatedInterpolationWalker) Exit(expr hcl.Expression) hcl.Diagnostics {
	w.contextStack = w.contextStack[:len(w.contextStack)-1]
	return nil
}

func (w *terraformDeprecatedInterpolationWalker) currentContext() terraformDeprecatedInterpolationContext {
	if len(w.contextStack) == 0 {
		return terraformDeprecatedInterpolationContextNormal
	}
	return w.contextStack[len(w.contextStack)-1]
}
