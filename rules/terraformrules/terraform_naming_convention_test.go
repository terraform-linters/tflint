package terraformrules

import (
	"fmt"
	"os"
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// Data blocks
func Test_TerraformNamingConventionRule_Data_DefaultEmpty(t *testing.T) {
	testDataSnakeCase(t, "default config", "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
}`)
}

func Test_TerraformNamingConventionRule_Data_DefaultFormat(t *testing.T) {
	testDataMixedSnakeCase(t, `default config (format="mixed_snake_case")`, `
rule "terraform_naming_convention" {
  enabled = true
  format  = "mixed_snake_case"
}`)
}

func Test_TerraformNamingConventionRule_Data_DefaultCustom(t *testing.T) {
	testDataSnakeCase(t, `default config (custom="^[a-z_]+$")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  custom  = "^[a-z][a-z]*(_[a-z]+)*$"
}`)
}

func Test_TerraformNamingConventionRule_Data_DefaultDisabled(t *testing.T) {
	testDataDisabled(t, `default config (format=null)`, `
rule "terraform_naming_convention" {
  enabled = true
  format  = "none"
}`)
}

func Test_TerraformNamingConventionRule_Data_DefaultFormat_OverrideFormat(t *testing.T) {
	testDataSnakeCase(t, `overridden config (format="snake_case")`, "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "mixed_snake_case"

  data {
    format = "snake_case"
  }
}`)
}

func Test_TerraformNamingConventionRule_Data_DefaultFormat_OverrideCustom(t *testing.T) {
	testDataSnakeCase(t, `overridden config (format="snake_case")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "mixed_snake_case"

  data {
    custom = "^[a-z][a-z]*(_[a-z]+)*$"
  }
}`)
}

func Test_TerraformNamingConventionRule_Data_DefaultCustom_OverrideFormat(t *testing.T) {
	testDataSnakeCase(t, `overridden config (format="snake_case")`, "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
  custom  = "^ignored$"

  data {
    format = "snake_case"
  }
}`)
}

func Test_TerraformNamingConventionRule_Data_DefaultCustom_OverrideCustom(t *testing.T) {
	testDataSnakeCase(t, `overridden config (format="snake_case")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  custom  = "^ignored$"

  data {
    custom = "^[a-z][a-z]*(_[a-z]+)*$"
  }
}`)
}

func Test_TerraformNamingConventionRule_Data_DefaultDisabled_OverrideFormat(t *testing.T) {
	testDataSnakeCase(t, `overridden config (format="snake_case")`, "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "none"

  data {
    format = "snake_case"
  }
}`)
}

func Test_TerraformNamingConventionRule_Data_DefaultDisabled_OverrideCustom(t *testing.T) {
	testDataSnakeCase(t, `overridden config (format="snake_case")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "none"

  data {
    custom = "^[a-z][a-z]*(_[a-z]+)*$"
  }
}`)
}

func Test_TerraformNamingConventionRule_Data_DefaultEmpty_OverrideDisabled(t *testing.T) {
	testDataDisabled(t, `overridden config (format=null)`, `
rule "terraform_naming_convention" {
  enabled = true

  data {
    format = "none"
  }
}`)
}

func Test_TerraformNamingConventionRule_Data_DefaultFormat_OverrideDisabled(t *testing.T) {
	testDataDisabled(t, `overridden config (format=null)`, `
rule "terraform_naming_convention" {
  enabled = true
  format  = "snake_case"

  data {
    format = "none"
  }
}`)
}

func Test_TerraformNamingConventionRule_Data_CustomFormats(t *testing.T) {
	testDataSnakeCase(t, `default config (custom_format="custom_snake_case")`, "format: Custom Snake Case", `
rule "terraform_naming_convention" {
  enabled = true
  format = "custom_snake_case"

  custom_formats = {
    custom_snake_case = {
      description = "Custom Snake Case"
      regex       = "^[a-z][a-z0-9]*(_[a-z0-9]+)*$"
    }
  }
}`)
}

func Test_TerraformNamingConventionRule_Data_CustomFormats_OverridePredefined(t *testing.T) {
	testDataSnakeCase(t, `default config (custom_format="snake_case")`, "format: Custom Snake Case", `
rule "terraform_naming_convention" {
  enabled = true
  format = "snake_case"

  custom_formats = {
    snake_case = {
      description = "Custom Snake Case"
      regex       = "^[a-z][a-z0-9]*(_[a-z0-9]+)*$"
    }
  }
}`)
}

func testDataSnakeCase(t *testing.T, testType string, formatName string, config string) {
	rule := NewTerraformNamingConventionRule()

	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: fmt.Sprintf("data: %s - Invalid snake_case with dash", testType),
			Content: `
data "aws_eip" "dash-name" {
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("data name `dash-name` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 27},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("data: %s - Invalid snake_case with camelCase", testType),
			Content: `
data "aws_eip" "camelCased" {
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("data name `camelCased` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 28},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("data: %s - Invalid snake_case with double underscore", testType),
			Content: `
data "aws_eip" "foo__bar" {
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("data name `foo__bar` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 26},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("data: %s - Invalid snake_case with underscore tail", testType),
			Content: `
data "aws_eip" "foo_bar_" {
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("data name `foo_bar_` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 26},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("data: %s - Invalid snake_case with Mixed_Snake_Case", testType),
			Content: `
data "aws_eip" "Foo_Bar" {
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("data name `Foo_Bar` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 25},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("data: %s - Invalid snake_case with count = 0", testType),
			Content: `
data "aws_eip" "camelCased" {
	count = 0
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("data name `camelCased` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 28},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("data: %s - Valid snake_case", testType),
			Content: `
data "aws_eip" "foo_bar" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("data: %s - Valid single word", testType),
			Content: `
data "aws_eip" "foo" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
	}

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"tests.tf": tc.Content}, loadConfigFromNamingConventionTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

func testDataMixedSnakeCase(t *testing.T, testType string, config string) {
	rule := NewTerraformNamingConventionRule()

	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: fmt.Sprintf("data: %s - Invalid mixed_snake_case with dash", testType),
			Content: `
data "aws_eip" "dash-name" {
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: "data name `dash-name` must match the following format: mixed_snake_case",
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 27},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("data: %s - Invalid mixed_snake_case with double underscore", testType),
			Content: `
data "aws_eip" "Foo__Bar" {
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: "data name `Foo__Bar` must match the following format: mixed_snake_case",
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 26},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("data: %s - Invalid mixed_snake_case with underscore tail", testType),
			Content: `
data "aws_eip" "Foo_Bar_" {
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: "data name `Foo_Bar_` must match the following format: mixed_snake_case",
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 26},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("data: %s - Valid snake_case", testType),
			Content: `
data "aws_eip" "foo_bar" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("data: %s - Valid single word", testType),
			Content: `
data "aws_eip" "foo" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("data: %s - Valid Mixed_Snake_Case", testType),
			Content: `
data "aws_eip" "Foo_Bar" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("data: %s - Valid single word with upper characters", testType),
			Content: `
data "aws_eip" "foo" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("data: %s - Valid PascalCase", testType),
			Content: `
data "aws_eip" "PascalCase" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("data: %s - Valid camelCase", testType),
			Content: `
data "aws_eip" "camelCase" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
	}

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"tests.tf": tc.Content}, loadConfigFromNamingConventionTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

func testDataDisabled(t *testing.T, testType string, config string) {
	rule := NewTerraformNamingConventionRule()

	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: fmt.Sprintf("data: %s - Valid mixed_snake_case with dash", testType),
			Content: `
data "aws_eip" "dash-name" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("data: %s - Valid snake_case", testType),
			Content: `
data "aws_eip" "foo_bar" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("data: %s - Valid single word", testType),
			Content: `
data "aws_eip" "foo" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("data: %s - Valid Mixed_Snake_Case", testType),
			Content: `
data "aws_eip" "Foo_Bar" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("data: %s - Valid single word upper characters", testType),
			Content: `
data "aws_eip" "Foo" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("data: %s - Valid PascalCase", testType),
			Content: `
data "aws_eip" "PascalCase" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("data: %s - Valid camelCase", testType),
			Content: `
data "aws_eip" "camelCase" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
	}

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"tests.tf": tc.Content}, loadConfigFromNamingConventionTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

// Local values
func Test_TerraformNamingConventionRule_Locals_DefaultEmpty(t *testing.T) {
	testLocalsSnakeCase(t, "default config", "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
}`)
}

func Test_TerraformNamingConventionRule_Locals_DefaultFormat(t *testing.T) {
	testLocalsMixedSnakeCase(t, `default config (format="mixed_snake_case")`, `
rule "terraform_naming_convention" {
  enabled = true
  format  = "mixed_snake_case"
}`)
}

func Test_TerraformNamingConventionRule_Locals_DefaultCustom(t *testing.T) {
	testLocalsSnakeCase(t, `default config (custom="^[a-z_]+$")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  custom  = "^[a-z][a-z]*(_[a-z]+)*$"
}`)
}

func Test_TerraformNamingConventionRule_Locals_DefaultDisabled(t *testing.T) {
	testLocalsDisabled(t, `default config (format=null)`, `
rule "terraform_naming_convention" {
  enabled = true
  format  = "none"
}`)
}

func Test_TerraformNamingConventionRule_Locals_DefaultFormat_OverrideFormat(t *testing.T) {
	testLocalsSnakeCase(t, `overridden config (format="snake_case")`, "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "mixed_snake_case"

  locals {
    format = "snake_case"
  }
}`)
}

func Test_TerraformNamingConventionRule_Locals_DefaultFormat_OverrideCustom(t *testing.T) {
	testLocalsSnakeCase(t, `overridden config (format="snake_case")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "mixed_snake_case"

  locals {
    custom = "^[a-z][a-z]*(_[a-z]+)*$"
  }
}`)
}

func Test_TerraformNamingConventionRule_Locals_DefaultCustom_OverrideFormat(t *testing.T) {
	testLocalsSnakeCase(t, `overridden config (format="snake_case")`, "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
  custom  = "^ignored$"

  locals {
    format = "snake_case"
  }
}`)
}

func Test_TerraformNamingConventionRule_Locals_DefaultCustom_OverrideCustom(t *testing.T) {
	testLocalsSnakeCase(t, `overridden config (format="snake_case")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  custom  = "^ignored$"

  locals {
    custom = "^[a-z][a-z]*(_[a-z]+)*$"
  }
}`)
}

func Test_TerraformNamingConventionRule_Locals_DefaultDisabled_OverrideFormat(t *testing.T) {
	testLocalsSnakeCase(t, `overridden config (format="snake_case")`, "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "none"

  locals {
    format = "snake_case"
  }
}`)
}

func Test_TerraformNamingConventionRule_Locals_DefaultDisabled_OverrideCustom(t *testing.T) {
	testLocalsSnakeCase(t, `overridden config (format="snake_case")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "none"

  locals {
    custom = "^[a-z][a-z]*(_[a-z]+)*$"
  }
}`)
}

func Test_TerraformNamingConventionRule_Locals_DefaultEmpty_OverrideDisabled(t *testing.T) {
	testLocalsDisabled(t, `overridden config (format=null)`, `
rule "terraform_naming_convention" {
  enabled = true

  locals {
    format = "none"
  }
}`)
}

func Test_TerraformNamingConventionRule_Locals_DefaultFormat_OverrideDisabled(t *testing.T) {
	testLocalsDisabled(t, `overridden config (format=null)`, `
rule "terraform_naming_convention" {
  enabled = true
  format  = "snake_case"

  locals {
    format = "none"
  }
}`)
}

func testLocalsSnakeCase(t *testing.T, testType string, formatName string, config string) {
	rule := NewTerraformNamingConventionRule()

	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: fmt.Sprintf("locals: %s - Invalid snake_case with dash", testType),
			Content: `
locals {
  dash-name = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("local value name `dash-name` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 24},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("locals: %s - Invalid snake_case with camelCase", testType),
			Content: `
locals {
  camelCased = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("local value name `camelCased` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 25},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("locals: %s - Invalid snake_case with double underscore", testType),
			Content: `
locals {
  foo__bar = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("local value name `foo__bar` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 23},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("locals: %s - Invalid snake_case with underscore tail", testType),
			Content: `
locals {
  foo_bar_ = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("local value name `foo_bar_` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 23},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("locals: %s - Invalid snake_case with Mixed_Snake_Case", testType),
			Content: `
locals {
  Foo_Bar = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("local value name `Foo_Bar` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 22},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("locals: %s - Valid snake_case", testType),
			Content: `
locals {
  foo_bar = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("locals: %s - Valid single word", testType),
			Content: `
locals {
  foo = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
	}

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"tests.tf": tc.Content}, loadConfigFromNamingConventionTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

func testLocalsMixedSnakeCase(t *testing.T, testType string, config string) {
	rule := NewTerraformNamingConventionRule()

	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: fmt.Sprintf("locals: %s - Invalid mixed_snake_case with dash", testType),
			Content: `
locals {
  dash-name = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: "local value name `dash-name` must match the following format: mixed_snake_case",
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 24},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("locals: %s - Invalid mixed_snake_case with double underscore", testType),
			Content: `
locals {
  Foo__Bar = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: "local value name `Foo__Bar` must match the following format: mixed_snake_case",
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 23},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("locals: %s - Invalid mixed_snake_case with underscore tail", testType),
			Content: `
locals {
  Foo_Bar_ = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: "local value name `Foo_Bar_` must match the following format: mixed_snake_case",
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 23},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("locals: %s - Valid snake_case", testType),
			Content: `
locals {
  foo_bar = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("locals: %s - Valid single word", testType),
			Content: `
locals {
  foo = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("locals: %s - Valid Mixed_Snake_Case", testType),
			Content: `
locals {
  Foo_Bar = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("locals: %s - Valid single word with upper characters", testType),
			Content: `
locals {
  Foo = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("locals: %s - Valid PascalCase", testType),
			Content: `
locals {
  PascalCase = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("locals: %s - Valid camelCase", testType),
			Content: `
locals {
  camelCase = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
	}

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"tests.tf": tc.Content}, loadConfigFromNamingConventionTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

func testLocalsDisabled(t *testing.T, testType string, config string) {
	rule := NewTerraformNamingConventionRule()

	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: fmt.Sprintf("locals: %s - Valid mixed_snake_case with dash", testType),
			Content: `
locals {
  dash-name = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("locals: %s - Valid snake_case", testType),
			Content: `
locals {
  foo_bar = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("locals: %s - Valid single word", testType),
			Content: `
locals {
  foo = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("locals: %s - Valid Mixed_Snake_Case", testType),
			Content: `
locals {
  Foo_Bar = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("locals: %s - Valid single word with upper characters", testType),
			Content: `
locals {
  Foo = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("locals: %s - Valid PascalCase", testType),
			Content: `
locals {
  PascalCase = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("locals: %s - Valid camelCase", testType),
			Content: `
locals {
  camelCase = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
	}

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"tests.tf": tc.Content}, loadConfigFromNamingConventionTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

// Module blocks
func Test_TerraformNamingConventionRule_Module_DefaultEmpty(t *testing.T) {
	testModuleSnakeCase(t, "default config", "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
}`)
}

func Test_TerraformNamingConventionRule_Module_DefaultFormat(t *testing.T) {
	testModuleMixedSnakeCase(t, `default config (format="mixed_snake_case")`, `
rule "terraform_naming_convention" {
  enabled = true
  format  = "mixed_snake_case"
}`)
}

func Test_TerraformNamingConventionRule_Module_DefaultCustom(t *testing.T) {
	testModuleSnakeCase(t, `default config (custom="^[a-z_]+$")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  custom  = "^[a-z][a-z]*(_[a-z]+)*$"
}`)
}

func Test_TerraformNamingConventionRule_Module_DefaultDisabled(t *testing.T) {
	testModuleDisabled(t, `default config (format=null)`, `
rule "terraform_naming_convention" {
  enabled = true
  format  = "none"
}`)
}

func Test_TerraformNamingConventionRule_Module_DefaultFormat_OverrideFormat(t *testing.T) {
	testModuleSnakeCase(t, `overridden config (format="snake_case")`, "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "mixed_snake_case"

  module {
    format = "snake_case"
  }
}`)
}

func Test_TerraformNamingConventionRule_Module_DefaultFormat_OverrideCustom(t *testing.T) {
	testModuleSnakeCase(t, `overridden config (format="snake_case")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "mixed_snake_case"

  module {
    custom = "^[a-z][a-z]*(_[a-z]+)*$"
  }
}`)
}

func Test_TerraformNamingConventionRule_Module_DefaultCustom_OverrideFormat(t *testing.T) {
	testModuleSnakeCase(t, `overridden config (format="snake_case")`, "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
  custom  = "^ignored$"

  module {
    format = "snake_case"
  }
}`)
}

func Test_TerraformNamingConventionRule_Module_DefaultCustom_OverrideCustom(t *testing.T) {
	testModuleSnakeCase(t, `overridden config (format="snake_case")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  custom  = "^ignored$"

  module {
    custom = "^[a-z][a-z]*(_[a-z]+)*$"
  }
}`)
}

func Test_TerraformNamingConventionRule_Module_DefaultDisabled_OverrideFormat(t *testing.T) {
	testModuleSnakeCase(t, `overridden config (format="snake_case")`, "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "none"

  module {
    format = "snake_case"
  }
}`)
}

func Test_TerraformNamingConventionRule_Module_DefaultDisabled_OverrideCustom(t *testing.T) {
	testModuleSnakeCase(t, `overridden config (format="snake_case")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "none"

  module {
    custom = "^[a-z][a-z]*(_[a-z]+)*$"
  }
}`)
}

func Test_TerraformNamingConventionRule_Module_DefaultEmpty_OverrideDisabled(t *testing.T) {
	testModuleDisabled(t, `overridden config (format=null)`, `
rule "terraform_naming_convention" {
  enabled = true

  module {
    format = "none"
  }
}`)
}

func Test_TerraformNamingConventionRule_Module_DefaultFormat_OverrideDisabled(t *testing.T) {
	testModuleDisabled(t, `overridden config (format=null)`, `
rule "terraform_naming_convention" {
  enabled = true
  format  = "snake_case"

  module {
    format = "none"
  }
}`)
}

func testModuleSnakeCase(t *testing.T, testType string, formatName string, config string) {
	rule := NewTerraformNamingConventionRule()

	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: fmt.Sprintf("module: %s - Invalid snake_case with dash", testType),
			Content: `
module "dash-name" {
  source = "./module"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("module name `dash-name` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 19},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("module: %s - Invalid snake_case with camelCase", testType),
			Content: `
module "camelCased" {
  source = "./module"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("module name `camelCased` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 20},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("module: %s - Invalid snake_case with double underscore", testType),
			Content: `
module "foo__bar" {
  source = "./module"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("module name `foo__bar` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 18},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("module: %s - Invalid snake_case with underscore tail", testType),
			Content: `
module "foo_bar_" {
  source = "./module"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("module name `foo_bar_` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 18},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("module: %s - Invalid snake_case with Mixed_Snake_Case", testType),
			Content: `
module "Foo_Bar" {
  source = "./module"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("module name `Foo_Bar` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 17},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("module: %s - Valid snake_case", testType),
			Content: `
module "foo_bar" {
  source = "./module"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("module: %s - Valid single word", testType),
			Content: `
module "foo" {
  source = "./module"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
	}

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"tests.tf": tc.Content}, loadConfigFromNamingConventionTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

func testModuleMixedSnakeCase(t *testing.T, testType string, config string) {
	rule := NewTerraformNamingConventionRule()

	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: fmt.Sprintf("module: %s - Invalid mixed_snake_case with dash", testType),
			Content: `
module "dash-name" {
  source = "./module"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: "module name `dash-name` must match the following format: mixed_snake_case",
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 19},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("module: %s - Invalid mixed_snake_case with double underscore", testType),
			Content: `
module "Foo__Bar" {
  source = "./module"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: "module name `Foo__Bar` must match the following format: mixed_snake_case",
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 18},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("module: %s - Invalid mixed_snake_case with underscore tail", testType),
			Content: `
module "Foo_Bar_" {
  source = "./module"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: "module name `Foo_Bar_` must match the following format: mixed_snake_case",
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 18},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("module: %s - Valid snake_case", testType),
			Content: `
module "foo_bar" {
  source = "./module"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("module: %s - Valid single word", testType),
			Content: `
module "foo" {
  source = "./module"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("module: %s - Valid Mixed_Snake_Case", testType),
			Content: `
module "Foo_Bar" {
  source = "./module"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("module: %s - Valid single word with upper characters", testType),
			Content: `
module "foo" {
  source = "./module"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("module: %s - Valid PascalCase", testType),
			Content: `
module "PascalCase" {
  source = "./module"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("module: %s - Valid camelCase", testType),
			Content: `
module "camelCase" {
  source = "./module"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
	}

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"tests.tf": tc.Content}, loadConfigFromNamingConventionTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

func testModuleDisabled(t *testing.T, testType string, config string) {
	rule := NewTerraformNamingConventionRule()

	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: fmt.Sprintf("module: %s - Valid mixed_snake_case with dash", testType),
			Content: `
module "dash-name" {
  source = "./module"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("module: %s - Valid snake_case", testType),
			Content: `
module "foo_bar" {
  source = "./module"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("module: %s - Valid single word", testType),
			Content: `
module "foo" {
  source = "./module"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("module: %s - Valid Mixed_Snake_Case", testType),
			Content: `
module "Foo_Bar" {
  source = "./module"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("module: %s - Valid single word upper characters", testType),
			Content: `
module "Foo" {
  source = "./module"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("module: %s - Valid PascalCase", testType),
			Content: `
module "PascalCase" {
  source = "./module"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("module: %s - Valid camelCase", testType),
			Content: `
module "camelCase" {
  source = "./module"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
	}

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"tests.tf": tc.Content}, loadConfigFromNamingConventionTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

// Output blocks
func Test_TerraformNamingConventionRule_Output_DefaultEmpty(t *testing.T) {
	testOutputSnakeCase(t, "default config", "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
}`)
}

func Test_TerraformNamingConventionRule_Output_DefaultFormat(t *testing.T) {
	testOutputMixedSnakeCase(t, `default config (format="mixed_snake_case")`, `
rule "terraform_naming_convention" {
  enabled = true
  format  = "mixed_snake_case"
}`)
}

func Test_TerraformNamingConventionRule_Output_DefaultCustom(t *testing.T) {
	testOutputSnakeCase(t, `default config (custom="^[a-z_]+$")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  custom  = "^[a-z][a-z]*(_[a-z]+)*$"
}`)
}

func Test_TerraformNamingConventionRule_Output_DefaultDisabled(t *testing.T) {
	testOutputDisabled(t, `default config (format=null)`, `
rule "terraform_naming_convention" {
  enabled = true
  format  = "none"
}`)
}

func Test_TerraformNamingConventionRule_Output_DefaultFormat_OverrideFormat(t *testing.T) {
	testOutputSnakeCase(t, `overridden config (format="snake_case")`, "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "mixed_snake_case"

  output {
    format = "snake_case"
  }
}`)
}

func Test_TerraformNamingConventionRule_Output_DefaultFormat_OverrideCustom(t *testing.T) {
	testOutputSnakeCase(t, `overridden config (format="snake_case")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "mixed_snake_case"

  output {
    custom = "^[a-z][a-z]*(_[a-z]+)*$"
  }
}`)
}

func Test_TerraformNamingConventionRule_Output_DefaultCustom_OverrideFormat(t *testing.T) {
	testOutputSnakeCase(t, `overridden config (format="snake_case")`, "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
  custom  = "^ignored$"

  output {
    format = "snake_case"
  }
}`)
}

func Test_TerraformNamingConventionRule_Output_DefaultCustom_OverrideCustom(t *testing.T) {
	testOutputSnakeCase(t, `overridden config (format="snake_case")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  custom  = "^ignored$"

  output {
    custom = "^[a-z][a-z]*(_[a-z]+)*$"
  }
}`)
}

func Test_TerraformNamingConventionRule_Output_DefaultDisabled_OverrideFormat(t *testing.T) {
	testOutputSnakeCase(t, `overridden config (format="snake_case")`, "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "none"

  output {
    format = "snake_case"
  }
}`)
}

func Test_TerraformNamingConventionRule_Output_DefaultDisabled_OverrideCustom(t *testing.T) {
	testOutputSnakeCase(t, `overridden config (format="snake_case")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "none"

  output {
    custom = "^[a-z][a-z]*(_[a-z]+)*$"
  }
}`)
}

func Test_TerraformNamingConventionRule_Output_DefaultEmpty_OverrideDisabled(t *testing.T) {
	testOutputDisabled(t, `overridden config (format=null)`, `
rule "terraform_naming_convention" {
  enabled = true

  output {
    format = "none"
  }
}`)
}

func Test_TerraformNamingConventionRule_Output_DefaultFormat_OverrideDisabled(t *testing.T) {
	testOutputDisabled(t, `overridden config (format=null)`, `
rule "terraform_naming_convention" {
  enabled = true
  format  = "snake_case"

  output {
    format = "none"
  }
}`)
}

func testOutputSnakeCase(t *testing.T, testType string, formatName string, config string) {
	rule := NewTerraformNamingConventionRule()

	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: fmt.Sprintf("output: %s - Invalid snake_case with dash", testType),
			Content: `
output "dash-name" {
  value = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("output name `dash-name` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 19},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("output: %s - Invalid snake_case with camelCase", testType),
			Content: `
output "camelCased" {
  value = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("output name `camelCased` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 20},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("output: %s - Invalid snake_case with double underscore", testType),
			Content: `
output "foo__bar" {
  value = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("output name `foo__bar` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 18},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("output: %s - Invalid snake_case with underscore tail", testType),
			Content: `
output "foo_bar_" {
  value = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("output name `foo_bar_` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 18},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("output: %s - Invalid snake_case with Mixed_Snake_Case", testType),
			Content: `
output "Foo_Bar" {
  value = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("output name `Foo_Bar` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 17},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("output: %s - Valid snake_case", testType),
			Content: `
output "foo_bar" {
  value = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("output: %s - Valid single word", testType),
			Content: `
output "foo" {
  value = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
	}

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"tests.tf": tc.Content}, loadConfigFromNamingConventionTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

func testOutputMixedSnakeCase(t *testing.T, testType string, config string) {
	rule := NewTerraformNamingConventionRule()

	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: fmt.Sprintf("output: %s - Invalid mixed_snake_case with dash", testType),
			Content: `
output "dash-name" {
  value = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: "output name `dash-name` must match the following format: mixed_snake_case",
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 19},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("output: %s - Invalid mixed_snake_case with double underscore", testType),
			Content: `
output "Foo__Bar" {
  value = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: "output name `Foo__Bar` must match the following format: mixed_snake_case",
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 18},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("output: %s - Invalid mixed_snake_case with underscore tail", testType),
			Content: `
output "Foo_Bar_" {
  value = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: "output name `Foo_Bar_` must match the following format: mixed_snake_case",
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 18},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("output: %s - Valid snake_case", testType),
			Content: `
output "foo_bar" {
  value = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("output: %s - Valid single word", testType),
			Content: `
output "foo" {
  value = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("output: %s - Valid Mixed_Snake_Case", testType),
			Content: `
output "Foo_Bar" {
  value = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("output: %s - Valid single word with upper characters", testType),
			Content: `
output "foo" {
  value = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("output: %s - Valid PascalCase", testType),
			Content: `
output "PascalCase" {
  value = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("output: %s - Valid camelCase", testType),
			Content: `
output "camelCase" {
  value = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
	}

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"tests.tf": tc.Content}, loadConfigFromNamingConventionTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

func testOutputDisabled(t *testing.T, testType string, config string) {
	rule := NewTerraformNamingConventionRule()

	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: fmt.Sprintf("output: %s - Valid mixed_snake_case with dash", testType),
			Content: `
output "dash-name" {
  value = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("output: %s - Valid snake_case", testType),
			Content: `
output "foo_bar" {
  value = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("output: %s - Valid single word", testType),
			Content: `
output "foo" {
  value = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("output: %s - Valid Mixed_Snake_Case", testType),
			Content: `
output "Foo_Bar" {
  value = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("output: %s - Valid single word upper characters", testType),
			Content: `
output "Foo" {
  value = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("output: %s - Valid PascalCase", testType),
			Content: `
output "PascalCase" {
  value = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("output: %s - Valid camelCase", testType),
			Content: `
output "camelCase" {
  value = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
	}

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"tests.tf": tc.Content}, loadConfigFromNamingConventionTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

// Resource blocks
func Test_TerraformNamingConventionRule_Resource_DefaultEmpty(t *testing.T) {
	testResourceSnakeCase(t, "default config", "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
}`)
}

func Test_TerraformNamingConventionRule_Resource_DefaultFormat(t *testing.T) {
	testResourceMixedSnakeCase(t, `default config (format="mixed_snake_case")`, `
rule "terraform_naming_convention" {
  enabled = true
  format  = "mixed_snake_case"
}`)
}

func Test_TerraformNamingConventionRule_Resource_DefaultCustom(t *testing.T) {
	testResourceSnakeCase(t, `default config (custom="^[a-z_]+$")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  custom  = "^[a-z][a-z]*(_[a-z]+)*$"
}`)
}

func Test_TerraformNamingConventionRule_Resource_DefaultDisabled(t *testing.T) {
	testResourceDisabled(t, `default config (format=null)`, `
rule "terraform_naming_convention" {
  enabled = true
  format  = "none"
}`)
}

func Test_TerraformNamingConventionRule_Resource_DefaultFormat_OverrideFormat(t *testing.T) {
	testResourceSnakeCase(t, `overridden config (format="snake_case")`, "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "mixed_snake_case"

  resource {
    format = "snake_case"
  }
}`)
}

func Test_TerraformNamingConventionRule_Resource_DefaultFormat_OverrideCustom(t *testing.T) {
	testResourceSnakeCase(t, `overridden config (format="snake_case")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "mixed_snake_case"

  resource {
    custom = "^[a-z][a-z]*(_[a-z]+)*$"
  }
}`)
}

func Test_TerraformNamingConventionRule_Resource_DefaultCustom_OverrideFormat(t *testing.T) {
	testResourceSnakeCase(t, `overridden config (format="snake_case")`, "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
  custom  = "^ignored$"

  resource {
    format = "snake_case"
  }
}`)
}

func Test_TerraformNamingConventionRule_Resource_DefaultCustom_OverrideCustom(t *testing.T) {
	testResourceSnakeCase(t, `overridden config (format="snake_case")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  custom  = "^ignored$"

  resource {
    custom = "^[a-z][a-z]*(_[a-z]+)*$"
  }
}`)
}

func Test_TerraformNamingConventionRule_Resource_DefaultDisabled_OverrideFormat(t *testing.T) {
	testResourceSnakeCase(t, `overridden config (format="snake_case")`, "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "none"

  resource {
    format = "snake_case"
  }
}`)
}

func Test_TerraformNamingConventionRule_Resource_DefaultDisabled_OverrideCustom(t *testing.T) {
	testResourceSnakeCase(t, `overridden config (format="snake_case")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "none"

  resource {
    custom = "^[a-z][a-z]*(_[a-z]+)*$"
  }
}`)
}

func Test_TerraformNamingConventionRule_Resource_DefaultEmpty_OverrideDisabled(t *testing.T) {
	testResourceDisabled(t, `overridden config (format=null)`, `
rule "terraform_naming_convention" {
  enabled = true

  resource {
    format = "none"
  }
}`)
}

func Test_TerraformNamingConventionRule_Resource_DefaultFormat_OverrideDisabled(t *testing.T) {
	testResourceDisabled(t, `overridden config (format=null)`, `
rule "terraform_naming_convention" {
  enabled = true
  format  = "snake_case"

  resource {
    format = "none"
  }
}`)
}

func testResourceSnakeCase(t *testing.T, testType string, formatName string, config string) {
	rule := NewTerraformNamingConventionRule()

	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: fmt.Sprintf("resource: %s - Invalid snake_case with dash", testType),
			Content: `
resource "aws_eip" "dash-name" {
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("resource name `dash-name` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 31},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("resource: %s - Invalid snake_case with camelCase", testType),
			Content: `
resource "aws_eip" "camelCased" {
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("resource name `camelCased` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 32},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("resource: %s - Invalid snake_case with double underscore", testType),
			Content: `
resource "aws_eip" "foo__bar" {
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("resource name `foo__bar` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 30},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("resource: %s - Invalid snake_case with underscore tail", testType),
			Content: `
resource "aws_eip" "foo_bar_" {
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("resource name `foo_bar_` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 30},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("resource: %s - Invalid snake_case with Mixed_Snake_Case", testType),
			Content: `
resource "aws_eip" "Foo_Bar" {
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("resource name `Foo_Bar` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 29},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("resource: %s - Valid snake_case", testType),
			Content: `
resource "aws_eip" "foo_bar" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("resource: %s - Valid single word", testType),
			Content: `
resource "aws_eip" "foo" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
	}

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"tests.tf": tc.Content}, loadConfigFromNamingConventionTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

func testResourceMixedSnakeCase(t *testing.T, testType string, config string) {
	rule := NewTerraformNamingConventionRule()

	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: fmt.Sprintf("resource: %s - Invalid mixed_snake_case with dash", testType),
			Content: `
resource "aws_eip" "dash-name" {
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: "resource name `dash-name` must match the following format: mixed_snake_case",
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 31},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("resource: %s - Invalid mixed_snake_case with double underscore", testType),
			Content: `
resource "aws_eip" "Foo__Bar" {
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: "resource name `Foo__Bar` must match the following format: mixed_snake_case",
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 30},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("resource: %s - Invalid mixed_snake_case with underscore tail", testType),
			Content: `
resource "aws_eip" "Foo_Bar_" {
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: "resource name `Foo_Bar_` must match the following format: mixed_snake_case",
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 30},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("resource: %s - Valid snake_case", testType),
			Content: `
resource "aws_eip" "foo_bar" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("resource: %s - Valid single word", testType),
			Content: `
resource "aws_eip" "foo" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("resource: %s - Valid Mixed_Snake_Case", testType),
			Content: `
resource "aws_eip" "Foo_Bar" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("resource: %s - Valid single word with upper characters", testType),
			Content: `
resource "aws_eip" "foo" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("resource: %s - Valid PascalCase", testType),
			Content: `
resource "aws_eip" "PascalCase" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("resource: %s - Valid camelCase", testType),
			Content: `
resource "aws_eip" "camelCase" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
	}

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"tests.tf": tc.Content}, loadConfigFromNamingConventionTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

func testResourceDisabled(t *testing.T, testType string, config string) {
	rule := NewTerraformNamingConventionRule()

	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: fmt.Sprintf("resource: %s - Valid mixed_snake_case with dash", testType),
			Content: `
resource "aws_eip" "dash-name" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("resource: %s - Valid snake_case", testType),
			Content: `
resource "aws_eip" "foo_bar" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("resource: %s - Valid single word", testType),
			Content: `
resource "aws_eip" "foo" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("resource: %s - Valid Mixed_Snake_Case", testType),
			Content: `
resource "aws_eip" "Foo_Bar" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("resource: %s - Valid single word upper characters", testType),
			Content: `
resource "aws_eip" "Foo" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("resource: %s - Valid PascalCase", testType),
			Content: `
resource "aws_eip" "PascalCase" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("resource: %s - Valid camelCase", testType),
			Content: `
resource "aws_eip" "camelCase" {
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
	}

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"tests.tf": tc.Content}, loadConfigFromNamingConventionTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

// Variable blocks
func Test_TerraformNamingConventionRule_Variable_DefaultEmpty(t *testing.T) {
	testVariableSnakeCase(t, "default config", "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
}`)
}

func Test_TerraformNamingConventionRule_Variable_DefaultFormat(t *testing.T) {
	testVariableMixedSnakeCase(t, `default config (format="mixed_snake_case")`, `
rule "terraform_naming_convention" {
  enabled = true
  format  = "mixed_snake_case"
}`)
}

func Test_TerraformNamingConventionRule_Variable_DefaultCustom(t *testing.T) {
	testVariableSnakeCase(t, `default config (custom="^[a-z_]+$")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  custom  = "^[a-z][a-z]*(_[a-z]+)*$"
}`)
}

func Test_TerraformNamingConventionRule_Variable_DefaultDisabled(t *testing.T) {
	testVariableDisabled(t, `default config (format=null)`, `
rule "terraform_naming_convention" {
  enabled = true
  format  = "none"
}`)
}

func Test_TerraformNamingConventionRule_Variable_DefaultFormat_OverrideFormat(t *testing.T) {
	testVariableSnakeCase(t, `overridden config (format="snake_case")`, "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "mixed_snake_case"

  variable {
    format = "snake_case"
  }
}`)
}

func Test_TerraformNamingConventionRule_Variable_DefaultFormat_OverrideCustom(t *testing.T) {
	testVariableSnakeCase(t, `overridden config (format="snake_case")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "mixed_snake_case"

  variable {
    custom = "^[a-z][a-z]*(_[a-z]+)*$"
  }
}`)
}

func Test_TerraformNamingConventionRule_Variable_DefaultCustom_OverrideFormat(t *testing.T) {
	testVariableSnakeCase(t, `overridden config (format="snake_case")`, "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
  custom  = "^ignored$"

  variable {
    format = "snake_case"
  }
}`)
}

func Test_TerraformNamingConventionRule_Variable_DefaultCustom_OverrideCustom(t *testing.T) {
	testVariableSnakeCase(t, `overridden config (format="snake_case")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  custom  = "^ignored$"

  variable {
    custom = "^[a-z][a-z]*(_[a-z]+)*$"
  }
}`)
}

func Test_TerraformNamingConventionRule_Variable_DefaultDisabled_OverrideFormat(t *testing.T) {
	testVariableSnakeCase(t, `overridden config (format="snake_case")`, "format: snake_case", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "none"

  variable {
    format = "snake_case"
  }
}`)
}

func Test_TerraformNamingConventionRule_Variable_DefaultDisabled_OverrideCustom(t *testing.T) {
	testVariableSnakeCase(t, `overridden config (format="snake_case")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  format  = "none"

  variable {
    custom = "^[a-z][a-z]*(_[a-z]+)*$"
  }
}`)
}

func Test_TerraformNamingConventionRule_Variable_DefaultEmpty_OverrideDisabled(t *testing.T) {
	testVariableDisabled(t, `overridden config (format=null)`, `
rule "terraform_naming_convention" {
  enabled = true

  variable {
    format = "none"
  }
}`)
}

func Test_TerraformNamingConventionRule_Variable_DefaultFormat_OverrideDisabled(t *testing.T) {
	testVariableDisabled(t, `overridden config (format=null)`, `
rule "terraform_naming_convention" {
  enabled = true
  format  = "snake_case"

  variable {
    format = "none"
  }
}`)
}

func testVariableSnakeCase(t *testing.T, testType string, formatName string, config string) {
	rule := NewTerraformNamingConventionRule()

	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: fmt.Sprintf("variable: %s - Invalid snake_case with dash", testType),
			Content: `
variable "dash-name" {
  description = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("variable name `dash-name` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 21},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("variable: %s - Invalid snake_case with camelCase", testType),
			Content: `
variable "camelCased" {
  description = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("variable name `camelCased` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 22},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("variable: %s - Invalid snake_case with double underscore", testType),
			Content: `
variable "foo__bar" {
  description = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("variable name `foo__bar` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 20},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("variable: %s - Invalid snake_case with underscore tail", testType),
			Content: `
variable "foo_bar_" {
  description = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("variable name `foo_bar_` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 20},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("variable: %s - Invalid snake_case with Mixed_Snake_Case", testType),
			Content: `
variable "Foo_Bar" {
  description = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: fmt.Sprintf("variable name `Foo_Bar` must match the following %s", formatName),
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 19},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("variable: %s - Valid snake_case", testType),
			Content: `
variable "foo_bar" {
  description = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("variable: %s - Valid single word", testType),
			Content: `
variable "foo" {
  description = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
	}

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"tests.tf": tc.Content}, loadConfigFromNamingConventionTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

func testVariableMixedSnakeCase(t *testing.T, testType string, config string) {
	rule := NewTerraformNamingConventionRule()

	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: fmt.Sprintf("variable: %s - Invalid mixed_snake_case with dash", testType),
			Content: `
variable "dash-name" {
  description = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: "variable name `dash-name` must match the following format: mixed_snake_case",
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 21},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("variable: %s - Invalid mixed_snake_case with double underscore", testType),
			Content: `
variable "Foo__Bar" {
  description = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: "variable name `Foo__Bar` must match the following format: mixed_snake_case",
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 20},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("variable: %s - Invalid mixed_snake_case with underscore tail", testType),
			Content: `
variable "Foo_Bar_" {
  description = "invalid"
}`,
			Config: config,
			Expected: tflint.Issues{
				{
					Rule:    rule,
					Message: "variable name `Foo_Bar_` must match the following format: mixed_snake_case",
					Range: hcl.Range{
						Filename: "tests.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 20},
					},
				},
			},
		},
		{
			Name: fmt.Sprintf("variable: %s - Valid snake_case", testType),
			Content: `
variable "foo_bar" {
  description = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("variable: %s - Valid single word", testType),
			Content: `
variable "foo" {
  description = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("variable: %s - Valid Mixed_Snake_Case", testType),
			Content: `
variable "Foo_Bar" {
  description = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("variable: %s - Valid single word with upper characters", testType),
			Content: `
variable "foo" {
  description = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("variable: %s - Valid PascalCase", testType),
			Content: `
variable "PascalCase" {
  description = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("variable: %s - Valid camelCase", testType),
			Content: `
variable "camelCase" {
  description = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
	}

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"tests.tf": tc.Content}, loadConfigFromNamingConventionTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

func testVariableDisabled(t *testing.T, testType string, config string) {
	rule := NewTerraformNamingConventionRule()

	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: fmt.Sprintf("variable: %s - Valid mixed_snake_case with dash", testType),
			Content: `
variable "dash-name" {
  description = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("variable: %s - Valid snake_case", testType),
			Content: `
variable "foo_bar" {
  description = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("variable: %s - Valid single word", testType),
			Content: `
variable "foo" {
  description = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("variable: %s - Valid Mixed_Snake_Case", testType),
			Content: `
variable "Foo_Bar" {
  description = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("variable: %s - Valid single word upper characters", testType),
			Content: `
variable "Foo" {
  description = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("variable: %s - Valid PascalCase", testType),
			Content: `
variable "PascalCase" {
  description = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
		{
			Name: fmt.Sprintf("variable: %s - Valid camelCase", testType),
			Content: `
variable "camelCase" {
  description = "valid"
}`,
			Config:   config,
			Expected: tflint.Issues{},
		},
	}

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"tests.tf": tc.Content}, loadConfigFromNamingConventionTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

// TODO: Replace with TestRunner
func loadConfigFromNamingConventionTempFile(t *testing.T, content string) *tflint.Config {
	if content == "" {
		return tflint.EmptyConfig()
	}

	tmpfile, err := os.CreateTemp("", "terraform_naming_convention")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	config, err := tflint.LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	return config
}
