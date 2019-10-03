package tflint

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	hcl "github.com/hashicorp/hcl/v2"
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
		opts := []cmp.Option{cmpopts.IgnoreFields(moduleVariable{}, "Callers")}
		if !cmp.Equal(ret, tc.Expected, opts...) {
			t.Fatalf("Failed `%s` test: diff=%s", tc.Name, cmp.Diff(ret, tc.Expected, opts...))
		}
	}
}

func Test_callers(t *testing.T) {
	cases := []struct {
		Name     string
		Var      *moduleVariable
		Expected []hcl.Range
	}{
		{
			Name: "root only",
			Var: &moduleVariable{
				Root:      true,
				Parents:   []*moduleVariable{},
				DeclRange: hcl.Range{Filename: "foo.tf"},
			},
			Expected: []hcl.Range{
				{Filename: "foo.tf"},
			},
		},
		{
			Name: "all roots",
			Var: &moduleVariable{
				Root: false,
				Parents: []*moduleVariable{
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
			Expected: []hcl.Range{
				{Filename: "baz.tf"},
				{Filename: "bar.tf"},
				{Filename: "foo.tf"},
			},
		},
	}

	for _, tc := range cases {
		roots := tc.Var.roots()
		if len(roots) != 1 {
			t.Fatalf("Failed `%s` test: expected 1 root, but get `%d` roots", tc.Name, len(roots))
		}

		ret := roots[0].callers()
		if !cmp.Equal(ret, tc.Expected) {
			t.Fatalf("Failed `%s` test: diff=%s", tc.Name, cmp.Diff(ret, tc.Expected))
		}
	}
}
