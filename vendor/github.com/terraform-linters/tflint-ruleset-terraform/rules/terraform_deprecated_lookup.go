package rules

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
)

// TerraformDeprecatedLookupRule warns about usage of the legacy dot syntax for indexes (foo.0)
type TerraformDeprecatedLookupRule struct {
	tflint.DefaultRule
}

// NewTerraformDeprecatedIndexRule return a new rule
func NewTerraformDeprecatedLookupRule() *TerraformDeprecatedLookupRule {
	return &TerraformDeprecatedLookupRule{}
}

// Name returns the rule name
func (r *TerraformDeprecatedLookupRule) Name() string {
	return "terraform_deprecated_lookup"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformDeprecatedLookupRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *TerraformDeprecatedLookupRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformDeprecatedLookupRule) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check walks all expressions and emit issues if deprecated index syntax is found
func (r *TerraformDeprecatedLookupRule) Check(runner tflint.Runner) error {
	path, err := runner.GetModulePath()
	if err != nil {
		return err
	}
	if !path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	diags := runner.WalkExpressions(tflint.ExprWalkFunc(func(e hcl.Expression) hcl.Diagnostics {
		call, ok := e.(*hclsyntax.FunctionCallExpr)
		if ok && call.Name == "lookup" && len(call.Args) == 2 {
			if err := runner.EmitIssueWithFix(
				r,
				"Lookup with 2 arguments is deprecated",
				call.Range(),
				func(f tflint.Fixer) error {
					return f.ReplaceText(call.Range(), f.TextAt(call.Args[0].Range()), "[", f.TextAt(call.Args[1].Range()), "]")
				},
			); err != nil {
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
		return nil
	}))
	if diags.HasErrors() {
		return diags
	}

	return nil
}
