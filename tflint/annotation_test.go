package tflint

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/wata727/tflint/issue"
)

func Test_NewAnnotations(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	src, err := ioutil.ReadFile(filepath.Join(currentDir, "test-fixtures", "annotations", "resource.tf"))
	if err != nil {
		t.Fatal(err)
	}
	tokens, diags := hclsyntax.LexConfig(src, "resource.tf", hcl.Pos{Byte: 0, Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatal(diags)
	}
	ret := NewAnnotations(tokens)

	expected := Annotations{
		{
			Content: "aws_instance_invalid_type",
			Token: hclsyntax.Token{
				Type:  hclsyntax.TokenComment,
				Bytes: []byte("/* tflint-ignore: aws_instance_invalid_type */"),
				Range: hcl.Range{
					Filename: "resource.tf",
					Start:    hcl.Pos{Line: 2, Column: 5, Byte: 36},
					End:      hcl.Pos{Line: 2, Column: 51, Byte: 82},
				},
			},
		},
		{
			Content: "aws_instance_invalid_type",
			Token: hclsyntax.Token{
				Type:  hclsyntax.TokenComment,
				Bytes: []byte("// tflint-ignore: aws_instance_invalid_type\n"),
				Range: hcl.Range{
					Filename: "resource.tf",
					Start:    hcl.Pos{Line: 3, Column: 32, Byte: 114},
					End:      hcl.Pos{Line: 4, Column: 1, Byte: 158},
				},
			},
		},
		{
			Content: "aws_instance_invalid_type",
			Token: hclsyntax.Token{
				Type:  hclsyntax.TokenComment,
				Bytes: []byte("# tflint-ignore: aws_instance_invalid_type This is also comment\n"),
				Range: hcl.Range{
					Filename: "resource.tf",
					Start:    hcl.Pos{Line: 4, Column: 5, Byte: 162},
					End:      hcl.Pos{Line: 5, Column: 1, Byte: 226},
				},
			},
		},
	}

	if !cmp.Equal(expected, ret) {
		t.Fatalf("Test failed. Diff: %s", cmp.Diff(expected, ret))
	}
}

func Test_IsAffected(t *testing.T) {
	issue := &issue.Issue{
		Detector: "test_rule",
		Type:     issue.ERROR,
		Message:  "Test rule",
		Line:     2,
		File:     "test.tf",
	}

	cases := []struct {
		Name       string
		Annotation Annotation
		Expected   bool
	}{
		{
			Name: "affected (same line)",
			Annotation: Annotation{
				Content: "test_rule",
				Token: hclsyntax.Token{
					Type: hclsyntax.TokenComment,
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2},
					},
				},
			},
			Expected: true,
		},
		{
			Name: "affected (above line)",
			Annotation: Annotation{
				Content: "test_rule",
				Token: hclsyntax.Token{
					Type: hclsyntax.TokenComment,
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1},
					},
				},
			},
			Expected: true,
		},
		{
			Name: "not affected (under line)",
			Annotation: Annotation{
				Content: "test_rule",
				Token: hclsyntax.Token{
					Type: hclsyntax.TokenComment,
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3},
					},
				},
			},
			Expected: false,
		},
		{
			Name: "not affected (another filename)",
			Annotation: Annotation{
				Content: "test_rule",
				Token: hclsyntax.Token{
					Type: hclsyntax.TokenComment,
					Range: hcl.Range{
						Filename: "test2.tf",
						Start:    hcl.Pos{Line: 2},
					},
				},
			},
			Expected: false,
		},
		{
			Name: "not affected (another rule)",
			Annotation: Annotation{
				Content: "test_another_rule",
				Token: hclsyntax.Token{
					Type: hclsyntax.TokenComment,
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2},
					},
				},
			},
			Expected: false,
		},
		{
			Name: "affected (all)",
			Annotation: Annotation{
				Content: "all",
				Token: hclsyntax.Token{
					Type: hclsyntax.TokenComment,
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 2},
					},
				},
			},
			Expected: true,
		},
	}

	for _, tc := range cases {
		ret := tc.Annotation.IsAffected(issue)
		if ret != tc.Expected {
			t.Fatalf("Failed `%s` test: expected=%t, got=%t", tc.Name, tc.Expected, ret)
		}
	}
}
