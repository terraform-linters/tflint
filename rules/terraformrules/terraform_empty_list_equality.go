package terraformrules

import (
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint/tflint"
)

// TerraformEmptyListEqualityRule checks whether is there a comparison with an empty list
type TerraformEmptyListEqualityRule struct{}

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
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether the list is being compared with static empty list
func (r *TerraformEmptyListEqualityRule) Check(runner *tflint.Runner) error {
	if !runner.TFConfig.Path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	if err := r.checkEmptyList(runner); err != nil {
		return err
	}

	return nil
}

// checkEmptyList visits all blocks that can contain expressions and checks for comparisons with static empty list
func (r *TerraformEmptyListEqualityRule) checkEmptyList(runner *tflint.Runner) error {
	return runner.WalkExpressions(func(expr hcl.Expression) error {
		if conditionalExpr, ok := expr.(*hclsyntax.ConditionalExpr); ok {
			var issuesRangeSet = make(map[hcl.Range]struct{})
			r.searchEmptyList(conditionalExpr.Condition, runner, conditionalExpr.Range(), issuesRangeSet)
			r.emitEmptyListEqualityIssues(issuesRangeSet, runner)
		}
		return nil
	})
}

// searchEmptyList Searches for comparisons with static empty list in the given expression
func (r *TerraformEmptyListEqualityRule) searchEmptyList(expr hcl.Expression, runner *tflint.Runner, exprRange hcl.Range, issuesRangeSet map[hcl.Range]struct{}) {
	if binaryOpExpr, ok := expr.(*hclsyntax.BinaryOpExpr); ok {
		if binaryOpExpr.Op.Type.FriendlyName() == "bool" {
			r.searchEmptyList(binaryOpExpr.RHS, runner, binaryOpExpr.Range(), issuesRangeSet)
			r.searchEmptyList(binaryOpExpr.LHS, runner, binaryOpExpr.Range(), issuesRangeSet)
		}
	} else if binaryOpExpr, ok := expr.(*hclsyntax.BinaryOpExpr); ok {
		r.searchEmptyList(binaryOpExpr, runner, binaryOpExpr.Range(), issuesRangeSet)
	} else if parenthesesExpr, ok := expr.(*hclsyntax.ParenthesesExpr); ok {
		r.searchEmptyList(parenthesesExpr.Expression, runner, parenthesesExpr.Range(), issuesRangeSet)
	} else if tupleConsExpr, ok := expr.(*hclsyntax.TupleConsExpr); ok {
		if len(tupleConsExpr.Exprs) == 0 {
			issuesRangeSet[exprRange] = struct{}{}
		}
	}
}

// emitEmptyListEqualityIssues emits issues for each found comparison with static empty list
func (r *TerraformEmptyListEqualityRule) emitEmptyListEqualityIssues(exprRanges map[hcl.Range]struct{}, runner *tflint.Runner) {
	for exprRange := range exprRanges {
		runner.EmitIssue(
			r,
			"Comparing a collection with an empty list is invalid. To detect an empty collection, check its length.",
			exprRange,
		)
	}
}
