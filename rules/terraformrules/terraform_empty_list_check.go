package terraformrules

import (
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint/tflint"
)

// TerraformEmptyListCheckRule checks whether is there a comparison with an empty list
type TerraformEmptyListCheckRule struct{}

// NewTerraformCommentSyntaxRule returns a new rule
func NewTerraformEmptyListCheckRule() *TerraformEmptyListCheckRule {
	return &TerraformEmptyListCheckRule{}
}

// Name returns the rule name
func (r *TerraformEmptyListCheckRule) Name() string {
	return "terraform_empty_list_check"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformEmptyListCheckRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *TerraformEmptyListCheckRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformEmptyListCheckRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether the list is being compared with static empty list
func (r *TerraformEmptyListCheckRule) Check(runner *tflint.Runner) error {
	if !runner.TFConfig.Path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	files := make(map[string]*struct{})
	for _, variable := range runner.TFConfig.Module.Variables {
		files[variable.DeclRange.Filename] = nil
	}

	for filename := range files {
		if err := r.checkEmptyList(runner, filename); err != nil {
			return err
		}
	}

	return nil
}

func (r *TerraformEmptyListCheckRule) checkEmptyList(runner *tflint.Runner, filename string) error {
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

func checkEmptyList(tupleConsExpr *hclsyntax.TupleConsExpr, runner *tflint.Runner, r *TerraformEmptyListCheckRule, binaryOpExpr *hclsyntax.BinaryOpExpr) {
	if len(tupleConsExpr.Exprs) == 0 {
		runner.EmitIssue(
			r,
			"List is compared with [] instead of checking if length is 0.",
			binaryOpExpr.Range(),
		)
	}
}
