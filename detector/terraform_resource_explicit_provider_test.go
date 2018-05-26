package detector

import (
	"reflect"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
)

func TestDetectTerraformResourceExplicitProvider(t *testing.T) {
	cases := []struct {
		Name   string
		Src    string
		Issues []*issue.Issue
	}{
		{
			Name: "implicit provider",
			Src: `
resource "aws_instance" "web" {
  ami           = "ami-b73b63a0"
  instance_type = "t2.micro"

  tags {
    Name = "HelloWorld"
  }
}`,
			Issues: []*issue.Issue{
				{
					Detector: "terraform_resource_explicit_provider",
					Type:     "WARNING",
					Message:  "Resource \"web\" provider is implicit",
					Line:     2,
					File:     "test.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/terraform_resource_explicit_provider.md",
				},
			},
		},
		{
			Name: "explicit provider",
			Src: `
resource "aws_instance" "web" {
  provider = "aws.west"

  ami           = "ami-b73b63a0"
  instance_type = "t2.micro"

  tags {
    Name = "HelloWorld"
  }
}`,
			Issues: []*issue.Issue{},
		},
	}

	for _, tc := range cases {
		var issues = []*issue.Issue{}
		err := TestDetectByCreatorName(
			"CreateTerraformResourceExplicitProviderDetector",
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
