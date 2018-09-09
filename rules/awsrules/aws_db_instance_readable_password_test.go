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

func Test_AwsDBInstanceReadablePassword(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected issue.Issues
	}{
		{
			Name: "write password directly",
			Content: `
resource "aws_db_instance" "mysql" {
  password = "super_secret"
}`,
			Expected: []*issue.Issue{
				{
					Detector: "aws_db_instance_readable_password",
					Type:     issue.WARNING,
					Message:  "Password for the master DB user is readable. Recommend using environment variables or variable files.",
					Line:     3,
					File:     "resource.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_db_instance_readable_password.md",
				},
			},
		},
		{
			Name: "with default variable",
			Content: `
variable "password" {
  default = "super_secret"
}

resource "aws_db_instance" "mysql" {
  password = "${var.password}"
}`,
			Expected: []*issue.Issue{
				{
					Detector: "aws_db_instance_readable_password",
					Type:     issue.WARNING,
					Message:  "Password for the master DB user is readable. Recommend using environment variables or variable files.",
					Line:     7,
					File:     "resource.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_db_instance_readable_password.md",
				},
			},
		},
		{
			Name: "with no default variable",
			Content: `
variable "password" {}

resource "aws_db_instance" "mysql" {
  password = "${var.password}"
}`,
			Expected: []*issue.Issue{},
		},
		{
			Name: "with two variables, the one has default",
			Content: `
variable "head_password" {}
variable "tail_password" {
  default = "tails"
}

resource "aws_db_instance" "mysql" {
  password = "${var.head_password}-${var.tail_password}"
}`,
			Expected: []*issue.Issue{},
		},
		{
			Name: "with two variables, both has default",
			Content: `
variable "head_password" {
  default = "heads"
}
variable "tail_password" {
  default = "tails"
}

resource "aws_db_instance" "mysql" {
  password = "${var.head_password}-${var.tail_password}"
}`,
			Expected: []*issue.Issue{
				{
					Detector: "aws_db_instance_readable_password",
					Type:     issue.WARNING,
					Message:  "Password for the master DB user is readable. Recommend using environment variables or variable files.",
					Line:     10,
					File:     "resource.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_db_instance_readable_password.md",
				},
			},
		},
	}

	dir, err := ioutil.TempDir("", "AwsDBInstanceReadablePassword")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	for _, tc := range cases {
		loader, err := configload.NewLoader(&configload.Config{})
		if err != nil {
			t.Fatal(err)
		}

		err = ioutil.WriteFile(dir+"/resource.tf", []byte(tc.Content), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		mod, diags := loader.Parser().LoadConfigDir(dir)
		if diags.HasErrors() {
			t.Fatal(diags)
		}
		cfg, tfdiags := configs.BuildConfig(mod, configs.DisabledModuleWalker)
		if tfdiags.HasErrors() {
			t.Fatal(tfdiags)
		}

		runner := tflint.NewRunner(tflint.EmptyConfig(), cfg, map[string]*terraform.InputValue{})
		rule := NewAwsDBInstanceReadablePasswordRule()

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		if !cmp.Equal(tc.Expected, runner.Issues) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues))
		}
	}
}
