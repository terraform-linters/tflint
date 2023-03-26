package rules

import (
	"fmt"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// AwsRoute53RecordEvalOnRootCtxExampleRule checks whether ...
type AwsRoute53RecordEvalOnRootCtxExampleRule struct {
	tflint.DefaultRule
}

// NewAwsRoute53RecordEvalOnRootCtxExampleRule returns a new rule
func NewAwsRoute53RecordEvalOnRootCtxExampleRule() *AwsRoute53RecordEvalOnRootCtxExampleRule {
	return &AwsRoute53RecordEvalOnRootCtxExampleRule{}
}

// Name returns the rule name
func (r *AwsRoute53RecordEvalOnRootCtxExampleRule) Name() string {
	return "aws_route53_record_eval_on_root_ctx_example"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsRoute53RecordEvalOnRootCtxExampleRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsRoute53RecordEvalOnRootCtxExampleRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsRoute53RecordEvalOnRootCtxExampleRule) Link() string {
	return ""
}

// Check checks whether ...
func (r *AwsRoute53RecordEvalOnRootCtxExampleRule) Check(runner tflint.Runner) error {
	resources, err := runner.GetResourceContent("aws_route53_record", &hclext.BodySchema{
		Attributes: []hclext.AttributeSchema{{Name: "name"}},
	}, &tflint.GetModuleContentOption{ModuleCtx: tflint.RootModuleCtxType})
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
				fmt.Sprintf("record name (root): %#v", name),
				attribute.Expr.Range(),
			)
		}, &tflint.EvaluateExprOption{ModuleCtx: tflint.RootModuleCtxType})
		if err != nil {
			return err
		}
	}

	return nil
}
