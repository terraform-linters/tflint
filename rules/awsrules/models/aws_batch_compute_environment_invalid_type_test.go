// This file generated by `tools/model-rule-gen/main.go`. DO NOT EDIT

package models

import (
	"testing"

	"github.com/wata727/tflint/tflint"
)

func Test_AwsBatchComputeEnvironmentInvalidTypeRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "It includes invalid characters",
			Content: `
resource "aws_batch_compute_environment" "foo" {
	type = "CONTROLLED"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsBatchComputeEnvironmentInvalidTypeRule(),
					Message: `type is not a valid value`,
				},
			},
		},
		{
			Name: "It is valid",
			Content: `
resource "aws_batch_compute_environment" "foo" {
	type = "MANAGED"
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewAwsBatchComputeEnvironmentInvalidTypeRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resource.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssuesWithoutRange(t, tc.Expected, runner.Issues())
	}
}
