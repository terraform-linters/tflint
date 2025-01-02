package lang

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang/marks"

	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

func TestScopeEvalContext(t *testing.T) {
	data := &dataForTests{
		CountAttrs: map[string]cty.Value{
			"index": cty.NumberIntVal(0),
		},
		ForEachAttrs: map[string]cty.Value{
			"key":   cty.StringVal("a"),
			"value": cty.NumberIntVal(1),
		},
		LocalValues: map[string]cty.Value{
			"foo": cty.StringVal("bar"),
		},
		PathAttrs: map[string]cty.Value{
			"module": cty.StringVal("foo/bar"),
		},
		TerraformAttrs: map[string]cty.Value{
			"workspace": cty.StringVal("default"),
		},
		InputVariables: map[string]cty.Value{
			"baz": cty.StringVal("boop"),
		},
	}

	tests := []struct {
		Expr string
		Want map[string]cty.Value
	}{
		{
			`12`,
			map[string]cty.Value{},
		},
		{
			`count.index`,
			map[string]cty.Value{
				"count": cty.ObjectVal(map[string]cty.Value{
					"index": cty.NumberIntVal(0),
				}),
				"resource":  cty.DynamicVal,
				"ephemeral": cty.DynamicVal.Mark(marks.Ephemeral),
				"data":      cty.DynamicVal,
				"module":    cty.DynamicVal,
				"self":      cty.DynamicVal,
			},
		},
		{
			`each.key`,
			map[string]cty.Value{
				"each": cty.ObjectVal(map[string]cty.Value{
					"key": cty.StringVal("a"),
				}),
				"resource":  cty.DynamicVal,
				"ephemeral": cty.DynamicVal.Mark(marks.Ephemeral),
				"data":      cty.DynamicVal,
				"module":    cty.DynamicVal,
				"self":      cty.DynamicVal,
			},
		},
		{
			`each.value`,
			map[string]cty.Value{
				"each": cty.ObjectVal(map[string]cty.Value{
					"value": cty.NumberIntVal(1),
				}),
				"resource":  cty.DynamicVal,
				"ephemeral": cty.DynamicVal.Mark(marks.Ephemeral),
				"data":      cty.DynamicVal,
				"module":    cty.DynamicVal,
				"self":      cty.DynamicVal,
			},
		},
		{
			`local.foo`,
			map[string]cty.Value{
				"local": cty.ObjectVal(map[string]cty.Value{
					"foo": cty.StringVal("bar"),
				}),
				"resource":  cty.DynamicVal,
				"ephemeral": cty.DynamicVal.Mark(marks.Ephemeral),
				"data":      cty.DynamicVal,
				"module":    cty.DynamicVal,
				"self":      cty.DynamicVal,
			},
		},
		{
			`null_resource.foo`,
			map[string]cty.Value{
				"null_resource": cty.DynamicVal,
				"resource":      cty.DynamicVal,
				"ephemeral":     cty.DynamicVal.Mark(marks.Ephemeral),
				"data":          cty.DynamicVal,
				"module":        cty.DynamicVal,
				"self":          cty.DynamicVal,
			},
		},
		{
			`null_resource.foo.attr`,
			map[string]cty.Value{
				"null_resource": cty.DynamicVal,
				"resource":      cty.DynamicVal,
				"ephemeral":     cty.DynamicVal.Mark(marks.Ephemeral),
				"data":          cty.DynamicVal,
				"module":        cty.DynamicVal,
				"self":          cty.DynamicVal,
			},
		},
		{
			`null_resource.multi`,
			map[string]cty.Value{
				"null_resource": cty.DynamicVal,
				"resource":      cty.DynamicVal,
				"ephemeral":     cty.DynamicVal.Mark(marks.Ephemeral),
				"data":          cty.DynamicVal,
				"module":        cty.DynamicVal,
				"self":          cty.DynamicVal,
			},
		},
		{
			`null_resource.multi[1]`,
			map[string]cty.Value{
				"null_resource": cty.DynamicVal,
				"resource":      cty.DynamicVal,
				"ephemeral":     cty.DynamicVal.Mark(marks.Ephemeral),
				"data":          cty.DynamicVal,
				"module":        cty.DynamicVal,
				"self":          cty.DynamicVal,
			},
		},
		{
			`null_resource.each["each1"]`,
			map[string]cty.Value{
				"null_resource": cty.DynamicVal,
				"resource":      cty.DynamicVal,
				"ephemeral":     cty.DynamicVal.Mark(marks.Ephemeral),
				"data":          cty.DynamicVal,
				"module":        cty.DynamicVal,
				"self":          cty.DynamicVal,
			},
		},
		{
			`null_resource.each["each1"].attr`,
			map[string]cty.Value{
				"null_resource": cty.DynamicVal,
				"resource":      cty.DynamicVal,
				"ephemeral":     cty.DynamicVal.Mark(marks.Ephemeral),
				"data":          cty.DynamicVal,
				"module":        cty.DynamicVal,
				"self":          cty.DynamicVal,
			},
		},
		{
			`foo(null_resource.multi, null_resource.multi[1])`,
			map[string]cty.Value{
				"null_resource": cty.DynamicVal,
				"resource":      cty.DynamicVal,
				"ephemeral":     cty.DynamicVal.Mark(marks.Ephemeral),
				"data":          cty.DynamicVal,
				"module":        cty.DynamicVal,
				"self":          cty.DynamicVal,
			},
		},
		{
			`path.module`,
			map[string]cty.Value{
				"path": cty.ObjectVal(map[string]cty.Value{
					"module": cty.StringVal("foo/bar"),
				}),
				"resource":  cty.DynamicVal,
				"ephemeral": cty.DynamicVal.Mark(marks.Ephemeral),
				"data":      cty.DynamicVal,
				"module":    cty.DynamicVal,
				"self":      cty.DynamicVal,
			},
		},
		{
			`terraform.workspace`,
			map[string]cty.Value{
				"terraform": cty.ObjectVal(map[string]cty.Value{
					"workspace": cty.StringVal("default"),
				}),
				"resource":  cty.DynamicVal,
				"ephemeral": cty.DynamicVal.Mark(marks.Ephemeral),
				"data":      cty.DynamicVal,
				"module":    cty.DynamicVal,
				"self":      cty.DynamicVal,
			},
		},
		{
			`var.baz`,
			map[string]cty.Value{
				"var": cty.ObjectVal(map[string]cty.Value{
					"baz": cty.StringVal("boop"),
				}),
				"resource":  cty.DynamicVal,
				"ephemeral": cty.DynamicVal.Mark(marks.Ephemeral),
				"data":      cty.DynamicVal,
				"module":    cty.DynamicVal,
				"self":      cty.DynamicVal,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Expr, func(t *testing.T) {
			expr, parseDiags := hclsyntax.ParseExpression([]byte(test.Expr), "", hcl.Pos{Line: 1, Column: 1})
			if len(parseDiags) != 0 {
				t.Errorf("unexpected diagnostics during parse")
				for _, diag := range parseDiags {
					t.Errorf("- %s", diag)
				}
				return
			}

			refs, refsDiags := ReferencesInExpr(expr)
			if refsDiags.HasErrors() {
				t.Fatal(refsDiags)
			}

			funcCalls, funcDiags := FunctionCallsInExpr(expr)
			if funcDiags.HasErrors() {
				t.Fatal(funcDiags)
			}

			scope := &Scope{
				Data: data,
			}
			ctx, ctxDiags := scope.EvalContext(refs, funcCalls)
			if ctxDiags.HasErrors() {
				t.Fatal(ctxDiags)
			}

			// For easier test assertions we'll just remove any top-level
			// empty objects from our variables map.
			for k, v := range ctx.Variables {
				if v.RawEquals(cty.EmptyObjectVal) {
					delete(ctx.Variables, k)
				}
			}

			gotVal := cty.ObjectVal(ctx.Variables)
			wantVal := cty.ObjectVal(test.Want)

			if !gotVal.RawEquals(wantVal) {
				// We'll JSON-ize our values here just so it's easier to
				// read them in the assertion output.
				gotJSON := formattedJSONValue(gotVal)
				wantJSON := formattedJSONValue(wantVal)

				t.Errorf(
					"wrong result\nexpr: %s\ngot:  %s\nwant: %s",
					test.Expr, gotJSON, wantJSON,
				)
			}
		})
	}
}

func formattedJSONValue(val cty.Value) string {
	val = cty.UnknownAsNull(val) // since JSON can't represent unknowns
	j, err := ctyjson.Marshal(val, val.Type())
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	json.Indent(&buf, j, "", "  ")
	return buf.String()
}
