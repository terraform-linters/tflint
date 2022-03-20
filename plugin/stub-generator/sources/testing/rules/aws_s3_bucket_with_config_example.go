package rules

import (
	"fmt"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// AwsS3BucketWithConfigExampleRule checks whether ...
type AwsS3BucketWithConfigExampleRule struct {
	tflint.DefaultRule
}

type awsS3BucketWithConfigExampleRuleConfig struct {
	Name string `hclext:"name"`
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
func (r *AwsS3BucketWithConfigExampleRule) Severity() tflint.Severity {
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

	resources, err := runner.GetResourceContent("aws_s3_bucket", &hclext.BodySchema{
		Attributes: []hclext.AttributeSchema{{Name: "bucket"}},
	}, nil)
	if err != nil {
		return err
	}

	for _, resource := range resources.Blocks {
		attribute, exists := resource.Body.Attributes["bucket"]
		if !exists {
			continue
		}

		var bucket string
		err := runner.EvaluateExpr(attribute.Expr, &bucket, nil)

		err = runner.EnsureNoError(err, func() error {
			return runner.EmitIssue(
				r,
				fmt.Sprintf("bucket name is %s, config=%s", bucket, config.Name),
				attribute.Expr.Range(),
			)
		})
		if err != nil {
			return err
		}
	}

	return nil
}
