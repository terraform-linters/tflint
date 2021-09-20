package terraformrules

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformUnusedRequiredProvidersRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name:     "empty",
			Content:  "",
			Expected: tflint.Issues{},
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
			Expected: tflint.Issues{},
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
			Expected: tflint.Issues{},
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
			Expected: tflint.Issues{},
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
			Expected: tflint.Issues{},
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
			Expected: tflint.Issues{},
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
			Expected: tflint.Issues{},
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
			Expected: tflint.Issues{},
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
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformUnusedRequiredProvidersRule(),
					Message: "provider 'null' is declared in required_providers but not used by the module",
					Range: hcl.Range{
						Filename: "module.tf",
						Start: hcl.Pos{
							Line:   4,
							Column: 14,
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
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformUnusedRequiredProvidersRule(),
					Message: "provider 'null' is declared in required_providers but not used by the module",
					Range: hcl.Range{
						Filename: "module.tf",
						Start: hcl.Pos{
							Line:   4,
							Column: 14,
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
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformUnusedRequiredProvidersRule(),
					Message: "provider 'null' is declared in required_providers but not used by the module",
					Range: hcl.Range{
						Filename: "module.tf",
						Start: hcl.Pos{
							Line:   4,
							Column: 14,
						},
						End: hcl.Pos{
							Line:   6,
							Column: 8,
						},
					},
				},
			},
		},
	}

	rule := NewTerraformUnusedRequiredProvidersRule()

	for _, tc := range cases {
		tc := tc

		t.Run(tc.Name, func(t *testing.T) {
			runner := tflint.TestRunner(t, map[string]string{"module.tf": tc.Content})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			tflint.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}
