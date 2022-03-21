package rules

import (
	"fmt"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

// AwsInstanceMapEvalExampleRule checks whether ...
type AwsInstanceMapEvalExampleRule struct {
	tflint.DefaultRule
}

// NewAwsInstanceMapEvalExampleRule returns a new rule
func NewAwsInstanceMapEvalExampleRule() *AwsInstanceMapEvalExampleRule {
	return &AwsInstanceMapEvalExampleRule{}
}

// Name returns the rule name
func (r *AwsInstanceMapEvalExampleRule) Name() string {
	return "aws_instance_map_eval_example"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsInstanceMapEvalExampleRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsInstanceMapEvalExampleRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsInstanceMapEvalExampleRule) Link() string {
	return ""
}

// Check checks whether ...
func (r *AwsInstanceMapEvalExampleRule) Check(runner tflint.Runner) error {
	resources, err := runner.GetResourceContent("aws_instance", &hclext.BodySchema{
		Attributes: []hclext.AttributeSchema{{Name: "tags"}},
	}, nil)
	if err != nil {
		return err
	}

	for _, resource := range resources.Blocks {
		attribute, exists := resource.Body.Attributes["tags"]
		if !exists {
			continue
		}

		wantType := cty.Map(cty.String)
		tags := map[string]string{}
		err := runner.EvaluateExpr(attribute.Expr, &tags, &tflint.EvaluateExprOption{WantType: &wantType})

		err = runner.EnsureNoError(err, func() error {
			return runner.EmitIssue(
				r,
				fmt.Sprintf("instance tags: %#v", tags),
				attribute.Expr.Range(),
			)
		})
		if err != nil {
			return err
		}
	}

	return nil
}
