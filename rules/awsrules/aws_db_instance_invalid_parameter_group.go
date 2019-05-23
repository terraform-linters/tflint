package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsDBInstanceInvalidParameterGroupRule checks whether DB parameter group actually exists
type AwsDBInstanceInvalidParameterGroupRule struct {
	resourceType    string
	attributeName   string
	parameterGroups map[string]bool
	dataPrepared    bool
}

// NewAwsDBInstanceInvalidParameterGroupRule returns new rule with default attributes
func NewAwsDBInstanceInvalidParameterGroupRule() *AwsDBInstanceInvalidParameterGroupRule {
	return &AwsDBInstanceInvalidParameterGroupRule{
		resourceType:    "aws_db_instance",
		attributeName:   "parameter_group_name",
		parameterGroups: map[string]bool{},
		dataPrepared:    false,
	}
}

// Name returns the rule name
func (r *AwsDBInstanceInvalidParameterGroupRule) Name() string {
	return "aws_db_instance_invalid_parameter_group"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsDBInstanceInvalidParameterGroupRule) Enabled() bool {
	return true
}

// Type returns the rule severity
func (r *AwsDBInstanceInvalidParameterGroupRule) Type() string {
	return issue.ERROR
}

// Link returns the rule reference link
func (r *AwsDBInstanceInvalidParameterGroupRule) Link() string {
	return ""
}

// Check checks whether `parameter_group_name` are included in the list retrieved by `DescribeDBParameterGroups`
func (r *AwsDBInstanceInvalidParameterGroupRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		if !r.dataPrepared {
			log.Print("[DEBUG] Fetch DB parameter groups")
			resp, err := runner.AwsClient.RDS.DescribeDBParameterGroups(&rds.DescribeDBParameterGroupsInput{})
			if err != nil {
				err := &tflint.Error{
					Code:    tflint.ExternalAPIError,
					Level:   tflint.ErrorLevel,
					Message: "An error occurred while describing DB parameter groups",
					Cause:   err,
				}
				log.Printf("[ERROR] %s", err)
				return err
			}
			for _, parameterGroup := range resp.DBParameterGroups {
				r.parameterGroups[*parameterGroup.DBParameterGroupName] = true
			}
			r.dataPrepared = true
		}

		var parameterGroup string
		err := runner.EvaluateExpr(attribute.Expr, &parameterGroup)

		return runner.EnsureNoError(err, func() error {
			if !r.parameterGroups[parameterGroup] {
				runner.EmitIssue(
					r,
					fmt.Sprintf("\"%s\" is invalid parameter group name.", parameterGroup),
					attribute.Expr.Range(),
				)
			}
			return nil
		})
	})
}
