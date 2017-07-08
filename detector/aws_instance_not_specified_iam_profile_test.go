package detector

import (
	"reflect"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
)

func TestDetectAwsInstanceNotSpecifiedIAMProfile(t *testing.T) {
	cases := []struct {
		Name   string
		Src    string
		Issues []*issue.Issue
	}{
		{
			Name: "iam_instance_profile is not specified",
			Src: `
resource "aws_instance" "web" {
    instance_type = "t2.2xlarge"
}`,
			Issues: []*issue.Issue{
				{
					Detector: "aws_instance_not_specified_iam_profile",
					Type:     "NOTICE",
					Message:  "\"iam_instance_profile\" is not specified. If you want to change it, you need to recreate instance. (Only less than Terraform 0.8.8)",
					Line:     2,
					File:     "test.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_instance_not_specified_iam_profile.md",
				},
			},
		},
		{
			Name: "iam_instance_profile is specified",
			Src: `
resource "aws_instance" "web" {
    instance_type = "t2.micro"
    iam_instance_profile = "test"
}`,
			Issues: []*issue.Issue{},
		},
	}

	for _, tc := range cases {
		var issues = []*issue.Issue{}
		err := TestDetectByCreatorName(
			"CreateAwsInstanceNotSpecifiedIAMProfileDetector",
			tc.Src,
			"",
			config.Init(),
			config.Init().NewAwsClient(),
			&issues,
		)
		if err != nil {
			t.Fatalf("\nERROR: %s", err)
		}

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(issues), pp.Sprint(tc.Issues), tc.Name)
		}
	}
}
