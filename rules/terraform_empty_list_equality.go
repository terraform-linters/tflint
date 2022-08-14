package rules

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
	"github.com/zclconf/go-cty/cty"
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

	if err := r.checkEmptyList(runner); err != nil {
		return err
	}

	return nil
}

// checkEmptyList visits all blocks that can contain expressions and checks for comparisons with static empty list
func (r *TerraformEmptyListEqualityRule) checkEmptyList(runner tflint.Runner) error {
	return WalkExpressions(runner, func(expr hcl.Expression) error {
		if binaryOpExpr, ok := expr.(*hclsyntax.BinaryOpExpr); ok && binaryOpExpr.Op.Type == cty.Bool {
			if tupleConsExpr, ok := binaryOpExpr.LHS.(*hclsyntax.TupleConsExpr); ok && len(tupleConsExpr.Exprs) == 0 {
				if err := r.emitIssue(binaryOpExpr.Range(), runner); err != nil {
					return err
				}
			} else if tupleConsExpr, ok := binaryOpExpr.RHS.(*hclsyntax.TupleConsExpr); ok && len(tupleConsExpr.Exprs) == 0 {
				if err := r.emitIssue(binaryOpExpr.Range(), runner); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// emitIssue emits issue for comparison with static empty list
func (r *TerraformEmptyListEqualityRule) emitIssue(exprRange hcl.Range, runner tflint.Runner) error {
	return runner.EmitIssue(
		r,
		"Comparing a collection with an empty list is invalid. To detect an empty collection, check its length.",
		exprRange,
	)
}
