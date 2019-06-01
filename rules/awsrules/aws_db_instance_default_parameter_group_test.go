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
	"github.com/wata727/tflint/project"
	"github.com/wata727/tflint/tflint"
)

func Test_AwsDBInstanceDefaultParameterGroup(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected issue.Issues
	}{
		{
			Name: "default.mysql5.6 is default parameter group",
			Content: `
resource "aws_db_instance" "db" {
    parameter_group_name = "default.mysql5.6"
}`,
			Expected: []*issue.Issue{
				{
					Detector: "aws_db_instance_default_parameter_group",
					Type:     "NOTICE",
					Message:  "\"default.mysql5.6\" is default parameter group. You cannot edit it.",
					Line:     3,
					File:     "resource.tf",
					Link:     project.ReferenceLink("aws_db_instance_default_parameter_group"),
				},
			},
		},
		{
			Name: "application5.6 is not default parameter group",
			Content: `
resource "aws_db_instance" "db" {
    parameter_group_name = "application5.6"
}`,
			Expected: []*issue.Issue{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsDBInstanceDefaultParameterGroup")
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
		rule := NewAwsDBInstanceDefaultParameterGroupRule()

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		if !cmp.Equal(tc.Expected, runner.Issues) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues))
		}
	}
}
