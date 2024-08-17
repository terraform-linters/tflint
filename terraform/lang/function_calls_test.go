package lang

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/zclconf/go-cty/cty"
)

func TestFunctionCallsInExpr(t *testing.T) {
	parse := func(src string) hcl.Expression {
		expr, diags := hclsyntax.ParseExpression([]byte(src), "", hcl.InitialPos)
		if diags.HasErrors() {
			t.Fatalf("failed to parse `%s`, %s", src, diags)
		}
		return expr
	}
	parseJSON := func(src string) hcl.Expression {
		expr, diags := json.ParseExpression([]byte(src), "")
		if diags.HasErrors() {
			t.Fatalf("failed to parse `%s`, %s", src, diags)
		}
		return expr
	}

	tests := []struct {
		name string
		expr hcl.Expression
		want []*FunctionCall
	}{
		{
			name: "nil expression",
			expr: nil,
			want: nil,
		},
		{
			name: "string",
			expr: parse(`"string"`),
			want: []*FunctionCall{},
		},
		{
			name: "string (JSON)",
			expr: parseJSON(`"string"`),
			want: []*FunctionCall{},
		},
		{
			name: "number (JSON)",
			expr: parseJSON(`123`),
			want: []*FunctionCall{},
		},
		{
			name: "null (JSON)",
			expr: parseJSON(`null`),
			want: []*FunctionCall{},
		},
		{
			name: "unknown (JSON)",
			expr: parseJSON(`"${var.foo}"`),
			want: []*FunctionCall{},
		},
		{
			name: "single function call",
			expr: parse(`md5("hello")`),
			want: []*FunctionCall{
				{Name: "md5", ArgsCount: 1},
			},
		},
		{
			name: "single function call (JSON)",
			expr: parseJSON(`"${md5(\"hello\")}"`),
			want: []*FunctionCall{
				{Name: "md5", ArgsCount: 1},
			},
		},
		{
			name: "multiple function calls",
			expr: parse(`[md5("hello"), "world", provider::tflint::world()]`),
			want: []*FunctionCall{
				{Name: "md5", ArgsCount: 1},
				{Name: "provider::tflint::world", ArgsCount: 0},
			},
		},
		{
			name: "multiple function calls (JSON)",
			expr: parseJSON(`["${md5(\"hello\")}", "world", "${provider::tflint::world()}"]`),
			want: []*FunctionCall{
				{Name: "md5", ArgsCount: 1},
				{Name: "provider::tflint::world", ArgsCount: 0},
			},
		},
		{
			name: "bound expr with native syntax",
			expr: hclext.BindValue(cty.StringVal("foo-Hello, John and Mike"), parse(`"foo-${hello("John", "Mike")}"`)),
			want: []*FunctionCall{
				{Name: "hello", ArgsCount: 2},
			},
		},
		{
			name: "bound expr with JSON syntax",
			expr: hclext.BindValue(cty.StringVal("foo-Hello, John and Mike"), parseJSON(`"foo-${hello(\"John\", \"Mike\")}"`)),
			want: []*FunctionCall{
				{Name: "hello", ArgsCount: 2},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, diags := FunctionCallsInExpr(test.expr)
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

func TestIsProviderDefined(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{
			name: "md5",
			want: false,
		},
		{
			name: "core::md5",
			want: false,
		},
		{
			name: "provider::tflint::echo",
			want: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := &FunctionCall{Name: test.name}

			got := f.IsProviderDefined()

			if got != test.want {
				t.Errorf("got %t, want %t", got, test.want)
			}
		})
	}
}
