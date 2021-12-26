package rules

import (
	"fmt"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// AwsS3BucketWithConfigExampleRule checks whether ...
type AwsS3BucketWithConfigExampleRule struct{}

type awsS3BucketWithConfigExampleRuleConfig struct {
	Name string `hcl:"name"`

	Remain hcl.Body `hcl:",remain"`
}

// NewAwsS3BucketWithConfigExampleRule returns a new rule
func NewAwsS3BucketWithConfigExampleRule() *AwsS3BucketWithConfigExampleRule {
	return &AwsS3BucketWithConfigExampleRule{}
}

// Name returns the rule name
func (r *AwsS3BucketWithConfigExampleRule) Name() string {
	return "aws_s3_bucket_with_config_example"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsS3BucketWithConfigExampleRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *AwsS3BucketWithConfigExampleRule) Severity() string {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *AwsS3BucketWithConfigExampleRule) Link() string {
	return ""
}

// Check checks whether ...
func (r *AwsS3BucketWithConfigExampleRule) Check(runner tflint.Runner) error {
	config := awsS3BucketWithConfigExampleRuleConfig{}
	if err := runner.DecodeRuleConfig(r.Name(), &config); err != nil {
		return err
	}

	return runner.WalkResourceAttributes("aws_s3_bucket", "bucket", func(attribute *hcl.Attribute) error {
		var bucket string
		err := runner.EvaluateExpr(attribute.Expr, &bucket, nil)

		return runner.EnsureNoError(err, func() error {
			return runner.EmitIssueOnExpr(
				r,
				fmt.Sprintf("bucket name is %s, config=%s", bucket, config.Name),
				attribute.Expr,
			)
		})
	})
}
