package awsrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_AwsS3BucketInvalidName(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: "Wanted tags: Bar,Foo, found: bar,foo",
			Content: `
resource "aws_s3_bucket" "foo" {
  bucket = "prod.foo.domain.com"
  acl    = "private"

  tags = {
    Name        = "foo"
    Environment = "prod"
  }
}

resource "aws_s3_bucket" "bar" {
	bucket = "bar.domain.com"
	acl    = "private"

	tags = {
	  Name        = "bar"
	}
  }`,
			Config: `
rule "aws_s3_bucket_name_match_regex" {
	enabled = true
	regex = "^prod.*"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsS3BucketNameMatchRegexRule(),
					Message: "Bucket name bar.domain.com does not match regex ^prod.*",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 13, Column: 11},
						End:      hcl.Pos{Line: 13, Column: 27},
					},
				},
			},
		},
	}

	rule := NewAwsS3BucketNameMatchRegexRule()

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"resource.tf": tc.Content}, loadConfigfromTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
