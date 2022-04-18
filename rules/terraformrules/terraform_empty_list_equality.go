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

func (r *TerraformEmptyListEqualityRule) checkEmptyList(runner *tflint.Runner) error {
	return runner.WalkExpressions(func(expr hcl.Expression) error {
		if conditionalExpr, ok := expr.(*hclsyntax.ConditionalExpr); ok {
			if binaryOpExpr, ok := conditionalExpr.Condition.(*hclsyntax.BinaryOpExpr); ok {
				if binaryOpExpr.Op.Type.FriendlyName() == "bool" {
					if right, ok := binaryOpExpr.RHS.(*hclsyntax.TupleConsExpr); ok {
						checkEmptyList(right, runner, r, binaryOpExpr)
					}
					if left, ok := binaryOpExpr.LHS.(*hclsyntax.TupleConsExpr); ok {
						checkEmptyList(left, runner, r, binaryOpExpr)
					}
				}
			}
		}
		return nil
	})
}

func checkEmptyList(tupleConsExpr *hclsyntax.TupleConsExpr, runner *tflint.Runner, r *TerraformEmptyListEqualityRule, binaryOpExpr *hclsyntax.BinaryOpExpr) {
	if len(tupleConsExpr.Exprs) == 0 {
		runner.EmitIssue(
			r,
			"Comparing a collection with an empty list is invalid. To detect an empty collection, check its length.",
			binaryOpExpr.Range(),
		)
	}
}
