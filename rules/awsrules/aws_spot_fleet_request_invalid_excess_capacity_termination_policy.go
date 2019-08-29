package awsrules

import (
	"log"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/tflint"
)

// AwsSpotFleetRequestInvalidExcessCapacityTerminationPolicyRule checks the pattern is valid
type AwsSpotFleetRequestInvalidExcessCapacityTerminationPolicyRule struct {
	resourceType  string
	attributeName string
	enum          []string
}

// NewAwsSpotFleetRequestInvalidExcessCapacityTerminationPolicyRule returns new rule with default attributes
func NewAwsSpotFleetRequestInvalidExcessCapacityTerminationPolicyRule() *AwsSpotFleetRequestInvalidExcessCapacityTerminationPolicyRule {
	return &AwsSpotFleetRequestInvalidExcessCapacityTerminationPolicyRule{
		resourceType:  "aws_spot_fleet_request",
		attributeName: "excess_capacity_termination_policy",
		enum: []string{
			"Default",
			"NoTermination",
		},
	}
}

// Name returns the rule name
func (r *AwsSpotFleetRequestInvalidExcessCapacityTerminationPolicyRule) Name() string {
	return "aws_spot_fleet_request_invalid_excess_capacity_termination_policy"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsSpotFleetRequestInvalidExcessCapacityTerminationPolicyRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsSpotFleetRequestInvalidExcessCapacityTerminationPolicyRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsSpotFleetRequestInvalidExcessCapacityTerminationPolicyRule) Link() string {
	return ""
}

// Check checks the pattern is valid
func (r *AwsSpotFleetRequestInvalidExcessCapacityTerminationPolicyRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

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
					`excess_capacity_termination_policy is not a valid value`,
					attribute.Expr.Range(),
				)
			}
			return nil
		})
	})
}
