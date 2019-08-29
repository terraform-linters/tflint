package tflint

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl2/hcl"
)

func Test_roots(t *testing.T) {
	cases := []struct {
		Name     string
		Var      *moduleVariable
		Expected []*moduleVariable
	}{
		{
			Name: "root only",
			Var: &moduleVariable{
				Root:      true,
				Parents:   []*moduleVariable{},
				DeclRange: hcl.Range{Filename: "foo.tf"},
			},
			Expected: []*moduleVariable{
				{
					Root:      true,
					Parents:   []*moduleVariable{},
					DeclRange: hcl.Range{Filename: "foo.tf"},
				},
			},
		},
		{
			Name: "all roots",
			Var: &moduleVariable{
				Root: false,
				Parents: []*moduleVariable{
					{
						Root:      true,
						Parents:   []*moduleVariable{},
						DeclRange: hcl.Range{Filename: "bar.tf"},
					},
					{
						Root:      false,
						Parents:   []*moduleVariable{},
						DeclRange: hcl.Range{Filename: "bar.tf"},
					},
					{
						Root: false,
						Parents: []*moduleVariable{
							{
								Root:      true,
								Parents:   []*moduleVariable{},
								DeclRange: hcl.Range{Filename: "baz.tf"},
							},
						},
						DeclRange: hcl.Range{Filename: "bar.tf"},
					},
				},
				DeclRange: hcl.Range{Filename: "foo.tf"},
			},
			Expected: []*moduleVariable{
				{
					Root:      true,
					Parents:   []*moduleVariable{},
					DeclRange: hcl.Range{Filename: "bar.tf"},
				},
				{
					Root:      true,
					Parents:   []*moduleVariable{},
					DeclRange: hcl.Range{Filename: "baz.tf"},
				},
			},
		},
	}

	for _, tc := range cases {
		ret := tc.Var.roots()
		if !cmp.Equal(ret, tc.Expected) {
			t.Fatalf("Failed `%s` test: diff=%s", tc.Name, cmp.Diff(ret, tc.Expected))
		}
	}
}
