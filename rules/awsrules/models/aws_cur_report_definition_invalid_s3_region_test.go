// This file generated by `tools/model-rule-gen/main.go`. DO NOT EDIT

package models

import (
	"testing"

	"github.com/wata727/tflint/tflint"
)

func Test_AwsCurReportDefinitionInvalidS3RegionRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "It includes invalid characters",
			Content: `
resource "aws_cur_report_definition" "foo" {
	s3_region = "us-gov-east-1"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsCurReportDefinitionInvalidS3RegionRule(),
					Message: `s3_region is not a valid value`,
				},
			},
		},
		{
			Name: "It is valid",
			Content: `
resource "aws_cur_report_definition" "foo" {
	s3_region = "us-east-1"
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewAwsCurReportDefinitionInvalidS3RegionRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"resource.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssuesWithoutRange(t, tc.Expected, runner.Issues())
	}
}
