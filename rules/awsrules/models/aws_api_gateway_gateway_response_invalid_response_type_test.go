// This file generated by `tools/model-rule-gen/main.go`. DO NOT EDIT

package models

import (
	"testing"

	"github.com/wata727/tflint/tflint"
)

func Test_AwsAPIGatewayGatewayResponseInvalidResponseTypeRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "It includes invalid characters",
			Content: `
resource "aws_api_gateway_gateway_response" "foo" {
	response_type = "4XX"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsAPIGatewayGatewayResponseInvalidResponseTypeRule(),
					Message: `response_type is not a valid value`,
				},
			},
		},
		{
			Name: "It is valid",
			Content: `
resource "aws_api_gateway_gateway_response" "foo" {
	response_type = "UNAUTHORIZED"
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewAwsAPIGatewayGatewayResponseInvalidResponseTypeRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resource.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssuesWithoutRange(t, tc.Expected, runner.Issues())
	}
}
