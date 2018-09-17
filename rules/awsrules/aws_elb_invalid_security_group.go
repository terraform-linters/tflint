package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsELBInvalidSecurityGroupRule checks whether security groups actually exists
type AwsELBInvalidSecurityGroupRule struct {
	resourceType   string
	attributeName  string
	securityGroups map[string]bool
	dataPrepared   bool
}

// NewAwsELBInvalidSecurityGroupRule returns new rule with default attributes
func NewAwsELBInvalidSecurityGroupRule() *AwsELBInvalidSecurityGroupRule {
	return &AwsELBInvalidSecurityGroupRule{
		resourceType:   "aws_elb",
		attributeName:  "security_groups",
		securityGroups: map[string]bool{},
		dataPrepared:   false,
	}
}

// Name returns the rule name
func (r *AwsELBInvalidSecurityGroupRule) Name() string {
	return "aws_elb_invalid_security_group"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsELBInvalidSecurityGroupRule) Enabled() bool {
	return true
}

// Check checks whether `security_groups` are included in the list retrieved by `DescribeSecurityGroups`
func (r *AwsELBInvalidSecurityGroupRule) Check(runner *tflint.Runner) error {
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

		var securityGroups []string
		err := runner.EvaluateExpr(attribute.Expr, &securityGroups)

		return runner.EnsureNoError(err, func() error {
			for _, securityGroup := range securityGroups {
				if !r.securityGroups[securityGroup] {
					runner.Issues = append(runner.Issues, &issue.Issue{
						Detector: r.Name(),
						Type:     issue.ERROR,
						Message:  fmt.Sprintf("\"%s\" is invalid security group.", securityGroup),
						Line:     attribute.Range.Start.Line,
						File:     runner.GetFileName(attribute.Range.Filename),
					})
				}
			}
			return nil
		})
	})
}
