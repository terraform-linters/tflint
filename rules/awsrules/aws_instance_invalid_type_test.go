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

func Test_AwsInstanceInvalidType(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected issue.Issues
	}{
		{
			Name: "basic",
			Content: `
resource "aws_instance" "invalid" {
  instance_type = "t1.2xlarge"
}

resource "aws_instance" "valid" {
  instance_type = "t2.micro"
}`,
			Expected: []*issue.Issue{
				{
					Detector: "aws_instance_invalid_type",
					Type:     issue.ERROR,
					Message:  "\"t1.2xlarge\" is invalid instance type.",
					Line:     3,
					File:     "instances.tf",
				},
			},
		},
	}

	dir, err := ioutil.TempDir("", "AwsInstanceInvalidType")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	err = os.Chdir(dir)
	if err != nil {
		t.Fatal(err)
	}

	loader, err := configload.NewLoader(&configload.Config{})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range cases {
		err := ioutil.WriteFile(dir+"/instances.tf", []byte(tc.Content), os.ModePerm)
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
		rule := NewAwsInstanceInvalidTypeRule()
		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		if !cmp.Equal(tc.Expected, runner.Issues) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues))
		}
	}
}
