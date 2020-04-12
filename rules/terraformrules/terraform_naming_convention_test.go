package terraformrules

import (
	"fmt"
	"io/ioutil"
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
  format  = ""
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
  format  = ""

  data {
    format = "snake_case"
  }
}`)
}

func Test_TerraformNamingConventionRule_Data_DefaultDisabled_OverrideCustom(t *testing.T) {
	testDataSnakeCase(t, `overridden config (format="snake_case")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  format  = ""

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
    format = ""
  }
}`)
}

func Test_TerraformNamingConventionRule_Data_DefaultFormat_OverrideDisabled(t *testing.T) {
	testDataDisabled(t, `overridden config (format=null)`, `
rule "terraform_naming_convention" {
  enabled = true
  format  = "snake_case"

  data {
    format = ""
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
			Name: fmt.Sprintf("data: %s - Invalid mixed_snake_case with dash", testType),
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
  format  = ""
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
  format  = ""

  locals {
    format = "snake_case"
  }
}`)
}

func Test_TerraformNamingConventionRule_Locals_DefaultDisabled_OverrideCustom(t *testing.T) {
	testLocalsSnakeCase(t, `overridden config (format="snake_case")`, "RegExp: ^[a-z][a-z]*(_[a-z]+)*$", `
rule "terraform_naming_convention" {
  enabled = true
  format  = ""

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
    format = ""
  }
}`)
}

func Test_TerraformNamingConventionRule_Locals_DefaultFormat_OverrideDisabled(t *testing.T) {
	testLocalsDisabled(t, `overridden config (format=null)`, `
rule "terraform_naming_convention" {
  enabled = true
  format  = "snake_case"

  locals {
    format = ""
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
			Name: fmt.Sprintf("locals: %s - Invalid mixed_snake_case with dash", testType),
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

// TODO: Replace with TestRunner
func loadConfigFromNamingConventionTempFile(t *testing.T, content string) *tflint.Config {
	if content == "" {
		return tflint.EmptyConfig()
	}

	tmpfile, err := ioutil.TempFile("", "terraform_naming_convention")
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
