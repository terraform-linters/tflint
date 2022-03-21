package rules

import (
	"fmt"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

// AwsAutoscalingGroupCtyEvalExampleRule checks whether ...
type AwsAutoscalingGroupCtyEvalExampleRule struct {
	tflint.DefaultRule
}

// NewAwsAutoScalingGroupCtyEvalExample returns a new rule
func NewAwsAutoscalingGroupCtyEvalExampleRule() *AwsAutoscalingGroupCtyEvalExampleRule {
	return &AwsAutoscalingGroupCtyEvalExampleRule{}
}

// Name returns the rule name
func (r *AwsAutoscalingGroupCtyEvalExampleRule) Name() string {
	return "aws_autoscaling_group_cty_eval_example"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsAutoscalingGroupCtyEvalExampleRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsAutoscalingGroupCtyEvalExampleRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsAutoscalingGroupCtyEvalExampleRule) Link() string {
	return ""
}

// Check checks whether ...
func (r *AwsAutoscalingGroupCtyEvalExampleRule) Check(runner tflint.Runner) error {
	type tag struct {
		Key               string `cty:"key"`
		Value             string `cty:"value"`
		PropagateAtLaunch bool   `cty:"propagate_at_launch"`
	}

	resources, err := runner.GetResourceContent("aws_autoscaling_group", &hclext.BodySchema{
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

		wantType := cty.List(cty.Object(map[string]cty.Type{
			"key":                 cty.String,
			"value":               cty.String,
			"propagate_at_launch": cty.Bool,
		}))
		var tags []tag
		err := runner.EvaluateExpr(attribute.Expr, &tags, &tflint.EvaluateExprOption{WantType: &wantType})

		err = runner.EnsureNoError(err, func() error {
			return runner.EmitIssue(
				r,
				fmt.Sprintf("autoscaling tags: %#v", tags),
				attribute.Expr.Range(),
			)
		})
		if err != nil {
			return err
		}
	}

	return nil
}
