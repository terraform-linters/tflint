// This file generated by `generator/`. DO NOT EDIT

package models

import (
	"fmt"
	"log"
	"regexp"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsWorklinkFleetInvalidNameRule checks the pattern is valid
type AwsWorklinkFleetInvalidNameRule struct {
	resourceType  string
	attributeName string
	max           int
	min           int
	pattern       *regexp.Regexp
}

// NewAwsWorklinkFleetInvalidNameRule returns new rule with default attributes
func NewAwsWorklinkFleetInvalidNameRule() *AwsWorklinkFleetInvalidNameRule {
	return &AwsWorklinkFleetInvalidNameRule{
		resourceType:  "aws_worklink_fleet",
		attributeName: "name",
		max:           48,
		min:           1,
		pattern:       regexp.MustCompile(`^[a-z0-9](?:[a-z0-9\-]{0,46}[a-z0-9])?$`),
	}
}

// Name returns the rule name
func (r *AwsWorklinkFleetInvalidNameRule) Name() string {
	return "aws_worklink_fleet_invalid_name"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsWorklinkFleetInvalidNameRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsWorklinkFleetInvalidNameRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsWorklinkFleetInvalidNameRule) Link() string {
	return ""
}

// Check checks the pattern is valid
func (r *AwsWorklinkFleetInvalidNameRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		var val string
		err := runner.EvaluateExpr(attribute.Expr, &val)

		return runner.EnsureNoError(err, func() error {
			if len(val) > r.max {
				runner.EmitIssue(
					r,
					"name must be 48 characters or less",
					attribute.Expr.Range(),
				)
			}
			if len(val) < r.min {
				runner.EmitIssue(
					r,
					"name must be 1 characters or higher",
					attribute.Expr.Range(),
				)
			}
			if !r.pattern.MatchString(val) {
				runner.EmitIssue(
					r,
					fmt.Sprintf(`"%s" does not match valid pattern %s`, truncateLongMessage(val), `^[a-z0-9](?:[a-z0-9\-]{0,46}[a-z0-9])?$`),
					attribute.Expr.Range(),
				)
			}
			return nil
		})
	})
}
