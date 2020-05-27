package plugin

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	hcl "github.com/hashicorp/hcl/v2"
	tfplugin "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/tflint"
	"github.com/zclconf/go-cty/cty"
)

func Test_Attributes(t *testing.T) {
	source := `
resource "aws_instance" "foo" {
  instance_type = "t2.micro"
}`

	server := NewServer(tflint.TestRunner(t, map[string]string{"main.tf": source}), map[string][]byte{"main.tf": []byte(source)})
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
	expected := []*tfplugin.Attribute{
		{
			Name: "instance_type",
			Expr: []byte(`"t2.micro"`),
			ExprRange: hcl.Range{
				Filename: "main.tf",
				Start:    hcl.Pos{Line: 3, Column: 19},
				End:      hcl.Pos{Line: 3, Column: 29},
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
	opt := cmpopts.IgnoreFields(hcl.Pos{}, "Byte")
	if !cmp.Equal(expected, resp.Attributes, opt) {
		t.Fatalf("Attributes are not matched: %s", cmp.Diff(expected, resp.Attributes, opt))
	}
}

func Test_EvalExpr(t *testing.T) {
	source := `
variable "instance_type" {
  default = "t2.micro"
}`

	server := NewServer(tflint.TestRunner(t, map[string]string{"main.tf": source}), map[string][]byte{"main.tf": []byte(source)})
	req := &tfplugin.EvalExprRequest{
		Expr: []byte(`var.instance_type`),
		ExprRange: hcl.Range{
			Filename: "template.tf",
			Start:    hcl.Pos{Line: 1, Column: 1},
			End:      hcl.Pos{Line: 1, Column: 1},
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

	server := NewServer(tflint.TestRunner(t, map[string]string{"main.tf": source}), map[string][]byte{"main.tf": []byte(source)})
	req := &tfplugin.EvalExprRequest{
		Expr: []byte(`var.instance_type`),
		ExprRange: hcl.Range{
			Filename: "template.tf",
			Start:    hcl.Pos{Line: 1, Column: 1},
			End:      hcl.Pos{Line: 1, Column: 1},
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
		Message: "Unknown value found in template.tf:1; Please use environment variables or tfvars to set the value",
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

	server := NewServer(runner, map[string][]byte{})
	req := &tfplugin.EmitIssueRequest{
		Rule:    rule,
		Message: "This is test rule",
		Location: hcl.Range{
			Filename: "main.tf",
			Start:    hcl.Pos{Line: 3, Column: 3},
			End:      hcl.Pos{Line: 3, Column: 30},
		},
		Expr: []byte("1"),
		ExprRange: hcl.Range{
			Filename: "template.tf",
			Start:    hcl.Pos{Line: 1, Column: 1},
			End:      hcl.Pos{Line: 1, Column: 1},
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
