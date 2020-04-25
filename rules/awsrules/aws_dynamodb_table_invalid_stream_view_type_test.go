package awsrules

import (
	"testing"

	"github.com/terraform-linters/tflint/tflint"
)

func Test_AwsDynamoDBTableInvalidStreamViewTypeRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "It includes invalid characters",
			Content: `
resource "aws_dynamodb_table" "foo" {
	stream_view_type = "OLD_AND_NEW_IMAGE"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsDynamoDBTableInvalidStreamViewTypeRule(),
					Message: `"OLD_AND_NEW_IMAGE" is an invalid value as stream_view_type`,
				},
			},
		},
		{
			Name: "It is valid",
			Content: `
resource "aws_dynamodb_table" "foo" {
	stream_view_type = "NEW_IMAGE"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "empty string",
			Content: `
resource "aws_dynamodb_table" "foo" {
	stream_view_type = ""
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewAwsDynamoDBTableInvalidStreamViewTypeRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resource.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssuesWithoutRange(t, tc.Expected, runner.Issues)
	}
}
