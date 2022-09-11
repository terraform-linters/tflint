package tflint

import (
	"testing"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func Test_isEvaluableResource(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected int
	}{
		{
			Name: "no meta-arguments",
			Content: `
resource "null_resource" "test" {
}`,
			Expected: 1,
		},
		{
			Name: "count is not zero (literal)",
			Content: `
resource "null_resource" "test" {
  count = 1
}`,
			Expected: 1,
		},
		{
			Name: "count is not zero (variable)",
			Content: `
variable "foo" {
  default = 1
}

resource "null_resource" "test" {
  count = var.foo
}`,
			Expected: 1,
		},
		{
			Name: "count is unknown",
			Content: `
variable "foo" {}

resource "null_resource" "test" {
  count = var.foo
}`,
			Expected: 0,
		},
		{
			Name: "count is sensitive",
			Content: `
variable "foo" {
  default = 1
  sensitive = true
}

resource "null_resource" "test" {
  count = var.foo
}`,
			Expected: 1,
		},
		{
			Name: "count is unevaluable",
			Content: `
resource "null_resource" "test" {
  count = local.foo
}`,
			Expected: 0,
		},
		{
			Name: "count is zero",
			Content: `
resource "null_resource" "test" {
  count = 0
}`,
			Expected: 0,
		},
		{
			// HINT: Terraform does not allow null as `count`
			Name: "count is null",
			Content: `
resource "null_resource" "test" {
  count = null
}`,
			Expected: 1,
		},
		{
			Name: "for_each is not empty (literal)",
			Content: `
resource "null_resource" "test" {
  for_each = {
    foo = "bar"
  }
}`,
			Expected: 1,
		},
		{
			Name: "for_each is not empty (variable)",
			Content: `
variable "object" {
  default = {
    foo = "bar"
  }
}

resource "null_resource" "test" {
  for_each = var.object
}`,
			Expected: 1,
		},
		{
			Name: "for_each is unknown",
			Content: `
variable "foo" {}

resource "null_resource" "test" {
  for_each = var.foo
}`,
			Expected: 0,
		},
		{
			Name: "for_each is unevaluable",
			Content: `
resource "null_resource" "test" {
  for_each = local.foo
}`,
			Expected: 0,
		},
		{
			Name: "for_each contains unevaluable",
			Content: `
resource "null_resource" "test" {
  for_each = {
    known   = "known"
    unknown = local.foo
  }
}`,
			Expected: 1,
		},
		{
			Name: "for_each is empty",
			Content: `
resource "null_resource" "test" {
  for_each = {}
}`,
			Expected: 0,
		},
		{
			Name: "for_each is not empty set",
			Content: `
resource "null_resource" "test" {
  for_each = toset(["foo", "bar"])
}`,
			Expected: 1,
		},
		{
			Name: "for_each is empty set",
			Content: `
resource "null_resource" "test" {
  for_each = toset([])
}`,
			Expected: 0,
		},
		{
			// HINT: Terraform does not allow null as `for_each`
			Name: "for_each is null",
			Content: `
resource "null_resource" "test" {
  for_each = null
}`,
			Expected: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			runner := TestRunner(t, map[string]string{"main.tf": tc.Content})

			resources, diags := runner.GetModuleContent(&hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body:       &hclext.BodySchema{},
					},
				},
			}, sdk.GetModuleContentOption{})
			if diags.HasErrors() {
				t.Fatalf("failed to parse: %s", diags)
			}
			if len(resources.Blocks) != tc.Expected {
				t.Fatalf("%d resources expected, but got %d resources", tc.Expected, len(resources.Blocks))
			}
		})
	}
}
