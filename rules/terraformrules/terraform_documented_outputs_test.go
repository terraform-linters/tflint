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

func Test_TerraformDocumentedOutputsRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected issue.Issues
	}{
		{
			Name: "no description",
			Content: `
output "endpoint" {
  value = aws_alb.main.dns_name
}`,
			Expected: []*issue.Issue{
				{
					Detector: "terraform_documented_outputs",
					Type:     issue.NOTICE,
					Message:  "`endpoint` output has no description",
					Line:     2,
					File:     "outputs.tf",
					Link:     project.ReferenceLink("terraform_documented_outputs"),
				},
			},
		},
		{
			Name: "empty description",
			Content: `
output "endpoint" {
  value = aws_alb.main.dns_name
  description = ""
}`,
			Expected: []*issue.Issue{
				{
					Detector: "terraform_documented_outputs",
					Type:     issue.NOTICE,
					Message:  "`endpoint` output has no description",
					Line:     2,
					File:     "outputs.tf",
					Link:     project.ReferenceLink("terraform_documented_outputs"),
				},
			},
		},
		{
			Name: "with description",
			Content: `
output "endpoint" {
  value = aws_alb.main.dns_name
  description = "DNS Endpoint"
}`,
			Expected: []*issue.Issue{},
		},
	}

	dir, err := ioutil.TempDir("", "TerraformDocumentedOutputs")
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

		err = ioutil.WriteFile(dir+"/outputs.tf", []byte(tc.Content), os.ModePerm)
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
		rule := NewTerraformDocumentedOutputsRule()

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		if !cmp.Equal(tc.Expected, runner.Issues) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues))
		}
	}
}
