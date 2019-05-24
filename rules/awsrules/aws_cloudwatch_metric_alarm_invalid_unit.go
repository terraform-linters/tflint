package awsrules

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsCloudwatchMetricAlarmInvalidUnitRule checks whether the valid unit is used in an alerm
type AwsCloudwatchMetricAlarmInvalidUnitRule struct {
	resourceType  string
	attributeName string
	validUnits    map[string]bool
}

// NewAwsCloudwatchMetricAlarmInvalidUnitRule returns new rule with default attributes
func NewAwsCloudwatchMetricAlarmInvalidUnitRule() *AwsCloudwatchMetricAlarmInvalidUnitRule {
	return &AwsCloudwatchMetricAlarmInvalidUnitRule{
		resourceType:  "aws_cloudwatch_metric_alarm",
		attributeName: "unit",
		// @see http://docs.aws.amazon.com/cli/latest/reference/cloudwatch/put-metric-alarm.html
		validUnits: map[string]bool{
			"Seconds":          true,
			"Microseconds":     true,
			"Milliseconds":     true,
			"Bytes":            true,
			"Kilobytes":        true,
			"Megabytes":        true,
			"Gigabytes":        true,
			"Terabytes":        true,
			"Bits":             true,
			"Kilobits":         true,
			"Megabits":         true,
			"Gigabits":         true,
			"Terabits":         true,
			"Percent":          true,
			"Count":            true,
			"Bytes/Second":     true,
			"Kilobytes/Second": true,
			"Megabytes/Second": true,
			"Gigabytes/Second": true,
			"Terabytes/Second": true,
			"Bits/Second":      true,
			"Kilobits/Second":  true,
			"Megabits/Second":  true,
			"Gigabits/Second":  true,
			"Terabits/Second":  true,
			"Count/Second":     true,
			"None":             true,
		},
	}
}

// Name returns the rule name
func (r *AwsCloudwatchMetricAlarmInvalidUnitRule) Name() string {
	return "aws_cloudwatch_metric_alarm_invalid_unit"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsCloudwatchMetricAlarmInvalidUnitRule) Enabled() bool {
	return true
}

// Type returns the rule severity
func (r *AwsCloudwatchMetricAlarmInvalidUnitRule) Type() string {
	return issue.ERROR
}

// Link returns the rule reference link
func (r *AwsCloudwatchMetricAlarmInvalidUnitRule) Link() string {
	return ""
}

// Check checks whether `unit` is included in the valid unit list
func (r *AwsCloudwatchMetricAlarmInvalidUnitRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		var unit string
		err := runner.EvaluateExpr(attribute.Expr, &unit)

		return runner.EnsureNoError(err, func() error {
			if !r.validUnits[unit] {
				runner.EmitIssue(
					r,
					fmt.Sprintf("\"%s\" is invalid unit.", unit),
					attribute.Expr.Range(),
				)
			}
			return nil
		})
	})
}
