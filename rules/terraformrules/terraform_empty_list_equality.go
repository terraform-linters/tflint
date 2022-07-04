package terraformrules

import (
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint/tflint"
)

// TerraformEmptyListEqualityRule checks whether is there a comparison with an empty list
type TerraformEmptyListEqualityRule struct{}
type void struct{}

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

func (r *TerraformEmptyListEqualityRule) checkEmptyList(runner *tflint.Runner) error {
	return runner.WalkExpressions(func(expr hcl.Expression) error {
		if conditionalExpr, ok := expr.(*hclsyntax.ConditionalExpr); ok {
			var issuesRangeSet = make(map[hcl.Range]void)
			searchEmptyList(conditionalExpr.Condition, runner, r, conditionalExpr.Range(), issuesRangeSet)
			emitEmptyListEqualityIssues(issuesRangeSet, runner, r)
		}
		return nil
	})
}

func searchEmptyList(expr hcl.Expression, runner *tflint.Runner, rule *TerraformEmptyListEqualityRule, exprRange hcl.Range, issuesRangeSet map[hcl.Range]void) {
	if binaryOpExpr, ok := expr.(*hclsyntax.BinaryOpExpr); ok {
		if binaryOpExpr.Op.Type.FriendlyName() == "bool" {
			searchEmptyList(binaryOpExpr.RHS, runner, rule, binaryOpExpr.Range(), issuesRangeSet)
			searchEmptyList(binaryOpExpr.LHS, runner, rule, binaryOpExpr.Range(), issuesRangeSet)
		}
	} else if binaryOpExpr, ok := expr.(*hclsyntax.BinaryOpExpr); ok {
		searchEmptyList(binaryOpExpr, runner, rule, binaryOpExpr.Range(), issuesRangeSet)
	} else if parenthesesExpr, ok := expr.(*hclsyntax.ParenthesesExpr); ok {
		searchEmptyList(parenthesesExpr.Expression, runner, rule, parenthesesExpr.Range(), issuesRangeSet)
	} else if tupleConsExpr, ok := expr.(*hclsyntax.TupleConsExpr); ok {
		if len(tupleConsExpr.Exprs) == 0 {
			issuesRangeSet[exprRange] = void{}
		}
	}
}

func emitEmptyListEqualityIssues(exprRanges map[hcl.Range]void, runner *tflint.Runner, rule *TerraformEmptyListEqualityRule) {
	for exprRange := range exprRanges {
		runner.EmitIssue(
			rule,
			"Comparing a collection with an empty list is invalid. To detect an empty collection, check its length.",
			exprRange,
		)
	}
}
