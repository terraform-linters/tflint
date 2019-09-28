package awsrules

import (
	"log"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/tflint"
)

// AwsS3BucketInvalidACLRule checks the pattern is valid
type AwsS3BucketInvalidACLRule struct {
	resourceType  string
	attributeName string
	enum          []string
}

// NewAwsS3BucketInvalidACLRule returns new rule with default attributes
func NewAwsS3BucketInvalidACLRule() *AwsS3BucketInvalidACLRule {
	return &AwsS3BucketInvalidACLRule{
		resourceType:  "aws_s3_bucket",
		attributeName: "acl",
		enum: []string{
			"private",
			"public-read",
			"public-read-write",
			"authenticated-read",
			"log-delivery-write",
		},
	}
}

// Name returns the rule name
func (r *AwsS3BucketInvalidACLRule) Name() string {
	return "aws_s3_bucket_invalid_acl"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsS3BucketInvalidACLRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsS3BucketInvalidACLRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsS3BucketInvalidACLRule) Link() string {
	return ""
}

// Check checks the pattern is valid
func (r *AwsS3BucketInvalidACLRule) Check(runner *tflint.Runner) error {
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
					`acl is not a valid value`,
					attribute.Expr.Range(),
				)
			}
			return nil
		})
	})
}
