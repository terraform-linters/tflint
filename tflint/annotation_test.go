package tflint

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func Test_NewAnnotations(t *testing.T) {
	src := `
resource "aws_instance" "foo" {
  /* tflint-ignore: aws_instance_invalid_type, terraform_deprecated_syntax */
  instance_type = "t2.micro" // tflint-ignore: aws_instance_invalid_type
  # tflint-ignore: aws_instance_invalid_type
  iam_instance_profile = "foo" # This is also comment
  // This is also comment
  instance_type_reason = "t2.micro" // tflint-ignore: aws_instance_invalid_type // With reason
  # tflint-ignore: aws_instance_invalid_type # With reason
  iam_instance_profile_reason = "foo" # This is also comment
}`

	file, diags := hclsyntax.ParseConfig([]byte(src), "resource.tf", hcl.Pos{Byte: 0, Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatal(diags)
	}
	ret, diags := NewAnnotations("resource.tf", file)
	if diags.HasErrors() {
		t.Fatal(diags)
	}

	expected := Annotations{
		{
			Content: "aws_instance_invalid_type, terraform_deprecated_syntax",
			Token: hclsyntax.Token{
				Type:  hclsyntax.TokenComment,
				Bytes: []byte("/* tflint-ignore: aws_instance_invalid_type, terraform_deprecated_syntax */"),
				Range: hcl.Range{
					Filename: "resource.tf",
					Start:    hcl.Pos{Line: 3, Column: 3},
					End:      hcl.Pos{Line: 3, Column: 78},
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
					Start:    hcl.Pos{Line: 4, Column: 30},
					End:      hcl.Pos{Line: 5, Column: 1},
				},
			},
		},
		{
			Content: "aws_instance_invalid_type",
			Token: hclsyntax.Token{
				Type:  hclsyntax.TokenComment,
				Bytes: []byte("# tflint-ignore: aws_instance_invalid_type\n"),
				Range: hcl.Range{
					Filename: "resource.tf",
					Start:    hcl.Pos{Line: 5, Column: 3},
					End:      hcl.Pos{Line: 6, Column: 1},
				},
			},
		},
		{
			Content: "aws_instance_invalid_type",
			Token: hclsyntax.Token{
				Type:  hclsyntax.TokenComment,
				Bytes: []byte("// tflint-ignore: aws_instance_invalid_type // With reason\n"),
				Range: hcl.Range{
					Filename: "resource.tf",
					Start:    hcl.Pos{Line: 8, Column: 37},
					End:      hcl.Pos{Line: 9, Column: 1},
				},
			},
		},
		{
			Content: "aws_instance_invalid_type",
			Token: hclsyntax.Token{
				Type:  hclsyntax.TokenComment,
				Bytes: []byte("# tflint-ignore: aws_instance_invalid_type # With reason\n"),
				Range: hcl.Range{
					Filename: "resource.tf",
					Start:    hcl.Pos{Line: 9, Column: 3},
					End:      hcl.Pos{Line: 10, Column: 1},
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

	tests := []struct {
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
			Name: "affected (multiple rules)",
			Annotation: Annotation{
				Content: "other_rule, test_rule",
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
			Name: "not affected (multiple rules)",
			Annotation: Annotation{
				Content: "other_rule_a, other_rule_b",
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

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			got := test.Annotation.IsAffected(issue)
			if got != test.Expected {
				t.Fatalf("want=%t, got=%t", test.Expected, got)
			}
		})
	}
}
