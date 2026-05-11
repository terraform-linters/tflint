package cmd

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func TestFillNoRangeIssueFilenames(t *testing.T) {
	issues := tflint.Issues{
		{
			Range: hcl.Range{},
			Callers: []hcl.Range{
				{},
				{Filename: "main.tf"},
			},
		},
		{
			Range: hcl.Range{Filename: "subdir/main.tf"},
		},
	}

	fillNoRangeIssueFilenames("subdir", issues)

	expected := tflint.Issues{
		{
			Range: hcl.Range{Filename: "subdir"},
			Callers: []hcl.Range{
				{Filename: "subdir"},
				{Filename: "main.tf"},
			},
		},
		{
			Range: hcl.Range{Filename: "subdir/main.tf"},
		},
	}

	if diff := cmp.Diff(expected, issues); diff != "" {
		t.Fatal(diff)
	}
}
