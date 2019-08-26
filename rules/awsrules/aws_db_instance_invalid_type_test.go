package awsrules

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/hashicorp/terraform/terraform"
	"github.com/wata727/tflint/tflint"
)

func Test_AwsDBInstanceInvalidType(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "m4.2xlarge is invalid",
			Content: `
resource "aws_db_instance" "mysql" {
    instance_class = "m4.2xlarge"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsDBInstanceInvalidTypeRule(),
					Message: "\"m4.2xlarge\" is invalid instance type.",
					Range: hcl.Range{
						Filename: "resource.tf",
						Start:    hcl.Pos{Line: 3, Column: 22},
						End:      hcl.Pos{Line: 3, Column: 34},
					},
				},
			},
		},
		{
			Name: "db.m4.2xlarge is valid",
			Content: `
resource "aws_db_instance" "mysql" {
    instance_class = "db.m4.2xlarge"
}`,
			Expected: tflint.Issues{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsDBInstanceInvalidType")
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

		runner, err := tflint.NewRunner(tflint.EmptyConfig(), map[string]tflint.Annotations{}, cfg, map[string]*terraform.InputValue{})
		if err != nil {
			t.Fatal(err)
		}
		rule := NewAwsDBInstanceInvalidTypeRule()

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsDBInstanceInvalidTypeRule{}),
			cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}
