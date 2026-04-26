// SPDX-License-Identifier: MPL-2.0

package lang

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang/marks"
	"github.com/zclconf/go-cty/cty"
)

func TestScopeFunctionsEval(t *testing.T) {
	scope := &Scope{
		Data: &dataForTests{
			LocalValues: map[string]cty.Value{
				"greeting_template": cty.StringVal("Hello, ${name}!"),
			},
		},
		BaseDir: "./testdata/functions-test",
	}

	tests := []struct {
		name string
		expr string
		want cty.Value
	}{
		{
			name: "plain core function",
			expr: `upper("hello")`,
			want: cty.StringVal("HELLO"),
		},
		{
			name: "core namespace alias",
			expr: `core::upper("hello")`,
			want: cty.StringVal("HELLO"),
		},
		{
			name: "templatefile",
			expr: `templatefile("hello.tmpl", {name = "Jodie"})`,
			want: cty.StringVal("Hello, Jodie!"),
		},
		{
			name: "core templatefile alias",
			expr: `core::templatefile("hello.tmpl", {name = "Namespaced Jodie"})`,
			want: cty.StringVal("Hello, Namespaced Jodie!"),
		},
		{
			name: "templatestring",
			expr: `templatestring(local.greeting_template, {name = "Alex"})`,
			want: cty.StringVal("Hello, Alex!"),
		},
		{
			name: "terraform provider encode expr",
			expr: `provider::terraform::encode_expr(["a", true])`,
			want: cty.StringVal(`["a", true]`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr := mustParseExpression(t, test.expr)
			got, diags := scope.EvalExpr(expr, cty.DynamicPseudoType)
			assertNoDiagnostics(t, diags)
			if !got.RawEquals(test.want) {
				t.Fatalf("wrong result\ngot:  %#v\nwant: %#v", got, test.want)
			}
		})
	}
}

func TestScopeFunctionsRegistryContainsTerraformProviderBuiltins(t *testing.T) {
	scope := &Scope{BaseDir: "./testdata/functions-test"}
	functions := scope.Functions()

	decode, ok := functions["provider::terraform::decode_tfvars"]
	if !ok {
		t.Fatal("decode_tfvars not registered")
	}
	decoded, err := decode.Call([]cty.Value{cty.StringVal("name = \"Jodie\"\n")})
	if err != nil {
		t.Fatalf("decode_tfvars failed: %s", err)
	}
	if !decoded.RawEquals(cty.ObjectVal(map[string]cty.Value{"name": cty.StringVal("Jodie")})) {
		t.Fatalf("unexpected decode_tfvars result: %#v", decoded)
	}

	encode, ok := functions["provider::terraform::encode_tfvars"]
	if !ok {
		t.Fatal("encode_tfvars not registered")
	}
	encoded, err := encode.Call([]cty.Value{cty.ObjectVal(map[string]cty.Value{"name": cty.StringVal("Jodie")})})
	if err != nil {
		t.Fatalf("encode_tfvars failed: %s", err)
	}
	if !encoded.RawEquals(cty.StringVal("name = \"Jodie\"\n")) {
		t.Fatalf("unexpected encode_tfvars result: %#v", encoded)
	}
}

func TestScopePureOnlyMakesImpureFunctionsUnknown(t *testing.T) {
	scope := &Scope{BaseDir: "./testdata/functions-test", PureOnly: true}

	tests := []struct {
		name string
		args []cty.Value
	}{
		{name: "timestamp"},
		{name: "uuid"},
		{name: "bcrypt", args: []cty.Value{cty.StringVal("hello")}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := scope.Functions()[test.name].Call(test.args)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if got.Type() != cty.String || got.IsKnown() {
				t.Fatalf("expected unknown string, got %#v", got)
			}
		})
	}
}

func TestEvalContextAddsProviderMocks(t *testing.T) {
	scope := &Scope{}
	expr := mustParseExpression(t, `provider::example::mystery("hello", 1)`)

	functionCalls, diags := FunctionCallsInExpr(expr)
	assertNoDiagnostics(t, diags)

	ctx, diags := scope.EvalContext(nil, functionCalls)
	assertNoDiagnostics(t, diags)

	mock, ok := ctx.Functions["provider::example::mystery"]
	if !ok {
		t.Fatal("provider-defined mock function was not installed")
	}

	got, err := mock.Call([]cty.Value{cty.StringVal("hello"), cty.NumberIntVal(1)})
	if err != nil {
		t.Fatalf("mock call failed: %s", err)
	}
	if !got.RawEquals(cty.DynamicVal) {
		t.Fatalf("unexpected mock result: %#v", got)
	}
}

func TestNewMockFunction(t *testing.T) {
	tests := []struct {
		name string
		args []cty.Value
	}{
		{name: "no args"},
		{name: "unknown arg", args: []cty.Value{cty.UnknownVal(cty.String)}},
		{name: "marked arg", args: []cty.Value{cty.StringVal("secret").Mark(marks.Sensitive)}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NewMockFunction(&FunctionCall{Name: "provider::demo::func"}).Call(test.args)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !got.RawEquals(cty.DynamicVal) {
				t.Fatalf("unexpected mock result: %#v", got)
			}
		})
	}
}

func mustParseExpression(t *testing.T, src string) hcl.Expression {
	t.Helper()

	expr, diags := hclsyntax.ParseExpression([]byte(src), "test.hcl", hcl.Pos{Line: 1, Column: 1})
	assertNoDiagnostics(t, diags)
	return expr
}

func assertNoDiagnostics(t *testing.T, diags hcl.Diagnostics) {
	t.Helper()
	if diags.HasErrors() {
		for _, diag := range diags {
			t.Errorf("%s: %s", diag.Summary, diag.Detail)
		}
		t.Fatal("unexpected diagnostics")
	}
}
