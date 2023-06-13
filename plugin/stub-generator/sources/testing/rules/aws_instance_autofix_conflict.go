package rules

import (
	"fmt"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// AwsInstanceAutofixConflict checks whether ...
type AwsInstanceAutofixConflict struct {
	tflint.DefaultRule
}

// NewAwsInstanceAutofixConflictRule returns a new rule
func NewAwsInstanceAutofixConflictRule() *AwsInstanceAutofixConflict {
	return &AwsInstanceAutofixConflict{}
}

// Name returns the rule name
func (r *AwsInstanceAutofixConflict) Name() string {
	return "aws_instance_autofix_conflict"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsInstanceAutofixConflict) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsInstanceAutofixConflict) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsInstanceAutofixConflict) Link() string {
	return ""
}

// Check checks whether ...
func (r *AwsInstanceAutofixConflict) Check(runner tflint.Runner) error {
	resources, err := runner.GetResourceContent("aws_instance", &hclext.BodySchema{
		Attributes: []hclext.AttributeSchema{{Name: "instance_type"}},
	}, nil)
	if err != nil {
		return err
	}

	for _, resource := range resources.Blocks {
		attribute, exists := resource.Body.Attributes["instance_type"]
		if !exists {
			continue
		}

		err := runner.EvaluateExpr(attribute.Expr, func(instanceType string) error {
			if instanceType != "[AUTO_FIXED]" {
				return nil
			}

			return runner.EmitIssueWithFix(
				r,
				fmt.Sprintf("instance type is %s", instanceType),
				attribute.Expr.Range(),
				func(f tflint.Fixer) error {
					// Add a new issue for terraform_autofix_comment rule
					return f.ReplaceText(attribute.Expr.Range(), `"t2.micro" // autofixed`)
				},
			)
		}, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
