package terraformrules

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

func Test_TerraformDocumentedVariablesRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected issue.Issues
	}{
		{
			Name: "no description",
			Content: `
variable "no_description" {
  default = "default"
}`,
			Expected: []*issue.Issue{
				{
					Detector: "terraform_documented_variables",
					Type:     issue.NOTICE,
					Message:  "`no_description` variable has no description",
					Line:     2,
					File:     "variables.tf",
					Link:     project.ReferenceLink("terraform_documented_variables"),
				},
			},
		},
		{
			Name: "empty description",
			Content: `
variable "empty_description" {
  description = ""
}`,
			Expected: []*issue.Issue{
				{
					Detector: "terraform_documented_variables",
					Type:     issue.NOTICE,
					Message:  "`empty_description` variable has no description",
					Line:     2,
					File:     "variables.tf",
					Link:     project.ReferenceLink("terraform_documented_variables"),
				},
			},
		},
		{
			Name: "with description",
			Content: `
variable "with_description" {
  description = "This is description"
}`,
			Expected: []*issue.Issue{},
		},
	}

	dir, err := ioutil.TempDir("", "TerraformDocumentedVariables")
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

		err = ioutil.WriteFile(dir+"/variables.tf", []byte(tc.Content), os.ModePerm)
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
		rule := NewTerraformDocumentedVariablesRule()

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		if !cmp.Equal(tc.Expected, runner.Issues) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues))
		}
	}
}
