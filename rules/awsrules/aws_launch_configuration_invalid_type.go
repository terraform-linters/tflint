package awsrules

import (
	"fmt"
	"log"

	instances "github.com/cristim/ec2-instances-info"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsLaunchConfigurationInvalidTypeRule checks whether "aws_instance" has invalid intance type.
type AwsLaunchConfigurationInvalidTypeRule struct {
	resourceType  string
	attributeName string
	instanceTypes map[string]bool
}

// NewAwsLaunchConfigurationInvalidTypeRule returns new rule with default attributes
func NewAwsLaunchConfigurationInvalidTypeRule() *AwsLaunchConfigurationInvalidTypeRule {
	rule := &AwsLaunchConfigurationInvalidTypeRule{
		resourceType:  "aws_launch_configuration",
		attributeName: "instance_type",
		instanceTypes: map[string]bool{},
	}

	data, err := instances.Data()
	if err != nil {
		// Maybe this is bug
		panic(err)
	}

	for _, i := range *data {
		rule.instanceTypes[i.InstanceType] = true
	}

	return rule
}

// Name returns the rule name
func (r *AwsLaunchConfigurationInvalidTypeRule) Name() string {
	return "aws_launch_configuration_invalid_type"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsLaunchConfigurationInvalidTypeRule) Enabled() bool {
	return true
}

// Type returns the rule severity
func (r *AwsLaunchConfigurationInvalidTypeRule) Type() string {
	return issue.ERROR
}

// Link returns the rule reference link
func (r *AwsLaunchConfigurationInvalidTypeRule) Link() string {
	return ""
}

// Check checks whether "aws_instance" has invalid instance type.
func (r *AwsLaunchConfigurationInvalidTypeRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		var instanceType string
		err := runner.EvaluateExpr(attribute.Expr, &instanceType)

		return runner.EnsureNoError(err, func() error {
			if !r.instanceTypes[instanceType] {
				runner.EmitIssue(
					r,
					fmt.Sprintf("\"%s\" is invalid instance type.", instanceType),
					attribute.Expr.Range(),
				)
			}
			return nil
		})
	})
}
