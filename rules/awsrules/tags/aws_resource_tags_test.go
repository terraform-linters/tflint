package tags

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-linters/tflint/tflint"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

// Test_Tags tests that AWS resource tags are correctly detected
func Test_Tags(t *testing.T) {
	cases := []struct {
		Name       string
		ConfigFile string
		Content    string
		Expected   tflint.Issues
	}{
		{
			Name: "Incorrect tags are present",
			ConfigFile: `
config {
  tags = [ "foo", "bar" ]
}
`,
			Content: `
resource "aws_instance" "foo" {
	tags = {
    Foo = "bar",
    Bar = "baz"
  }
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsInstanceTagsRule(),
					Message: "Wanted tags: foo,bar, found: Bar,Foo\n",
				},
			},
		},
	}

	dir, err := ioutil.TempDir("", "_Test_Tags_")
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

	cfgFile := filepath.Join(dir, "tflint.hcl")

	for _, tc := range cases {
		err = ioutil.WriteFile(cfgFile, []byte(tc.ConfigFile), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		cfg, err := tflint.LoadConfig(cfgFile)
		if err != nil {
			t.Fatal(err)
		}

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

		tfCfg, tfdiags := configs.BuildConfig(mod, configs.DisabledModuleWalker)
		if tfdiags.HasErrors() {
			t.Fatal(tfdiags)
		}

		runner, err := tflint.NewRunner(cfg, map[string]tflint.Annotations{}, tfCfg, map[string]*terraform.InputValue{})
		if err != nil {
			t.Fatal(err)
		}

		rule := NewAwsInstanceTagsRule()

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		opts := []cmp.Option{
			cmpopts.IgnoreUnexported(AwsInstanceTagsRule{}),
			cmpopts.IgnoreFields(tflint.Issue{}, "Range"),
		}
		if !cmp.Equal(tc.Expected, runner.Issues, opts...) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues, opts...))
		}
	}
}
