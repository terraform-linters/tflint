package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/tflint"
)

// AwsDBInstanceInvalidOptionGroupRule checks whether option group actually exists
type AwsDBInstanceInvalidOptionGroupRule struct {
	resourceType  string
	attributeName string
	optionGroups  map[string]bool
	dataPrepared  bool
}

// NewAwsDBInstanceInvalidOptionGroupRule returns new rule with default attributes
func NewAwsDBInstanceInvalidOptionGroupRule() *AwsDBInstanceInvalidOptionGroupRule {
	return &AwsDBInstanceInvalidOptionGroupRule{
		resourceType:  "aws_db_instance",
		attributeName: "option_group_name",
		optionGroups:  map[string]bool{},
		dataPrepared:  false,
	}
}

// Name returns the rule name
func (r *AwsDBInstanceInvalidOptionGroupRule) Name() string {
	return "aws_db_instance_invalid_option_group"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsDBInstanceInvalidOptionGroupRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsDBInstanceInvalidOptionGroupRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsDBInstanceInvalidOptionGroupRule) Link() string {
	return ""
}

// Check checks whether `option_group_name` are included in the list retrieved by `DescribeOptionGroups`
func (r *AwsDBInstanceInvalidOptionGroupRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		if !r.dataPrepared {
			log.Print("[DEBUG] Fetch option groups")
			resp, err := runner.AwsClient.RDS.DescribeOptionGroups(&rds.DescribeOptionGroupsInput{})
			if err != nil {
				err := &tflint.Error{
					Code:    tflint.ExternalAPIError,
					Level:   tflint.ErrorLevel,
					Message: "An error occurred while describing option groups",
					Cause:   err,
				}
				log.Printf("[ERROR] %s", err)
				return err
			}
			for _, optionGroup := range resp.OptionGroupsList {
				r.optionGroups[*optionGroup.OptionGroupName] = true
			}
			r.dataPrepared = true
		}

		var optionGroup string
		err := runner.EvaluateExpr(attribute.Expr, &optionGroup)

		return runner.EnsureNoError(err, func() error {
			if !r.optionGroups[optionGroup] {
				runner.EmitIssue(
					r,
					fmt.Sprintf("\"%s\" is invalid option group name.", optionGroup),
					attribute.Expr.Range(),
				)
			}
			return nil
		})
	})
}
