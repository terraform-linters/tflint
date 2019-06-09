package awsrules

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsS3BucketInvalidACLRule checks whether "aws_s3_bucket" has invalid ACL setting.
type AwsS3BucketInvalidACLRule struct {
	resourceType  string
	attributeName string
	aclTypes      map[string]bool
}

// NewAwsS3BucketInvalidACLRule returns new rule with default attributes
func NewAwsS3BucketInvalidACLRule() *AwsS3BucketInvalidACLRule {
	return &AwsS3BucketInvalidACLRule{
		resourceType:  "aws_s3_bucket",
		attributeName: "acl",
		aclTypes: map[string]bool{
			"private":                   true,
			"public-read":               true,
			"public-read-write":         true,
			"aws-exec-read":             true,
			"authenticated-read":        true,
			"bucket-owner-read":         true,
			"bucket-owner-full-control": true,
			"log-delivery-write":        true,
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

// Type returns the rule severity
func (r *AwsS3BucketInvalidACLRule) Type() string {
	return issue.ERROR
}

// Link returns the rule reference link
func (r *AwsS3BucketInvalidACLRule) Link() string {
	return ""
}

// Check checks whether "aws_s3_bucket" has invalid ACL type.
func (r *AwsS3BucketInvalidACLRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		var acl string
		err := runner.EvaluateExpr(attribute.Expr, &acl)

		return runner.EnsureNoError(err, func() error {
			if !r.aclTypes[acl] {
				runner.EmitIssue(
					r,
					fmt.Sprintf("\"%s\" is invalid canned ACL type.", acl),
					attribute.Expr.Range(),
				)
			}
			return nil
		})
	})
}
