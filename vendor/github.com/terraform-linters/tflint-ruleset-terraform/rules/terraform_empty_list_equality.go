package rules

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
)

// TerraformEmptyListEqualityRule checks whether is there a comparison with an empty list
type TerraformEmptyListEqualityRule struct {
	tflint.DefaultRule
}

// NewTerraformCommentSyntaxRule returns a new rule
func NewTerraformEmptyListEqualityRule() *TerraformEmptyListEqualityRule {
	return &TerraformEmptyListEqualityRule{}
}

// Name returns the rule name
func (r *TerraformEmptyListEqualityRule) Name() string {
	return "terraform_empty_list_equality"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformEmptyListEqualityRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *TerraformEmptyListEqualityRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformEmptyListEqualityRule) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check checks whether the list is being compared with static empty list
func (r *TerraformEmptyListEqualityRule) Check(runner tflint.Runner) error {
	path, err := runner.GetModulePath()
	if err != nil {
		return err
	}
	if !path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	if diags := r.checkEmptyList(runner); diags.HasErrors() {
		return diags
	}

	return nil
}

// checkEmptyList visits all blocks that can contain expressions and checks for comparisons with static empty list
func (r *TerraformEmptyListEqualityRule) checkEmptyList(runner tflint.Runner) hcl.Diagnostics {
	return runner.WalkExpressions(tflint.ExprWalkFunc(func(expr hcl.Expression) hcl.Diagnostics {
		if binaryOpExpr, ok := expr.(*hclsyntax.BinaryOpExpr); ok && (binaryOpExpr.Op == hclsyntax.OpEqual || binaryOpExpr.Op == hclsyntax.OpNotEqual) {
			if tupleConsExpr, ok := binaryOpExpr.LHS.(*hclsyntax.TupleConsExpr); ok && len(tupleConsExpr.Exprs) == 0 {
				if err := r.emitIssue(binaryOpExpr, binaryOpExpr.RHS, runner); err != nil {
					return hcl.Diagnostics{
						{
							Severity: hcl.DiagError,
							Summary:  "failed to call EmitIssueWithFix()",
							Detail:   err.Error(),
						},
					}
				}
			} else if tupleConsExpr, ok := binaryOpExpr.RHS.(*hclsyntax.TupleConsExpr); ok && len(tupleConsExpr.Exprs) == 0 {
				if err := r.emitIssue(binaryOpExpr, binaryOpExpr.LHS, runner); err != nil {
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
}

// emitIssue emits issue for comparison with static empty list
func (r *TerraformEmptyListEqualityRule) emitIssue(binaryOpExpr *hclsyntax.BinaryOpExpr, hs hcl.Expression, runner tflint.Runner) error {
	var opStr string
	if binaryOpExpr.Op == hclsyntax.OpEqual {
		opStr = "=="
	} else {
		opStr = "!="
	}

	return runner.EmitIssueWithFix(
		r,
		"Comparing a collection with an empty list is invalid. To detect an empty collection, check its length.",
		binaryOpExpr.Range(),
		func(f tflint.Fixer) error {
			return f.ReplaceText(binaryOpExpr.Range(), "length(", f.TextAt(hs.Range()), ") ", opStr, " 0")
		},
	)
}
