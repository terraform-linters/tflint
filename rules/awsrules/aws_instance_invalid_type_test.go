package awsrules

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
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
}

resource "aws_instance" "missing_key" {
  ami = "ami-12345678"
}`,
			Expected: []*issue.Issue{
				{
					Detector: "aws_instance_invalid_type",
					Type:     issue.ERROR,
					Message:  "\"t1.2xlarge\" is invalid instance type.",
					Line:     3,
					File:     "instances.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_instance_invalid_type.md",
				},
			},
		},
	}

	dir, err := ioutil.TempDir("", "AwsInstanceInvalidType")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	loader, err := configload.NewLoader(&configload.Config{})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range cases {
		err := ioutil.WriteFile(dir+"/instances.tf", []byte(tc.Content), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		mod, diags := loader.Parser().LoadConfigDir(dir)
		if diags.HasErrors() {
			panic(diags)
		}
		cfg, tfdiags := configs.BuildConfig(mod, configs.DisabledModuleWalker)
		if tfdiags.HasErrors() {
			panic(tfdiags)
		}

		runner := tflint.NewRunner(cfg)
		rule := &AwsInstanceInvalidTypeRule{}
		rule.PreProcess()
		rule.Check(runner)

		if !cmp.Equal(tc.Expected, runner.Issues) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues))
		}
	}
}
