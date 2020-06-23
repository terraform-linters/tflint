// This file generated by `generator/`. DO NOT EDIT

package models

import (
	"testing"

	"github.com/terraform-linters/tflint/tflint"
)

func Test_AwsEcsTaskDefinitionInvalidNetworkModeRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "It includes invalid characters",
			Content: `
resource "aws_ecs_task_definition" "foo" {
	network_mode = "vpc"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsEcsTaskDefinitionInvalidNetworkModeRule(),
					Message: `"vpc" is an invalid value as network_mode`,
				},
			},
		},
		{
			Name: "It is valid",
			Content: `
resource "aws_ecs_task_definition" "foo" {
	network_mode = "bridge"
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewAwsEcsTaskDefinitionInvalidNetworkModeRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resource.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssuesWithoutRange(t, tc.Expected, runner.Issues)
	}
}
