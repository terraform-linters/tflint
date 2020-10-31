package awsrules

import (
	"fmt"
	"log"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsDBInstanceInvalidTypeRule checks whether "aws_db_instance" has invalid intance type.
type AwsDBInstanceInvalidTypeRule struct {
	resourceType  string
	attributeName string
	instanceTypes map[string]bool
}

// NewAwsDBInstanceInvalidTypeRule returns new rule with default attributes
func NewAwsDBInstanceInvalidTypeRule() *AwsDBInstanceInvalidTypeRule {
	return &AwsDBInstanceInvalidTypeRule{
		resourceType:  "aws_db_instance",
		attributeName: "instance_class",
		instanceTypes: map[string]bool{
			"db.cr1.8xlarge":   true,
			"db.cv11.18xlarge": true,
			"db.cv11.2xlarge":  true,
			"db.cv11.4xlarge":  true,
			"db.cv11.9xlarge":  true,
			"db.cv11.large":    true,
			"db.cv11.medium":   true,
			"db.cv11.small":    true,
			"db.cv11.xlarge":   true,
			"db.m1.large":      true,
			"db.m1.medium":     true,
			"db.m1.small":      true,
			"db.m1.xlarge":     true,
			"db.m2.2xlarge":    true,
			"db.m2.4xlarge":    true,
			"db.m2.xlarge":     true,
			"db.m3.2xlarge":    true,
			"db.m3.large":      true,
			"db.m3.medium":     true,
			"db.m3.xlarge":     true,
			"db.m4.10xlarge":   true,
			"db.m4.16xlarge":   true,
			"db.m4.2xlarge":    true,
			"db.m4.4xlarge":    true,
			"db.m4.large":      true,
			"db.m4.xlarge":     true,
			"db.m5.12xlarge":   true,
			"db.m5.16xlarge":   true,
			"db.m5.24xlarge":   true,
			"db.m5.2xlarge":    true,
			"db.m5.4xlarge":    true,
			"db.m5.8xlarge":    true,
			"db.m5.large":      true,
			"db.m5.xlarge":     true,
			"db.m6g.16xlarge":  true,
			"db.m6g.12xlarge":  true,
			"db.m6g.8xlarge":   true,
			"db.m6g.4xlarge":   true,
			"db.m6g.2xlarge":   true,
			"db.m6g.xlarge":    true,
			"db.m6g.large":     true,
			"db.mv11.12xlarge": true,
			"db.mv11.24xlarge": true,
			"db.mv11.2xlarge":  true,
			"db.mv11.4xlarge":  true,
			"db.mv11.large":    true,
			"db.mv11.medium":   true,
			"db.mv11.xlarge":   true,
			"db.r3.2xlarge":    true,
			"db.r3.4xlarge":    true,
			"db.r3.8xlarge":    true,
			"db.r3.large":      true,
			"db.r3.xlarge":     true,
			"db.r4.16xlarge":   true,
			"db.r4.2xlarge":    true,
			"db.r4.4xlarge":    true,
			"db.r4.8xlarge":    true,
			"db.r4.large":      true,
			"db.r4.xlarge":     true,
			"db.r5.12xlarge":   true,
			"db.r5.16xlarge":   true,
			"db.r5.24xlarge":   true,
			"db.r5.2xlarge":    true,
			"db.r5.4xlarge":    true,
			"db.r5.8xlarge":    true,
			"db.r5.large":      true,
			"db.r5.xlarge":     true,
			"db.r6g.16xlarge":  true,
			"db.r6g.12xlarge":  true,
			"db.r6g.4xlarge":   true,
			"db.r6g.2xlarge":   true,
			"db.r6g.xlarge":    true,
			"db.r6g.large":     true,
			"db.rv11.12xlarge": true,
			"db.rv11.24xlarge": true,
			"db.rv11.2xlarge":  true,
			"db.rv11.4xlarge":  true,
			"db.rv11.large":    true,
			"db.rv11.xlarge":   true,
			"db.t1.micro":      true,
			"db.t2.2xlarge":    true,
			"db.t2.large":      true,
			"db.t2.medium":     true,
			"db.t2.micro":      true,
			"db.t2.small":      true,
			"db.t2.xlarge":     true,
			"db.t3.2xlarge":    true,
			"db.t3.large":      true,
			"db.t3.medium":     true,
			"db.t3.micro":      true,
			"db.t3.small":      true,
			"db.t3.xlarge":     true,
			"db.x1.16xlarge":   true,
			"db.x1.32xlarge":   true,
			"db.x1e.16xlarge":  true,
			"db.x1e.2xlarge":   true,
			"db.x1e.32xlarge":  true,
			"db.x1e.4xlarge":   true,
			"db.x1e.8xlarge":   true,
			"db.x1e.xlarge":    true,
			"db.z1d.12xlarge":  true,
			"db.z1d.2xlarge":   true,
			"db.z1d.3xlarge":   true,
			"db.z1d.6xlarge":   true,
			"db.z1d.large":     true,
			"db.z1d.xlarge":    true,
		},
	}
}

// Name returns the rule name
func (r *AwsDBInstanceInvalidTypeRule) Name() string {
	return "aws_db_instance_invalid_type"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsDBInstanceInvalidTypeRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsDBInstanceInvalidTypeRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsDBInstanceInvalidTypeRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether "aws_db_instance" has invalid instance type.
func (r *AwsDBInstanceInvalidTypeRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

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
