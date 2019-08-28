package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/tflint"
)

// AwsDBInstanceInvalidVPCSecurityGroupRule checks whether security groups actually exists
type AwsDBInstanceInvalidVPCSecurityGroupRule struct {
	resourceType   string
	attributeName  string
	securityGroups map[string]bool
	dataPrepared   bool
}

// NewAwsDBInstanceInvalidVPCSecurityGroupRule returns new rule with default attributes
func NewAwsDBInstanceInvalidVPCSecurityGroupRule() *AwsDBInstanceInvalidVPCSecurityGroupRule {
	return &AwsDBInstanceInvalidVPCSecurityGroupRule{
		resourceType:   "aws_db_instance",
		attributeName:  "vpc_security_group_ids",
		securityGroups: map[string]bool{},
		dataPrepared:   false,
	}
}

// Name returns the rule name
func (r *AwsDBInstanceInvalidVPCSecurityGroupRule) Name() string {
	return "aws_db_instance_invalid_vpc_security_group"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsDBInstanceInvalidVPCSecurityGroupRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsDBInstanceInvalidVPCSecurityGroupRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsDBInstanceInvalidVPCSecurityGroupRule) Link() string {
	return ""
}

// Check checks whether `vpc_security_groups` are included in the list retrieved by `DescribeSecurityGroups`
func (r *AwsDBInstanceInvalidVPCSecurityGroupRule) Check(runner *tflint.Runner) error {
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
			for _, securityGroup := range resp.SecurityGroups {
				r.securityGroups[*securityGroup.GroupId] = true
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
