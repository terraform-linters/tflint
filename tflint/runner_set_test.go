package tflint

import (
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint/terraform"
	"github.com/zclconf/go-cty/cty"
)

func TestBuildRunners(t *testing.T) {
	for _, tc := range []struct {
		name              string
		files             map[string]string
		varfiles          []string
		wantModuleRunners int
		wantVariable      string
		wantVariableValue cty.Value
		wantAnnotated     string
	}{
		{
			name: "variable from auto-loaded varfile",
			files: map[string]string{
				"main.tf": `
variable "instance_type" {}

resource "aws_instance" "main" {
  instance_type = var.instance_type // tflint-ignore: aws_instance_invalid_type
}
`,
				"custom.auto.tfvars": `instance_type = "t2.micro"`,
			},
			wantModuleRunners: 0,
			wantVariable:      "instance_type",
			wantVariableValue: cty.StringVal("t2.micro"),
			wantAnnotated:     "main.tf",
		},
		{
			name: "variable from explicit varfile",
			files: map[string]string{
				"main.tf": `
variable "instance_type" {}
`,
				"custom.tfvars": `instance_type = "m5.xlarge"`,
			},
			varfiles:          []string{"custom.tfvars"},
			wantModuleRunners: 0,
			wantVariable:      "instance_type",
			wantVariableValue: cty.StringVal("m5.xlarge"),
		},
		{
			name: "with local module",
			files: map[string]string{
				"main.tf": `
variable "instance_type" {}

module "child" {
  source = "./module"
}
`,
				"module/main.tf": `
resource "aws_instance" "child" {}
`,
				"custom.auto.tfvars": `instance_type = "t2.large"`,
			},
			wantModuleRunners: 1,
			wantVariable:      "instance_type",
			wantVariableValue: cty.StringVal("t2.large"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			for name, src := range tc.files {
				if err := fs.WriteFile(name, []byte(src), os.ModePerm); err != nil {
					t.Fatal(err)
				}
			}

			wd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			loader, err := terraform.NewLoader(fs, wd)
			if err != nil {
				t.Fatal(err)
			}

			config := EmptyConfig()
			config.Varfiles = tc.varfiles

			runner, moduleRunners, err := BuildRunners(loader, config, wd, ".")
			if err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			if len(moduleRunners) != tc.wantModuleRunners {
				t.Errorf("Expected %d module runners, got %d", tc.wantModuleRunners, len(moduleRunners))
			}

			got := runner.Ctx.VariableValues[""][tc.wantVariable]
			if !got.RawEquals(tc.wantVariableValue) {
				t.Errorf("Expected variable %q to resolve to %#v, got %#v", tc.wantVariable, tc.wantVariableValue, got)
			}

			if tc.wantAnnotated != "" && len(runner.annotations[tc.wantAnnotated]) == 0 {
				t.Errorf("Expected annotations to be attached for %q, got none", tc.wantAnnotated)
			}
		})
	}
}
