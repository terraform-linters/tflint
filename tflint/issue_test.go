package tflint

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	hcl "github.com/hashicorp/hcl/v2"
)

func Test_Sort(t *testing.T) {
	issues := Issues{
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test2.tf",
				Start:    hcl.Pos{Line: 1, Column: 1},
				End:      hcl.Pos{Line: 1, Column: 2},
			},
		},
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 2, Column: 1},
				End:      hcl.Pos{Line: 2, Column: 2},
			},
		},
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 1, Column: 1},
				End:      hcl.Pos{Line: 1, Column: 2},
			},
		},
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 1, Column: 1},
				End:      hcl.Pos{Line: 2, Column: 3},
			},
		},
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 2, Column: 2},
				End:      hcl.Pos{Line: 2, Column: 3},
			},
		},
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 1, Column: 1},
				End:      hcl.Pos{Line: 2, Column: 4},
			},
		},
	}

	expected := Issues{
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 1, Column: 1},
				End:      hcl.Pos{Line: 2, Column: 4},
			},
		},
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 1, Column: 1},
				End:      hcl.Pos{Line: 2, Column: 3},
			},
		},
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 1, Column: 1},
				End:      hcl.Pos{Line: 1, Column: 2},
			},
		},
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 2, Column: 1},
				End:      hcl.Pos{Line: 2, Column: 2},
			},
		},
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test1.tf",
				Start:    hcl.Pos{Line: 2, Column: 2},
				End:      hcl.Pos{Line: 2, Column: 3},
			},
		},
		{
			Rule:    &testRule{},
			Message: "test",
			Range: hcl.Range{
				Filename: "test2.tf",
				Start:    hcl.Pos{Line: 1, Column: 1},
				End:      hcl.Pos{Line: 1, Column: 2},
			},
		},
	}

	got := issues.Sort()
	if !cmp.Equal(got, expected) {
		t.Fatalf("Failed: diff=%s", cmp.Diff(got, expected))
	}
}

func Test_Filter(t *testing.T) {
	issues := Issues{
		{
			Rule:    &testRule{},
			Message: "This is test rule",
			Range: hcl.Range{
				Filename: "template.tf",
				Start:    hcl.Pos{Line: 1},
			},
		},
		{
			Rule:    &testRule{},
			Message: "This is test rule",
			Range: hcl.Range{
				Filename: "resource.tf",
				Start:    hcl.Pos{Line: 1},
			},
		},
	}

	ret := issues.Filter([]string{"template.tf"})
	expected := Issues{
		{
			Rule:    &testRule{},
			Message: "This is test rule",
			Range: hcl.Range{
				Filename: "template.tf",
				Start:    hcl.Pos{Line: 1},
			},
		},
	}

	if !cmp.Equal(expected, ret) {
		t.Fatalf("Failed test: diff: %s", cmp.Diff(expected, ret))
	}
}
