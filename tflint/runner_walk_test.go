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
				{
					Start: hcl.Pos{
						Line:   3,
						Column: 9,
					},
					End: hcl.Pos{
						Line:   3,
						Column: 14,
					},
				},
			},
		},
		{
			Name: "data source",
			Content: `
data "null_dataresource" "test" {
  key = "foo"
}`,
			Expressions: []hcl.Range{
				{
					Start: hcl.Pos{
						Line:   3,
						Column: 9,
					},
					End: hcl.Pos{
						Line:   3,
						Column: 14,
					},
				},
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
				{
					Start: hcl.Pos{
						Line:   3,
						Column: 12,
					},
					End: hcl.Pos{
						Line:   3,
						Column: 22,
					},
				},
				{
					Start: hcl.Pos{
						Line:   4,
						Column: 12,
					},
					End: hcl.Pos{
						Line:   4,
						Column: 17,
					},
				},
			},
		},
		{
			Name: "provider config",
			Content: `
provider "p" {
  key = "foo"	
}`,
			Expressions: []hcl.Range{
				{
					Start: hcl.Pos{
						Line:   3,
						Column: 9,
					},
					End: hcl.Pos{
						Line:   3,
						Column: 14,
					},
				},
			},
		},
		{
			Name: "locals",
			Content: `
locals {
  key = "foo"	
}`,
			Expressions: []hcl.Range{
				{
					Start: hcl.Pos{
						Line:   3,
						Column: 9,
					},
					End: hcl.Pos{
						Line:   3,
						Column: 14,
					},
				},
			},
		},
		{
			Name: "output",
			Content: `
output "o" {
  value = "foo"	
}`,
			Expressions: []hcl.Range{
				{
					Start: hcl.Pos{
						Line:   3,
						Column: 11,
					},
					End: hcl.Pos{
						Line:   3,
						Column: 16,
					},
				},
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
				{
					Start: hcl.Pos{
						Line:   3,
						Column: 9,
					},
					End: hcl.Pos{
						Line:   3,
						Column: 14,
					},
				},
				{
					Start: hcl.Pos{
						Line:   6,
						Column: 22,
					},
					End: hcl.Pos{
						Line:   6,
						Column: 27,
					},
				},
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
				{
					Start: hcl.Pos{
						Line:   6,
						Column: 16,
					},
					End: hcl.Pos{
						Line:   6,
						Column: 21,
					},
				},
				{
					Start: hcl.Pos{
						Line:   7,
						Column: 19,
					},
					End: hcl.Pos{
						Line:   9,
						Column: 10,
					},
				},
				{
					Start: hcl.Pos{
						Line:   10,
						Column: 17,
					},
					End: hcl.Pos{
						Line:   12,
						Column: 11,
					},
				},
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
			Expressions: []hcl.Range{},
		},
	}

	for _, tc := range cases {
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
				t.Fatalf("Failed `%s` test: expected error is not occurred `%s`", tc.Name, tc.ErrorText)
			}

			opts := cmp.Options{
				cmpopts.IgnoreFields(hcl.Range{}, "Filename"),
				cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
				cmpopts.SortSlices(func(x, y hcl.Range) bool { return x.String() > y.String() }),
			}
			if !cmp.Equal(expressions, tc.Expressions, opts) {
				t.Fatalf("Failed `%s` test: diff=%s", tc.Name, cmp.Diff(expressions, tc.Expressions, opts))
			}
		} else if err.Error() != tc.ErrorText {
			t.Fatalf("Failed `%s` test: expected error is %s, but get %s", tc.Name, tc.ErrorText, err)
		}
	}
}
