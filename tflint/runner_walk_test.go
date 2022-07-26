package tflint

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	hcl "github.com/hashicorp/hcl/v2"
)

func Test_WalkExpressions(t *testing.T) {
	cases := []struct {
		Name        string
		Content     string
		Override    string
		JSON        bool
		Expressions []hcl.Range
		ErrorText   string
	}{
		{
			Name: "resource",
			Content: `
resource "null_resource" "test" {
  key = "foo"
}`,
			Expressions: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 9}, End: hcl.Pos{Line: 3, Column: 14}},
				{Start: hcl.Pos{Line: 3, Column: 10}, End: hcl.Pos{Line: 3, Column: 13}},
			},
		},
		{
			Name: "data source",
			Content: `
data "null_dataresource" "test" {
  key = "foo"
}`,
			Expressions: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 9}, End: hcl.Pos{Line: 3, Column: 14}},
				{Start: hcl.Pos{Line: 3, Column: 10}, End: hcl.Pos{Line: 3, Column: 13}},
			},
		},
		{
			Name: "module call",
			Content: `
module "m" {
  source = "./module"
  key    = "foo"
}`,
			Expressions: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 12}, End: hcl.Pos{Line: 3, Column: 22}},
				{Start: hcl.Pos{Line: 3, Column: 13}, End: hcl.Pos{Line: 3, Column: 21}},
				{Start: hcl.Pos{Line: 4, Column: 12}, End: hcl.Pos{Line: 4, Column: 17}},
				{Start: hcl.Pos{Line: 4, Column: 13}, End: hcl.Pos{Line: 4, Column: 16}},
			},
		},
		{
			Name: "provider config",
			Content: `
provider "p" {
  key = "foo"
}`,
			Expressions: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 9}, End: hcl.Pos{Line: 3, Column: 14}},
				{Start: hcl.Pos{Line: 3, Column: 10}, End: hcl.Pos{Line: 3, Column: 13}},
			},
		},
		{
			Name: "locals",
			Content: `
locals {
  key = "foo"
}`,
			Expressions: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 9}, End: hcl.Pos{Line: 3, Column: 14}},
				{Start: hcl.Pos{Line: 3, Column: 10}, End: hcl.Pos{Line: 3, Column: 13}},
			},
		},
		{
			Name: "output",
			Content: `
output "o" {
  value = "foo"
}`,
			Expressions: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 11}, End: hcl.Pos{Line: 3, Column: 16}},
				{Start: hcl.Pos{Line: 3, Column: 12}, End: hcl.Pos{Line: 3, Column: 15}},
			},
		},
		{
			Name: "resource with block",
			Content: `
resource "null_resource" "test" {
  key = "foo"

  lifecycle {
    ignore_changes = [key]
  }
}`,
			Expressions: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 9}, End: hcl.Pos{Line: 3, Column: 14}},
				{Start: hcl.Pos{Line: 3, Column: 10}, End: hcl.Pos{Line: 3, Column: 13}},
				{Start: hcl.Pos{Line: 6, Column: 22}, End: hcl.Pos{Line: 6, Column: 27}},
				{Start: hcl.Pos{Line: 6, Column: 23}, End: hcl.Pos{Line: 6, Column: 26}},
			},
		},
		{
			Name: "resource json",
			JSON: true,
			Content: `
{
  "resource": {
    "null_resource": {
      "test": {
        "key": "foo",
        "nested": {
          "key": "foo"
        },
        "list": [{
          "key": "foo"
        }]
      }
    }
  }
}`,
			Expressions: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 15}, End: hcl.Pos{Line: 15, Column: 4}},
			},
		},
		{
			Name: "merged config",
			Content: `
provider "aws" {
  region = "us-east-1"

  assume_role {
    role_arn = "arn:aws:iam::123412341234:role/ExampleRole"
  }
}`,
			Override: `
provider "aws" {
  region = "us-east-1"

  assume_role {
    role_arn = null
  }
}`,
			Expressions: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 12}, End: hcl.Pos{Line: 3, Column: 23}, Filename: "main.tf"},
				{Start: hcl.Pos{Line: 3, Column: 13}, End: hcl.Pos{Line: 3, Column: 22}, Filename: "main.tf"},
				{Start: hcl.Pos{Line: 6, Column: 16}, End: hcl.Pos{Line: 6, Column: 60}, Filename: "main.tf"},
				{Start: hcl.Pos{Line: 6, Column: 17}, End: hcl.Pos{Line: 6, Column: 59}, Filename: "main.tf"},
				{Start: hcl.Pos{Line: 3, Column: 12}, End: hcl.Pos{Line: 3, Column: 23}, Filename: "main_override.tf"},
				{Start: hcl.Pos{Line: 3, Column: 13}, End: hcl.Pos{Line: 3, Column: 22}, Filename: "main_override.tf"},
				{Start: hcl.Pos{Line: 6, Column: 16}, End: hcl.Pos{Line: 6, Column: 20}, Filename: "main_override.tf"},
			},
		},
		{
			Name: "nested attributes",
			Content: `
data "terraform_remote_state" "remote_state" {
  backend = "remote"

  config = {
    organization = "Organization"
    workspaces = {
      name = "${var.environment}"
    }
  }
}`,
			Expressions: []hcl.Range{
				{Start: hcl.Pos{Line: 3, Column: 13}, End: hcl.Pos{Line: 3, Column: 21}},
				{Start: hcl.Pos{Line: 3, Column: 14}, End: hcl.Pos{Line: 3, Column: 20}},
				{Start: hcl.Pos{Line: 5, Column: 12}, End: hcl.Pos{Line: 10, Column: 4}},
				{Start: hcl.Pos{Line: 6, Column: 5}, End: hcl.Pos{Line: 6, Column: 17}},
				{Start: hcl.Pos{Line: 6, Column: 20}, End: hcl.Pos{Line: 6, Column: 34}},
				{Start: hcl.Pos{Line: 6, Column: 21}, End: hcl.Pos{Line: 6, Column: 33}},
				{Start: hcl.Pos{Line: 7, Column: 5}, End: hcl.Pos{Line: 7, Column: 15}},
				{Start: hcl.Pos{Line: 7, Column: 18}, End: hcl.Pos{Line: 9, Column: 6}},
				{Start: hcl.Pos{Line: 8, Column: 7}, End: hcl.Pos{Line: 8, Column: 11}},
				{Start: hcl.Pos{Line: 8, Column: 14}, End: hcl.Pos{Line: 8, Column: 34}},
				{Start: hcl.Pos{Line: 8, Column: 17}, End: hcl.Pos{Line: 8, Column: 32}},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			filename := "main.tf"
			override := "main_override.tf"
			if tc.JSON {
				filename += ".json"
				override += ".json"
			}

			var runner *Runner
			if tc.Override != "" {
				runner = TestRunner(t, map[string]string{filename: tc.Content, override: tc.Override})
			} else {
				runner = TestRunner(t, map[string]string{filename: tc.Content})
			}
			expressions := make([]hcl.Range, 0)

			err := runner.WalkExpressions(func(expr hcl.Expression) error {
				expressions = append(expressions, expr.Range())
				return nil
			})
			if err == nil {
				if tc.ErrorText != "" {
					t.Fatalf("expected error is not occurred `%s`", tc.ErrorText)
				}

				opts := cmp.Options{
					cmpopts.IgnoreFields(hcl.Range{}, "Filename"),
					cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
					cmpopts.SortSlices(func(x, y hcl.Range) bool { return x.String() > y.String() }),
				}
				if diff := cmp.Diff(expressions, tc.Expressions, opts); diff != "" {
					t.Fatal(diff)
				}
			} else if err.Error() != tc.ErrorText {
				t.Fatalf("expected error is %s, but get %s", tc.ErrorText, err)
			}
		})
	}
}
