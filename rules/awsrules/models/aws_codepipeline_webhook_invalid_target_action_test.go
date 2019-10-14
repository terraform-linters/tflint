// This file generated by `tools/model-rule-gen/main.go`. DO NOT EDIT

package models

import (
	"testing"

	"github.com/wata727/tflint/tflint"
)

func Test_AwsCodepipelineWebhookInvalidTargetActionRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "It includes invalid characters",
			Content: `
resource "aws_codepipeline_webhook" "foo" {
	target_action = "Source/Example"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsCodepipelineWebhookInvalidTargetActionRule(),
					Message: `target_action does not match valid pattern ^[A-Za-z0-9.@\-_]+$`,
				},
			},
		},
		{
			Name: "It is valid",
			Content: `
resource "aws_codepipeline_webhook" "foo" {
	target_action = "Source"
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewAwsCodepipelineWebhookInvalidTargetActionRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resource.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssuesWithoutRange(t, tc.Expected, runner.Issues())
	}
}
