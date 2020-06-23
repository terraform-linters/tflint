// This file generated by `generator/`. DO NOT EDIT

package models

import (
	"fmt"
	"log"
	"regexp"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsElastictranscoderPipelineInvalidOutputBucketRule checks the pattern is valid
type AwsElastictranscoderPipelineInvalidOutputBucketRule struct {
	resourceType  string
	attributeName string
	pattern       *regexp.Regexp
}

// NewAwsElastictranscoderPipelineInvalidOutputBucketRule returns new rule with default attributes
func NewAwsElastictranscoderPipelineInvalidOutputBucketRule() *AwsElastictranscoderPipelineInvalidOutputBucketRule {
	return &AwsElastictranscoderPipelineInvalidOutputBucketRule{
		resourceType:  "aws_elastictranscoder_pipeline",
		attributeName: "output_bucket",
		pattern:       regexp.MustCompile(`^(\w|\.|-){1,255}$`),
	}
}

// Name returns the rule name
func (r *AwsElastictranscoderPipelineInvalidOutputBucketRule) Name() string {
	return "aws_elastictranscoder_pipeline_invalid_output_bucket"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsElastictranscoderPipelineInvalidOutputBucketRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsElastictranscoderPipelineInvalidOutputBucketRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsElastictranscoderPipelineInvalidOutputBucketRule) Link() string {
	return ""
}

// Check checks the pattern is valid
func (r *AwsElastictranscoderPipelineInvalidOutputBucketRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		var val string
		err := runner.EvaluateExpr(attribute.Expr, &val)

		return runner.EnsureNoError(err, func() error {
			if !r.pattern.MatchString(val) {
				runner.EmitIssue(
					r,
					fmt.Sprintf(`"%s" does not match valid pattern %s`, truncateLongMessage(val), `^(\w|\.|-){1,255}$`),
					attribute.Expr.Range(),
				)
			}
			return nil
		})
	})
}
