// This file generated by `generator/`. DO NOT EDIT

package models

import (
	"log"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsDynamoDBTableInvalidHashKeyRule checks the pattern is valid
type AwsDynamoDBTableInvalidHashKeyRule struct {
	resourceType  string
	attributeName string
	max           int
	min           int
}

// NewAwsDynamoDBTableInvalidHashKeyRule returns new rule with default attributes
func NewAwsDynamoDBTableInvalidHashKeyRule() *AwsDynamoDBTableInvalidHashKeyRule {
	return &AwsDynamoDBTableInvalidHashKeyRule{
		resourceType:  "aws_dynamodb_table",
		attributeName: "hash_key",
		max:           255,
		min:           1,
	}
}

// Name returns the rule name
func (r *AwsDynamoDBTableInvalidHashKeyRule) Name() string {
	return "aws_dynamodb_table_invalid_hash_key"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsDynamoDBTableInvalidHashKeyRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsDynamoDBTableInvalidHashKeyRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsDynamoDBTableInvalidHashKeyRule) Link() string {
	return ""
}

// Check checks the pattern is valid
func (r *AwsDynamoDBTableInvalidHashKeyRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		var val string
		err := runner.EvaluateExpr(attribute.Expr, &val)

		return runner.EnsureNoError(err, func() error {
			if len(val) > r.max {
				runner.EmitIssue(
					r,
					"hash_key must be 255 characters or less",
					attribute.Expr.Range(),
				)
			}
			if len(val) < r.min {
				runner.EmitIssue(
					r,
					"hash_key must be 1 characters or higher",
					attribute.Expr.Range(),
				)
			}
			return nil
		})
	})
}
