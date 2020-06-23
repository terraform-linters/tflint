// This file generated by `generator/`. DO NOT EDIT

package models

import (
	"fmt"
	"log"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsEcsServiceInvalidPropagateTagsRule checks the pattern is valid
type AwsEcsServiceInvalidPropagateTagsRule struct {
	resourceType  string
	attributeName string
	enum          []string
}

// NewAwsEcsServiceInvalidPropagateTagsRule returns new rule with default attributes
func NewAwsEcsServiceInvalidPropagateTagsRule() *AwsEcsServiceInvalidPropagateTagsRule {
	return &AwsEcsServiceInvalidPropagateTagsRule{
		resourceType:  "aws_ecs_service",
		attributeName: "propagate_tags",
		enum: []string{
			"TASK_DEFINITION",
			"SERVICE",
		},
	}
}

// Name returns the rule name
func (r *AwsEcsServiceInvalidPropagateTagsRule) Name() string {
	return "aws_ecs_service_invalid_propagate_tags"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsEcsServiceInvalidPropagateTagsRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsEcsServiceInvalidPropagateTagsRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsEcsServiceInvalidPropagateTagsRule) Link() string {
	return ""
}

// Check checks the pattern is valid
func (r *AwsEcsServiceInvalidPropagateTagsRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		var val string
		err := runner.EvaluateExpr(attribute.Expr, &val)

		return runner.EnsureNoError(err, func() error {
			found := false
			for _, item := range r.enum {
				if item == val {
					found = true
				}
			}
			if !found {
				runner.EmitIssue(
					r,
					fmt.Sprintf(`"%s" is an invalid value as propagate_tags`, truncateLongMessage(val)),
					attribute.Expr.Range(),
				)
			}
			return nil
		})
	})
}
