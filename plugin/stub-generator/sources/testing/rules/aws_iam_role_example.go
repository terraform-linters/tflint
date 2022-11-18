package rules

import (
	"fmt"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// AwsIAMRoleExampleRule checks whether ...
type AwsIAMRoleExampleRule struct {
	tflint.DefaultRule
}

// NewAwsIAMPolicyExampleRule returns a new rule
func NewAwsIAMRoleExampleRule() *AwsIAMRoleExampleRule {
	return &AwsIAMRoleExampleRule{}
}

// Name returns the rule name
func (r *AwsIAMRoleExampleRule) Name() string {
	return "aws_iam_role_example"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsIAMRoleExampleRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsIAMRoleExampleRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsIAMRoleExampleRule) Link() string {
	return ""
}

// Check checks whether ...
func (r *AwsIAMRoleExampleRule) Check(runner tflint.Runner) error {
	resources, err := runner.GetResourceContent("aws_iam_role", &hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type: "inline_policy",
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "name"}},
				},
			},
		},
	}, &tflint.GetModuleContentOption{
		ModuleCtx:  tflint.SelfModuleCtxType,
		ExpandMode: tflint.ExpandModeNone,
	})
	if err != nil {
		return err
	}

	for _, resource := range resources.Blocks {
		for _, policy := range resource.Body.Blocks {
			if err := runner.EmitIssue(r, "inline policy found", policy.DefRange); err != nil {
				return err
			}

			attribute, exists := policy.Body.Attributes["name"]
			if !exists {
				continue
			}

			var name string
			err := runner.EvaluateExpr(attribute.Expr, &name, nil)

			err = runner.EnsureNoError(err, func() error {
				return runner.EmitIssue(
					r,
					fmt.Sprintf("name is %s", name),
					attribute.Expr.Range(),
				)
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
