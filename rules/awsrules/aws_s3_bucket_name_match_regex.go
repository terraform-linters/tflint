package awsrules

import (
	"fmt"
	"log"
	"regexp"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsS3BucketNameRule checks ...
type AwsS3BucketNameRule struct {
	resourceType  string
	attributeName string
	// Add more field
	regex string
}

type awsS3BucketNameMatchRegexConfig struct {
	Regex string `hcl:"regex"`
}

// NewAwsS3BucketNameRule returns new rule with default attributes
func NewAwsS3BucketNameRule() *AwsS3BucketNameRule {
	return &AwsS3BucketNameRule{
		resourceType:  "aws_s3_bucket",
		attributeName: "bucket",
	}
}

// Name returns the rule name
func (r *AwsS3BucketNameRule) Name() string {
	return "aws_s3_bucket_name"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsS3BucketNameRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsS3BucketNameRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsS3BucketNameRule) Link() string {
	return ""
}

// Check if the name of the s3 bucket matches the regex defined in the rule
func (r *AwsS3BucketNameRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())
	config := awsS3BucketNameMatchRegexConfig{}
	if err := runner.DecodeRuleConfig(r.Name(), &config); err != nil {
		return err
	}
	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		var val string
		err := runner.EvaluateExpr(attribute.Expr, &val)
		return runner.EnsureNoError(err, func() error {
			if !regexp.MustCompile(config.Regex).MatchString(val) {
				runner.EmitIssue(
					r,
					fmt.Sprintf(`Bucket name %s does not match regex %s`, val, config.Regex),
					attribute.Expr.Range(),
				)
			}
			return nil
		})
	})
}
