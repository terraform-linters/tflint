package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsDBInstanceInvalidDBSubnetGroupRule checks whether DB subnet group actually exists
type AwsDBInstanceInvalidDBSubnetGroupRule struct {
	resourceType  string
	attributeName string
	subnetGroups  map[string]bool
	dataPrepared  bool
}

// NewAwsDBInstanceInvalidDBSubnetGroupRule returns new rule with default attributes
func NewAwsDBInstanceInvalidDBSubnetGroupRule() *AwsDBInstanceInvalidDBSubnetGroupRule {
	return &AwsDBInstanceInvalidDBSubnetGroupRule{
		resourceType:  "aws_db_instance",
		attributeName: "db_subnet_group_name",
		subnetGroups:  map[string]bool{},
		dataPrepared:  false,
	}
}

// Name returns the rule name
func (r *AwsDBInstanceInvalidDBSubnetGroupRule) Name() string {
	return "aws_db_instance_invalid_db_subnet_group"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsDBInstanceInvalidDBSubnetGroupRule) Enabled() bool {
	return true
}

// Check checks whether `db_subnet_group_name` are included in the list retrieved by `DescribeDBSubnetGroups`
func (r *AwsDBInstanceInvalidDBSubnetGroupRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		if !r.dataPrepared {
			log.Print("[DEBUG] Fetch DB subnet groups")
			resp, err := runner.AwsClient.RDS.DescribeDBSubnetGroups(&rds.DescribeDBSubnetGroupsInput{})
			if err != nil {
				err := &tflint.Error{
					Code:    tflint.ExternalAPIError,
					Level:   tflint.ErrorLevel,
					Message: "An error occurred while describing DB subnet groups",
					Cause:   err,
				}
				log.Printf("[ERROR] %s", err)
				return err
			}
			for _, subnetGroup := range resp.DBSubnetGroups {
				r.subnetGroups[*subnetGroup.DBSubnetGroupName] = true
			}
			r.dataPrepared = true
		}

		var subnetGroup string
		err := runner.EvaluateExpr(attribute.Expr, &subnetGroup)

		return runner.EnsureNoError(err, func() error {
			if !r.subnetGroups[subnetGroup] {
				runner.Issues = append(runner.Issues, &issue.Issue{
					Detector: r.Name(),
					Type:     issue.ERROR,
					Message:  fmt.Sprintf("\"%s\" is invalid DB subnet group name.", subnetGroup),
					Line:     attribute.Range.Start.Line,
					File:     runner.GetFileName(attribute.Range.Filename),
				})
			}
			return nil
		})
	})
}
