package rules

import (
	"fmt"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// AwsIAMPolicyExampleRule checks whether ...
type AwsIAMPolicyExampleRule struct {
	tflint.DefaultRule
}

// NewAwsIAMPolicyExampleRule returns a new rule
func NewAwsIAMPolicyExampleRule() *AwsIAMPolicyExampleRule {
	return &AwsIAMPolicyExampleRule{}
}

// Name returns the rule name
func (r *AwsIAMPolicyExampleRule) Name() string {
	return "aws_iam_policy_example"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsIAMPolicyExampleRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsIAMPolicyExampleRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsIAMPolicyExampleRule) Link() string {
	return ""
}

// Check checks whether ...
func (r *AwsIAMPolicyExampleRule) Check(runner tflint.Runner) error {
	resources, err := runner.GetResourceContent("aws_iam_policy", &hclext.BodySchema{
		Attributes: []hclext.AttributeSchema{{Name: "name"}},
	}, &tflint.GetModuleContentOption{
		ModuleCtx:  tflint.SelfModuleCtxType,
		ExpandMode: tflint.ExpandModeNone,
	})
	if err != nil {
		return err
	}

	for _, resource := range resources.Blocks {
		attribute, exists := resource.Body.Attributes["name"]
		if !exists {
			continue
		}

		err := runner.EvaluateExpr(attribute.Expr, func(name string) error {
			return runner.EmitIssue(
				r,
				fmt.Sprintf("name is %s", name),
				attribute.Expr.Range(),
			)
		}, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
