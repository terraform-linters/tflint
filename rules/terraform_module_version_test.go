package rules

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestTerraformModuleVersion_Registry(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected helper.Issues
	}{
		{
			Name: "version",
			Content: `
module "m" {
  source = "ns/name/provider"
	version = "1.0.0"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "multiple digits",
			Content: `
module "m" {
  source = "ns/name/provider"
	version = "10.0.0"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "version equals",
			Content: `
module "m" {
  source = "ns/name/provider"
	version = "= 1.0.0"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "prerelease",
			Content: `
module "m" {
  source = "ns/name/provider"
	version = "2.0.0-pre"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "custom host",
			Content: `
module "m" {
  source = "my.private.reigstry/ns/name/provider"
	version = "1.0.0"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "child module",
			Content: `
module "m" {
  source = "ns/name/provider//modules/child"
	version = "1.0.0"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "range",
			Content: `
module "m" {
  source = "ns/name/provider"
	version = "~> 1"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "multiple",
			Content: `
module "m" {
  source = "ns/name/provider"
	version = "1, 2, 3"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "missing",
			Content: `
module "m" {
  source = "ns/name/provider"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModuleVersionRule(),
					Message: `module "m" should specify a version`,
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 2, Column: 1},
						End:      hcl.Pos{Line: 2, Column: 11},
					},
				},
			},
		},
		{
			Name: "exact version valid",
			Content: `
module "m" {
  source = "ns/name/provider"
	version = "1.0.0"
}`,
			Config:   testTerraformModuleVersionExactConfig,
			Expected: helper.Issues{},
		},
		{
			Name: "exact version invalid: multiple constraints",
			Content: `
module "m" {
  source = "ns/name/provider"
	version = "1.0.0, 1.0.1"
}`,
			Config: testTerraformModuleVersionExactConfig,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModuleVersionRule(),
					Message: `module "m" should specify an exact version, but multiple constraints were found`,
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 4, Column: 2},
						End:      hcl.Pos{Line: 4, Column: 26},
					},
				},
			},
		},
		{
			Name: "exact version invalid: range operator",
			Content: `
module "m" {
  source = "ns/name/provider"
	version = "~> 1.0.0"
}`,
			Config: testTerraformModuleVersionExactConfig,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModuleVersionRule(),
					Message: `module "m" should specify an exact version, but a range was found`,
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 4, Column: 2},
						End:      hcl.Pos{Line: 4, Column: 22},
					},
				},
			},
		},
		{
			Name: "exact version invalid: partial version",
			Content: `
module "m" {
  source = "ns/name/provider"
	version = "1.0"
}`,
			Config: testTerraformModuleVersionExactConfig,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModuleVersionRule(),
					Message: `module "m" should specify an exact version, but a range was found`,
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 4, Column: 2},
						End:      hcl.Pos{Line: 4, Column: 17},
					},
				},
			},
		},
	}

	rule := NewTerraformModuleVersionRule()

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"module.tf": tc.Content, ".tflint.hcl": tc.Config})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}

func TestTerraformModuleVersion_NonRegistry(t *testing.T) {
	cases := []struct {
		Name   string
		Source string
	}{
		{
			Name:   "local",
			Source: "./local/dir",
		},
		{
			Name:   "github",
			Source: "github.com/hashicorp/example",
		},
		{
			Name:   "git",
			Source: "git::https://example.com/vpc.git",
		},
		{
			Name:   "https",
			Source: "https://example.com/vpc-module.zip",
		},
	}

	rule := NewTerraformModuleVersionRule()

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			content := fmt.Sprintf(testTerraformModuleVersionNonRegistrySource, tc.Source)
			runner := helper.TestRunner(t, map[string]string{"module.tf": content})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, helper.Issues{}, runner.Issues)
		})
	}
}

const testTerraformModuleVersionExactConfig = `
rule "terraform_module_version" {
	enabled = true
	exact = true
}
`

const testTerraformModuleVersionNonRegistrySource = `
module "m" {
	source = "%s"
}
`
