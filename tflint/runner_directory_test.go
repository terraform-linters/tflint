package tflint

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint/terraform"
	"github.com/zclconf/go-cty/cty"
)

func TestBuildRunnersExplicitDirectories(t *testing.T) {
	t.Setenv("TF_WORKSPACE", "")
	t.Setenv("TF_DATA_DIR", "")
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	for i := range 8 {
		name := fmt.Sprintf("root-%d", i)
		dir := filepath.Join(root, name)
		files := map[string]string{
			"main.tf": `
variable "from_default" {}
variable "from_auto" {}
variable "from_cli" {}
locals { value = file("value.txt") }
module "child" {
 source = "../shared"
 value = file("value.txt")
 count = length(fileset(".", "value.txt"))
}
module "remote" {
 source = "hashicorp/consul/aws"
 version = "0.0.1"
 value = templatefile("value.tmpl", { value = local.value })
}
`,
			"value.txt":                         name,
			"value.tmpl":                        "${value}",
			"terraform.tfvars":                  fmt.Sprintf("from_default = %q", name),
			"values.auto.tfvars":                fmt.Sprintf("from_auto = %q", name),
			"values.tfvars":                     fmt.Sprintf("from_cli = %q", name),
			".terraform/environment":            name,
			".terraform/modules/modules.json":   `{"Modules":[{"Key":"remote","Source":"hashicorp/consul/aws","Version":"0.0.1","Dir":".terraform/modules/remote"}]}`,
			".terraform/modules/remote/main.tf": `variable "value" {}`,
		}
		for path, content := range files {
			full := filepath.Join(dir, path)
			if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(full, []byte(content), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "shared"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "shared", "main.tf"), []byte(`variable "value" {}`), 0644); err != nil {
		t.Fatal(err)
	}
	t.Run("concurrent roots", func(t *testing.T) {
		for i := range 8 {
			name := fmt.Sprintf("root-%d", i)
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				dir := filepath.Join(root, name)
				loader, err := terraform.NewLoaderAt(afero.Afero{Fs: afero.NewOsFs()}, originalWd, dir)
				if err != nil {
					t.Fatal(err)
				}
				config := EmptyConfig()
				config.CallModuleType = terraform.CallAllModule
				config.Varfiles = []string{"values.tfvars"}
				runner, children, err := BuildRunners(loader, config, originalWd, ".")
				if err != nil {
					t.Fatal(err)
				}
				if len(children) != 2 {
					t.Fatalf("got %d children, want 2", len(children))
				}
				for expression, want := range map[string]string{
					`file("value.txt")`: name,
					`templatefile("value.tmpl", {value = local.value})`: name,
					`join(",", fileset(".", "value.txt"))`:              "value.txt",
					`abspath("value.txt")`:                              filepath.ToSlash(filepath.Join(dir, "value.txt")),
					`path.root`:                                         ".",
					`path.module`:                                       ".",
					`path.cwd`:                                          filepath.ToSlash(originalWd),
					`terraform.workspace`:                               name,
					`var.from_default`:                                  name,
					`var.from_auto`:                                     name,
					`var.from_cli`:                                      name,
				} {
					expr, diags := hclsyntax.ParseExpression([]byte(expression), "test.tf", hcl.InitialPos)
					if diags.HasErrors() {
						t.Fatal(diags)
					}
					value, diags := runner.Ctx.EvaluateExpr(expr, cty.String)
					if diags.HasErrors() {
						t.Fatalf("%s: %s", expression, diags)
					}
					if !value.RawEquals(cty.StringVal(want)) {
						t.Errorf("%s: got %s, want %s", expression, value.GoString(), strconv.Quote(want))
					}
				}
				for _, child := range children {
					expr, _ := hclsyntax.ParseExpression([]byte("var.value"), "test.tf", hcl.InitialPos)
					value, diags := child.Ctx.EvaluateExpr(expr, cty.String)
					if diags.HasErrors() || !value.RawEquals(cty.StringVal(name)) {
						t.Errorf("child value: %s, %s", value.GoString(), diags)
					}
				}
				filename, err := filepath.Rel(originalWd, filepath.Join(dir, "main.tf"))
				if err != nil {
					t.Fatal(err)
				}
				if _, ok := loader.Sources()[filename]; !ok {
					t.Errorf("missing diagnostic source %s", filename)
				}
			})
		}
	})
	wd, err := os.Getwd()
	if err != nil || wd != originalWd {
		t.Fatalf("working directory changed: %s, %v", wd, err)
	}
}
