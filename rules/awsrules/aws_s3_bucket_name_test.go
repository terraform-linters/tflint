package awsrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_AwsS3BucketInvalidRegex(t *testing.T) {
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
  bucket = "blue.foo.domain.com"
  acl    = "private"
}

resource "aws_s3_bucket" "bar" {
	bucket = "bar.domain.com"
	acl    = "private"

	tags = {
	  Name        = "bar"
	}
  }`,
			Config: `
rule "aws_s3_bucket_name" {
	enabled = true
	regex = "^blue.*"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsS3BucketNameRule(),
					Message: "Bucket name bar.domain.com does not match regex ^blue.*",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 8, Column: 11},
						End:      hcl.Pos{Line: 8, Column: 27},
					},
				},
			},
		},
	}

	rule := NewAwsS3BucketNameRule()

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"resource.tf": tc.Content}, loadConfigfromTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

func Test_AwsS3BucketInvalidPrefix(t *testing.T) {
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
  bucket = "blue.foo.domain.com"
  acl    = "private"
}

resource "aws_s3_bucket" "bar" {
	bucket = "bar.domain.com"
	acl    = "private"

	tags = {
	  Name        = "bar"
	}
  }`,
			Config: `
rule "aws_s3_bucket_name" {
	enabled = true
	prefix = "blue"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsS3BucketNameRule(),
					Message: "Bucket name bar.domain.com does not have prefix blue",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 8, Column: 11},
						End:      hcl.Pos{Line: 8, Column: 27},
					},
				},
			},
		},
	}

	rule := NewAwsS3BucketNameRule()

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"resource.tf": tc.Content}, loadConfigfromTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
