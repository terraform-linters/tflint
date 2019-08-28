package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/tflint"
)

// AwsALBInvalidSecurityGroupRule checks whether security groups actually exists
type AwsALBInvalidSecurityGroupRule struct {
	resourceType   string
	attributeName  string
	securityGroups map[string]bool
	dataPrepared   bool
}

// NewAwsALBInvalidSecurityGroupRule returns new rule with default attributes
func NewAwsALBInvalidSecurityGroupRule() *AwsALBInvalidSecurityGroupRule {
	return &AwsALBInvalidSecurityGroupRule{
		resourceType:   "aws_alb",
		attributeName:  "security_groups",
		securityGroups: map[string]bool{},
		dataPrepared:   false,
	}
}

// Name returns the rule name
func (r *AwsALBInvalidSecurityGroupRule) Name() string {
	return "aws_alb_invalid_security_group"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsALBInvalidSecurityGroupRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsALBInvalidSecurityGroupRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsALBInvalidSecurityGroupRule) Link() string {
	return ""
}

// Check checks whether `security_groups` are included in the list retrieved by `DescribeSecurityGroups`
func (r *AwsALBInvalidSecurityGroupRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		if !r.dataPrepared {
			log.Print("[DEBUG] Fetch security groups")
			resp, err := runner.AwsClient.EC2.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})
			if err != nil {
				err := &tflint.Error{
					Code:    tflint.ExternalAPIError,
					Level:   tflint.ErrorLevel,
					Message: "An error occurred while describing security groups",
					Cause:   err,
				}
				log.Printf("[ERROR] %s", err)
				return err
			}
			for _, sg := range resp.SecurityGroups {
				r.securityGroups[*sg.GroupId] = true
			}
			r.dataPrepared = true
		}

		return runner.EachStringSliceExprs(attribute.Expr, func(securityGroup string, expr hcl.Expression) {
			if !r.securityGroups[securityGroup] {
				runner.EmitIssue(
					r,
					fmt.Sprintf("\"%s\" is invalid security group.", securityGroup),
					expr.Range(),
				)
			}
		})
	})
}
