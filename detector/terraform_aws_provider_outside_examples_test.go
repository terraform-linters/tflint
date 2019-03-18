package detector

import (
	"testing"

	"reflect"

	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
)

func TestDetectTerraformProviderOutsideExamples(t *testing.T) {
	cases := []struct {
		Name   string
		Src    string
		Filename string
		Issues []*issue.Issue
	}{
		{
			Name: "aws provider in root dir",
			Src: `
provider "aws" {
}`,
			Filename: "main.tf",
			Issues: []*issue.Issue{
				{
					Detector: "terraform_aws_provider_outside_examples",
					Type:     "ERROR",
					Message:  "AWS Provider in non-example directory of remote module: main.tf Will probably result in resources being deployed inappropriately",
					Line:     3,
					File:     "test.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/terraform_module_pinned_source.md",
				},
			},
		},
		{
			Name: "aws provider in examples dir",
			Src: `
provider "aws" {
}`,
			Filename: "examples/main.tf",
			Issues: []*issue.Issue{},
		},
	}

	for _, tc := range cases {
		var issues = []*issue.Issue{}
		err := TestDetectByCreatorName(
			"CreateTerraformModulePinnedSourceDetector",
			tc.Src,
			"",
			config.Init(),
			config.Init().NewAwsClient(),
			&issues,
			tc.Filename,
		)
		if err != nil {
			t.Fatalf("\nERROR: %s", err)
		}

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(issues), pp.Sprint(tc.Issues), tc.Name)
		}
	}
}
