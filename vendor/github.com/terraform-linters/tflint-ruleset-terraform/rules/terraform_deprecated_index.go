package rules

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
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
	return true
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

	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	diags := runner.WalkExpressions(tflint.ExprWalkFunc(func(e hcl.Expression) hcl.Diagnostics {
		filename := e.Range().Filename
		file := files[filename]

		if json.IsJSONExpression(e) {
			r.checkJSONExpression(runner, e, file.Bytes)
			return nil
		}

		switch expr := e.(type) {
		case *hclsyntax.ScopeTraversalExpr:
			r.checkLegacyTraversalIndex(runner, expr.Traversal, file.Bytes)
		case *hclsyntax.RelativeTraversalExpr:
			r.checkLegacyTraversalIndex(runner, expr.Traversal, file.Bytes)
		case *hclsyntax.SplatExpr:
			if strings.HasPrefix(string(expr.MarkerRange.SliceBytes(file.Bytes)), ".") {
				if err := runner.EmitIssueWithFix(
					r,
					"List items should be accessed using square brackets",
					expr.MarkerRange,
					func(f tflint.Fixer) error {
						return f.ReplaceText(expr.MarkerRange, "[*]")
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
			}
		}
		return nil
	}))
	if diags.HasErrors() {
		return diags
	}

	return nil
}

func (r *TerraformDeprecatedIndexRule) checkLegacyTraversalIndex(runner tflint.Runner, traversal hcl.Traversal, file []byte) hcl.Diagnostics {
	for _, t := range traversal {
		if tn, ok := t.(hcl.TraverseIndex); ok {
			if strings.HasPrefix(string(t.SourceRange().SliceBytes(file)), ".") {
				if err := runner.EmitIssueWithFix(
					r,
					"List items should be accessed using square brackets",
					t.SourceRange(),
					func(f tflint.Fixer) error {
						return f.ReplaceText(t.SourceRange(), "[", f.ValueText(tn.Key), "]")
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
			}
		}
	}
	return nil
}

func (r *TerraformDeprecatedIndexRule) checkJSONExpression(runner tflint.Runner, e hcl.Expression, file []byte) hcl.Diagnostics {
	var diags hcl.Diagnostics

	for _, v := range e.Variables() {
		diags = append(diags, r.checkLegacyTraversalIndex(runner, v, file)...)
	}

	return diags
}
