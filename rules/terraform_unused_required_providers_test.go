package rules

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_TerraformUnusedRequiredProvidersRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name:     "empty",
			Content:  "",
			Expected: helper.Issues{},
		},
		{
			Name: "used - resource",
			Content: `
				terraform {
					required_providers {
						null = {
							source = "hashicorp/null"
						}
					}
				}
				resource "null_resource" "foo" {}
			`,
			Expected: helper.Issues{},
		},
		{
			Name: "used - data source",
			Content: `
				terraform {
					required_providers {
						null = {
							source = "hashicorp/null"
						}
					}
				}
				resource "null_data_source" "foo" {}
			`,
			Expected: helper.Issues{},
		},
		{
			Name: "used - resource provider override",
			Content: `
				terraform {
					required_providers {
						custom-null = {
							source = "custom/null"
						}
					}
				}
				resource "null_resource" "foo" {
					provider = custom-null
				}
			`,
			Expected: helper.Issues{},
		},
		{
			Name: "used - data source provider override",
			Content: `
				terraform {
					required_providers {
						custom-null = {
							source = "custom/null"
						}
					}
				}
				resource "null_data_source" "foo" {
					provider = custom-null
				}
			`,
			Expected: helper.Issues{},
		},
		{
			Name: "used - module provider override",
			Content: `
				terraform {
					required_providers {
						custom-null = {
							source = "custom/null"
						}
					}
				}
				module "m" {
					source = "./m"
					providers = {
						null = custom-null
					}
				}
			`,
			Expected: helper.Issues{},
		},
		{
			Name: "used - module provider override with alias",
			Content: `
				terraform {
					required_providers {
						null = {
							source = "hashicorp/null"
							configuration_aliases = [null.a]
						}
					}
				}
				module "m" {
					source = "./m"
					providers = {
						null = null.a
					}
				}
			`,
			Expected: helper.Issues{},
		},
		{
			Name: "used - provider",
			Content: `
				terraform {
					required_providers {
						null = {
							source = "hashicorp/null"
						}
					}
				}
				provider "null" {}
			`,
			Expected: helper.Issues{},
		},
		{
			Name: "unused",
			Content: `
				terraform {
					required_providers {
						null = {
							source = "hashicorp/null"
						}
					}
				}
			`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformUnusedRequiredProvidersRule(),
					Message: "provider 'null' is declared in required_providers but not used by the module",
					Range: hcl.Range{
						Filename: "module.tf",
						Start: hcl.Pos{
							Line:   4,
							Column: 7,
						},
						End: hcl.Pos{
							Line:   6,
							Column: 8,
						},
					},
				},
			},
		},
		{
			Name: "unused - override",
			Content: `
				terraform {
					required_providers {
						null = {
							source = "hashicorp/null"
						}
						custom-null = {
							source = "custom/null"
						}
					}
				}
				resource "null_resource" "foo" {
					provider = custom-null
				}
			`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformUnusedRequiredProvidersRule(),
					Message: "provider 'null' is declared in required_providers but not used by the module",
					Range: hcl.Range{
						Filename: "module.tf",
						Start: hcl.Pos{
							Line:   4,
							Column: 7,
						},
						End: hcl.Pos{
							Line:   6,
							Column: 8,
						},
					},
				},
			},
		},
		{
			Name: "unused - module",
			Content: `
				terraform {
					required_providers {
						null = {
							source = "hashicorp/null"
						}
					}
				}
				module "m" {
					source = "./m"
				}
			`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformUnusedRequiredProvidersRule(),
					Message: "provider 'null' is declared in required_providers but not used by the module",
					Range: hcl.Range{
						Filename: "module.tf",
						Start: hcl.Pos{
							Line:   4,
							Column: 7,
						},
						End: hcl.Pos{
							Line:   6,
							Column: 8,
						},
					},
				},
			},
		},
		{
			Name: "used - unevaluated resource",
			Content: `
				terraform {
					required_providers {
						null = {
							source = "hashicorp/null"
						}
					}
				}
				variable "foo" {}
				resource "null_resource" "foo" {
					count = var.foo
				}
			`,
			Expected: helper.Issues{},
		},
	}

	rule := NewTerraformUnusedRequiredProvidersRule()

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			runner := testRunner(t, map[string]string{"module.tf": tc.Content})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, tc.Expected, runner.Runner.(*helper.Runner).Issues)
		})
	}
}
