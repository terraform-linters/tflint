// This file generated by `tools/model-rule-gen/main.go`. DO NOT EDIT

package models

import (
	"testing"

	"github.com/wata727/tflint/tflint"
)

func Test_AwsCloudwatchLogStreamInvalidNameRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "It includes invalid characters",
			Content: `
resource "aws_cloudwatch_log_stream" "foo" {
	name = "Yoda:prod"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsCloudwatchLogStreamInvalidNameRule(),
					Message: `name does not match valid pattern ^[^:*]*$`,
				},
			},
		},
		{
			Name: "It is valid",
			Content: `
resource "aws_cloudwatch_log_stream" "foo" {
	name = "Yada"
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewAwsCloudwatchLogStreamInvalidNameRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resource.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssuesWithoutRange(t, tc.Expected, runner.Issues())
	}
}
