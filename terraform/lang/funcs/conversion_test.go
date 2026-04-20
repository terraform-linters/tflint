// SPDX-License-Identifier: MPL-2.0

package funcs

import (
	"fmt"
	"testing"

	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang/marks"
	"github.com/zclconf/go-cty/cty"
)

func TestMakeToFunc(t *testing.T) {
	tests := []struct {
		name     string
		value    cty.Value
		targetTy cty.Type
		want     cty.Value
		wantErr  string
	}{
		{
			name:     "preserves matching type",
			value:    cty.StringVal("hello"),
			targetTy: cty.String,
			want:     cty.StringVal("hello"),
		},
		{
			name:     "converts null dynamic to typed null",
			value:    cty.NullVal(cty.DynamicPseudoType),
			targetTy: cty.String,
			want:     cty.NullVal(cty.String),
		},
		{
			name:     "preserves unknown marks",
			value:    cty.UnknownVal(cty.String).Mark(marks.Ephemeral).Mark("trace"),
			targetTy: cty.Bool,
			want:     cty.UnknownVal(cty.Bool).Mark(marks.Ephemeral).Mark("trace"),
		},
		{
			name:     "converts tuple to list",
			value:    cty.TupleVal([]cty.Value{cty.StringVal("hello"), cty.True}),
			targetTy: cty.List(cty.String),
			want: cty.ListVal([]cty.Value{
				cty.StringVal("hello"),
				cty.StringVal("true"),
			}),
		},
		{
			name:     "preserves nested marks",
			value:    cty.ObjectVal(map[string]cty.Value{"value": cty.StringVal("world").Mark("nested")}).Mark("outer"),
			targetTy: cty.Map(cty.String),
			want: cty.MapVal(map[string]cty.Value{
				"value": cty.StringVal("world").Mark("nested"),
			}).Mark("outer"),
		},
		{
			name:     "string to bool error",
			value:    cty.StringVal("nope"),
			targetTy: cty.Bool,
			wantErr:  `cannot convert "nope" to bool; only the strings "true" or "false" are allowed`,
		},
		{
			name:     "sensitive string to number error",
			value:    cty.StringVal("secret").Mark(marks.Sensitive),
			targetTy: cty.Number,
			wantErr:  `cannot convert this sensitive string to number`,
		},
		{
			name:     "object shape mismatch",
			value:    cty.EmptyObjectVal,
			targetTy: cty.Object(map[string]cty.Type{"foo": cty.String}),
			wantErr:  `incompatible object type for conversion: attribute "foo" is required`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := MakeToFunc(test.targetTy).Call([]cty.Value{test.value})
			assertCtyResult(t, got, err, test.want, test.wantErr)
		})
	}
}

func TestEphemeralAsNull(t *testing.T) {
	tests := []struct {
		name  string
		input cty.Value
		want  cty.Value
	}{
		{
			name:  "simple value becomes null",
			input: cty.StringVal("127.0.0.1:12654").Mark(marks.Ephemeral),
			want:  cty.NullVal(cty.String),
		},
		{
			name:  "unknown ephemeral stays unknown",
			input: cty.UnknownVal(cty.String).RefineNotNull().Mark(marks.Ephemeral),
			want:  cty.UnknownVal(cty.String),
		},
		{
			name: "nested values are rewritten recursively",
			input: cty.ObjectVal(map[string]cty.Value{
				"stable": cty.StringVal("hello"),
				"secret": cty.StringVal("world").Mark(marks.Ephemeral).Mark("trace"),
			}),
			want: cty.ObjectVal(map[string]cty.Value{
				"stable": cty.StringVal("hello"),
				"secret": cty.NullVal(cty.String).Mark("trace"),
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := EphemeralAsNull(test.input)
			assertCtyResult(t, got, err, test.want, "")
		})
	}
}

func assertCtyResult(t *testing.T, got cty.Value, err error, want cty.Value, wantErr string) {
	t.Helper()

	if wantErr != "" {
		if err == nil {
			t.Fatal("succeeded; want error")
		}
		if err.Error() != wantErr {
			t.Fatalf("wrong error\ngot:  %s\nwant: %s", err.Error(), wantErr)
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !got.RawEquals(want) {
		t.Fatalf("wrong result\ngot:  %s\nwant: %s", formatCtyValue(got), formatCtyValue(want))
	}
}

func formatCtyValue(value cty.Value) string {
	return fmt.Sprintf("%#v", value)
}
