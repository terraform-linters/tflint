package tflint

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func Test_NewAnnotations(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = os.Chdir(currentDir); err != nil {
			t.Fatal(err)
		}
	}()

	src, err := os.ReadFile(filepath.Join(currentDir, "test-fixtures", "annotations", "resource.tf"))
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
					Start:    hcl.Pos{Line: 2, Column: 5},
					End:      hcl.Pos{Line: 2, Column: 51},
				},
			},
		},
		{
			Content: "aws_instance_invalid_type",
			Token: hclsyntax.Token{
				Type:  hclsyntax.TokenComment,
				Bytes: []byte(fmt.Sprintf("// tflint-ignore: aws_instance_invalid_type%s", newLine())),
				Range: hcl.Range{
					Filename: "resource.tf",
					Start:    hcl.Pos{Line: 3, Column: 32},
					End:      hcl.Pos{Line: 4, Column: 1},
				},
			},
		},
		{
			Content: "aws_instance_invalid_type",
			Token: hclsyntax.Token{
				Type:  hclsyntax.TokenComment,
				Bytes: []byte(fmt.Sprintf("# tflint-ignore: aws_instance_invalid_type This is also comment%s", newLine())),
				Range: hcl.Range{
					Filename: "resource.tf",
					Start:    hcl.Pos{Line: 4, Column: 5},
					End:      hcl.Pos{Line: 5, Column: 1},
				},
			},
		},
	}

	opts := cmpopts.IgnoreFields(hcl.Pos{}, "Byte")
	if !cmp.Equal(expected, ret, opts) {
		t.Fatalf("Test failed. Diff: %s", cmp.Diff(expected, ret, opts))
	}
}

func Test_IsAffected(t *testing.T) {
	issue := &Issue{
		Rule:    &testRule{},
		Message: "Test rule",
		Range: hcl.Range{
			Filename: "test.tf",
			Start:    hcl.Pos{Line: 2},
		},
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
