package awsrules

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/hashicorp/terraform/terraform"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

func Test_AwsS3BucketInvalidACL(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected issue.Issues
	}{
		{
			Name: "no-access is invalid",
			Content: `
resource "aws_s3_bucket" "bucket" {
		acl = "no-access"
}
`,
			Expected: []*issue.Issue{
				{
					Detector: "aws_s3_bucket_invalid_acl",
					Type:     "ERROR",
					Message:  "\"no-access\" is invalid canned ACL type.",
					Line:     3,
					File:     "resource.tf",
				},
			},
		},
		{
			Name: "public-read is valid",
			Content: `
resource "aws_s3_bucket" "bucket" {
		acl = "public-read"
}
`,
			Expected: []*issue.Issue{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsS3BucketInvalidACL")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	err = os.Chdir(dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range cases {
		loader, err := configload.NewLoader(&configload.Config{})
		if err != nil {
			t.Fatal(err)
		}

		err = ioutil.WriteFile(dir+"/resource.tf", []byte(tc.Content), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		mod, diags := loader.Parser().LoadConfigDir(".")
		if diags.HasErrors() {
			t.Fatal(diags)
		}
		cfg, tfdiags := configs.BuildConfig(mod, configs.DisabledModuleWalker)
		if tfdiags.HasErrors() {
			t.Fatal(tfdiags)
		}

		runner := tflint.NewRunner(tflint.EmptyConfig(), map[string]tflint.Annotations{}, cfg, map[string]*terraform.InputValue{})
		rule := NewAwsS3BucketInvalidACLRule()

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		if !cmp.Equal(tc.Expected, runner.Issues) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues))
		}
	}
}
