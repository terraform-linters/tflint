package plugin

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	tfplugin "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/tflint"
	"github.com/zclconf/go-cty/cty"
)

func Test_Attributes(t *testing.T) {
	source := `
resource "aws_instance" "foo" {
  instance_type = "t2.micro"
}`

	server := NewServer(tflint.TestRunner(t, map[string]string{"main.tf": source}))
	req := &tfplugin.AttributesRequest{
		Resource:      "aws_instance",
		AttributeName: "instance_type",
	}
	var resp tfplugin.AttributesResponse

	err := server.Attributes(req, &resp)
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	if resp.Err != nil {
		t.Fatalf("The response has an unexpected error: %s", resp.Err)
	}
	expected := hcl.Attributes{
		"instance_type": {
			Name: "instance_type",
			Expr: &hclsyntax.TemplateExpr{
				Parts: []hclsyntax.Expression{
					&hclsyntax.LiteralValueExpr{
						Val: cty.StringVal("t2.micro"),
						SrcRange: hcl.Range{
							Filename: "main.tf",
							Start:    hcl.Pos{Line: 3, Column: 20},
							End:      hcl.Pos{Line: 3, Column: 28},
						},
					},
				},
				SrcRange: hcl.Range{
					Filename: "main.tf",
					Start:    hcl.Pos{Line: 3, Column: 19},
					End:      hcl.Pos{Line: 3, Column: 29},
				},
			},
			Range: hcl.Range{
				Filename: "main.tf",
				Start:    hcl.Pos{Line: 3, Column: 3},
				End:      hcl.Pos{Line: 3, Column: 29},
			},
			NameRange: hcl.Range{
				Filename: "main.tf",
				Start:    hcl.Pos{Line: 3, Column: 3},
				End:      hcl.Pos{Line: 3, Column: 16},
			},
		},
	}
	opts := []cmp.Option{
		cmpopts.IgnoreUnexported(cty.Type{}, cty.Value{}),
		cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
	}
	if !cmp.Equal(expected, resp.Attributes, opts...) {
		t.Fatalf("Attributes are not matched: %s", cmp.Diff(expected, resp.Attributes, opts...))
	}
}

func Test_EvalExpr(t *testing.T) {
	source := `
variable "instance_type" {
  default = "t2.micro"
}`

	server := NewServer(tflint.TestRunner(t, map[string]string{"main.tf": source}))
	req := &tfplugin.EvalExprRequest{
		Expr: &hclsyntax.ScopeTraversalExpr{
			Traversal: hcl.Traversal{
				hcl.TraverseRoot{Name: "var"},
				hcl.TraverseAttr{Name: "instance_type"},
			},
		},
		Ret: "", // string value
	}
	var resp tfplugin.EvalExprResponse

	err := server.EvalExpr(req, &resp)
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	if resp.Err != nil {
		t.Fatalf("The response has an unexpected error: %s", resp.Err)
	}
	expected := cty.StringVal("t2.micro")
	opts := []cmp.Option{
		cmpopts.IgnoreUnexported(cty.Type{}, cty.Value{}),
		cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
	}
	if !cmp.Equal(expected, resp.Val, opts...) {
		t.Fatalf("Value is not matched: %s", cmp.Diff(expected, resp.Val, opts...))
	}
}

func Test_EvalExpr_errors(t *testing.T) {
	source := `variable "instance_type" {}`

	server := NewServer(tflint.TestRunner(t, map[string]string{"main.tf": source}))
	req := &tfplugin.EvalExprRequest{
		Expr: &hclsyntax.ScopeTraversalExpr{
			Traversal: hcl.Traversal{
				hcl.TraverseRoot{Name: "var"},
				hcl.TraverseAttr{Name: "instance_type"},
			},
		},
		Ret: "", // string value
	}
	var resp tfplugin.EvalExprResponse

	err := server.EvalExpr(req, &resp)
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	expected := tfplugin.Error{
		Code:    tfplugin.UnknownValueError,
		Level:   tfplugin.WarningLevel,
		Message: "Unknown value found in :0; Please use environment variables or tfvars to set the value",
		Cause:   nil,
	}
	if !cmp.Equal(expected, resp.Err) {
		t.Fatalf("Error it not matched: %s", cmp.Diff(expected, resp.Err))
	}
}

func Test_EmitIssue(t *testing.T) {
	runner := tflint.TestRunner(t, map[string]string{})
	rule := &tfplugin.RuleObject{
		Data: &tfplugin.RuleObjectData{
			Name:     "test_rule",
			Severity: tfplugin.ERROR,
		},
	}

	server := NewServer(runner)
	req := &tfplugin.EmitIssueRequest{
		Rule:    rule,
		Message: "This is test rule",
		Location: hcl.Range{
			Filename: "main.tf",
			Start:    hcl.Pos{Line: 3, Column: 3},
			End:      hcl.Pos{Line: 3, Column: 30},
		},
	}
	var resp interface{}

	err := server.EmitIssue(req, &resp)
	if err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}

	expected := tflint.Issues{
		{
			Rule:    rule,
			Message: "This is test rule",
			Range: hcl.Range{
				Filename: "main.tf",
				Start:    hcl.Pos{Line: 3, Column: 3},
				End:      hcl.Pos{Line: 3, Column: 30},
			},
		},
	}
	if !cmp.Equal(expected, runner.Issues) {
		t.Fatalf("Issue are not matched: %s", cmp.Diff(expected, runner.Issues))
	}
}
