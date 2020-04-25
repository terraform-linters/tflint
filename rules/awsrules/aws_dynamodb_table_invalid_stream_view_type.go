package awsrules

import (
	"fmt"
	"log"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsDynamoDBTableInvalidStreamViewTypeRule checks the pattern is valid
type AwsDynamoDBTableInvalidStreamViewTypeRule struct {
	resourceType  string
	attributeName string
	enum          []string
}

// NewAwsDynamoDBTableInvalidStreamViewTypeRule returns new rule with default attributes
func NewAwsDynamoDBTableInvalidStreamViewTypeRule() *AwsDynamoDBTableInvalidStreamViewTypeRule {
	return &AwsDynamoDBTableInvalidStreamViewTypeRule{
		resourceType:  "aws_dynamodb_table",
		attributeName: "stream_view_type",
		enum: []string{
			"",
			"NEW_IMAGE",
			"OLD_IMAGE",
			"NEW_AND_OLD_IMAGES",
			"KEYS_ONLY",
		},
	}
}

// Name returns the rule name
func (r *AwsDynamoDBTableInvalidStreamViewTypeRule) Name() string {
	return "aws_dynamodb_table_invalid_stream_view_type"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsDynamoDBTableInvalidStreamViewTypeRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsDynamoDBTableInvalidStreamViewTypeRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsDynamoDBTableInvalidStreamViewTypeRule) Link() string {
	return ""
}

// Check checks the pattern is valid
func (r *AwsDynamoDBTableInvalidStreamViewTypeRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

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
					fmt.Sprintf(`"%s" is an invalid value as stream_view_type`, val),
					attribute.Expr.Range(),
				)
			}
			return nil
		})
	})
}
