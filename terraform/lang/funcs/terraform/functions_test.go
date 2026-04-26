// SPDX-License-Identifier: MPL-2.0

package terraform

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

func TestEncodeTfvarsFunc(t *testing.T) {
	cases := []struct {
		name    string
		input   cty.Value
		want    cty.Value
		wantErr string
	}{
		{
			name: "object",
			input: cty.ObjectVal(map[string]cty.Value{
				"string": cty.StringVal("hello"),
				"number": cty.NumberIntVal(5),
				"bool":   cty.True,
				"set":    cty.SetVal([]cty.Value{cty.StringVal("beep"), cty.StringVal("boop")}),
				"list":   cty.SetVal([]cty.Value{cty.StringVal("bleep"), cty.StringVal("bloop")}),
				"tuple":  cty.SetVal([]cty.Value{cty.StringVal("bibble"), cty.StringVal("wibble")}),
				"map":    cty.MapVal(map[string]cty.Value{"one": cty.NumberIntVal(1)}),
				"object": cty.ObjectVal(map[string]cty.Value{"one": cty.NumberIntVal(1), "true": cty.True}),
				"null":   cty.NullVal(cty.String),
			}),
			want: cty.StringVal(`bool = true
list = ["bleep", "bloop"]
map = {
  one = 1
}
null   = null
number = 5
object = {
  one  = 1
  true = true
}
set    = ["beep", "boop"]
string = "hello"
tuple  = ["bibble", "wibble"]
`),
		},
		{name: "empty object", input: cty.EmptyObjectVal, want: cty.StringVal("")},
		{
			name: "map",
			input: cty.MapVal(map[string]cty.Value{
				"one":   cty.NumberIntVal(1),
				"two":   cty.NumberIntVal(2),
				"three": cty.NumberIntVal(3),
			}),
			want: cty.StringVal("one   = 1\nthree = 3\ntwo   = 2\n"),
		},
		{name: "empty map", input: cty.MapValEmpty(cty.String), want: cty.StringVal("")},
		{name: "unknown object", input: cty.UnknownVal(cty.EmptyObject), want: cty.UnknownVal(cty.String).RefineNotNull()},
		{name: "unknown map", input: cty.UnknownVal(cty.Map(cty.String)), want: cty.UnknownVal(cty.String).RefineNotNull()},
		{
			name: "partially unknown object",
			input: cty.ObjectVal(map[string]cty.Value{
				"string": cty.UnknownVal(cty.String),
			}),
			want: cty.UnknownVal(cty.String).RefineNotNull(),
		},
		{
			name: "partially unknown map",
			input: cty.MapVal(map[string]cty.Value{
				"string": cty.UnknownVal(cty.String),
			}),
			want: cty.UnknownVal(cty.String).RefineNotNull(),
		},
		{name: "null object", input: cty.NullVal(cty.EmptyObject), wantErr: "cannot encode a null value in tfvars syntax"},
		{name: "null map", input: cty.NullVal(cty.Map(cty.String)), wantErr: "cannot encode a null value in tfvars syntax"},
		{name: "string input", input: cty.StringVal("nope"), wantErr: "invalid value to encode: must be an object whose attribute names will become the encoded variable names"},
		{name: "number input", input: cty.Zero, wantErr: "invalid value to encode: must be an object whose attribute names will become the encoded variable names"},
		{name: "bool input", input: cty.False, wantErr: "invalid value to encode: must be an object whose attribute names will become the encoded variable names"},
		{name: "list input", input: cty.ListValEmpty(cty.String), wantErr: "invalid value to encode: must be an object whose attribute names will become the encoded variable names"},
		{name: "set input", input: cty.SetValEmpty(cty.String), wantErr: "invalid value to encode: must be an object whose attribute names will become the encoded variable names"},
		{name: "tuple input", input: cty.EmptyTupleVal, wantErr: "invalid value to encode: must be an object whose attribute names will become the encoded variable names"},
		{
			name: "invalid identifier",
			input: cty.ObjectVal(map[string]cty.Value{
				"not valid identifier": cty.StringVal("!"),
			}),
			wantErr: `invalid variable name "not valid identifier": must be a valid identifier, per Terraform's rules for input variable declarations`,
		},
	}

	runSingleArgCases(t, EncodeTfvarsFunc, cases)
}

func TestDecodeTfvarsFunc(t *testing.T) {
	cases := []struct {
		name    string
		input   cty.Value
		want    cty.Value
		wantErr string
	}{
		{
			name:  "simple values",
			input: cty.StringVal("string = \"hello\"\nnumber = 2"),
			want: cty.ObjectVal(map[string]cty.Value{
				"string": cty.StringVal("hello"),
				"number": cty.NumberIntVal(2),
			}),
		},
		{name: "empty source", input: cty.StringVal(""), want: cty.EmptyObjectVal},
		{name: "unknown source", input: cty.UnknownVal(cty.String), want: cty.UnknownVal(cty.DynamicPseudoType)},
		{name: "null source", input: cty.NullVal(cty.String), wantErr: "cannot decode tfvars from a null value"},
		{
			name:    "invalid syntax",
			input:   cty.StringVal("not valid syntax"),
			wantErr: `invalid tfvars syntax: <decode_tfvars argument>:1,17-17: Invalid block definition; Either a quoted string block label or an opening brace ("{") is expected here.`,
		},
		{
			name:    "missing newline",
			input:   cty.StringVal("foo = not valid syntax"),
			wantErr: `invalid tfvars syntax: <decode_tfvars argument>:1,11-16: Missing newline after argument; An argument definition must end with a newline.`,
		},
		{
			name:    "variable reference rejected",
			input:   cty.StringVal("foo = var.whatever"),
			wantErr: `invalid expression for variable "foo": <decode_tfvars argument>:1,7-10: Variables not allowed; Variables may not be used here.`,
		},
		{
			name:    "function call rejected",
			input:   cty.StringVal("foo = whatever()"),
			wantErr: `invalid expression for variable "foo": <decode_tfvars argument>:1,7-17: Function calls not allowed; Functions may not be called here.`,
		},
	}

	runSingleArgCases(t, DecodeTfvarsFunc, cases)
}

func TestEncodeExprFunc(t *testing.T) {
	cases := []struct {
		name    string
		input   cty.Value
		want    cty.Value
		wantErr string
	}{
		{name: "string", input: cty.StringVal("hello"), want: cty.StringVal(`"hello"`)},
		{name: "string with newlines", input: cty.StringVal("hello\nworld\n"), want: cty.StringVal(`"hello\nworld\n"`)},
		{name: "string with template interpolation", input: cty.StringVal("hel${lo"), want: cty.StringVal(`"hel$${lo"`)},
		{name: "string with template control", input: cty.StringVal("hel%{lo"), want: cty.StringVal(`"hel%%{lo"`)},
		{name: "string with backslash", input: cty.StringVal(`boop\boop`), want: cty.StringVal(`"boop\\boop"`)},
		{name: "empty string", input: cty.StringVal(""), want: cty.StringVal(`""`)},
		{name: "number", input: cty.NumberIntVal(2), want: cty.StringVal("2")},
		{name: "true", input: cty.True, want: cty.StringVal("true")},
		{name: "false", input: cty.False, want: cty.StringVal("false")},
		{name: "empty object", input: cty.EmptyObjectVal, want: cty.StringVal(`{}`)},
		{
			name: "object",
			input: cty.ObjectVal(map[string]cty.Value{
				"number": cty.NumberIntVal(5),
				"string": cty.StringVal("..."),
			}),
			want: cty.StringVal("{\n  number = 5\n  string = \"...\"\n}"),
		},
		{
			name: "map",
			input: cty.MapVal(map[string]cty.Value{
				"one": cty.NumberIntVal(1),
				"two": cty.NumberIntVal(2),
			}),
			want: cty.StringVal("{\n  one = 1\n  two = 2\n}"),
		},
		{name: "empty tuple", input: cty.EmptyTupleVal, want: cty.StringVal(`[]`)},
		{
			name: "tuple",
			input: cty.TupleVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.StringVal("..."),
			}),
			want: cty.StringVal(`[5, "..."]`),
		},
		{
			name: "set",
			input: cty.SetVal([]cty.Value{
				cty.NumberIntVal(1),
				cty.NumberIntVal(5),
				cty.NumberIntVal(20),
				cty.NumberIntVal(55),
			}),
			want: cty.StringVal(`[1, 5, 20, 55]`),
		},
		{name: "dynamic value", input: cty.DynamicVal, want: cty.UnknownVal(cty.String).RefineNotNull()},
		{name: "unknown number", input: cty.UnknownVal(cty.Number).RefineNotNull(), want: cty.UnknownVal(cty.String).RefineNotNull()},
		{
			name:  "unknown string",
			input: cty.UnknownVal(cty.String).RefineNotNull(),
			want: cty.UnknownVal(cty.String).Refine().NotNull().StringPrefixFull(`"`).NewValue(),
		},
		{
			name:  "unknown object",
			input: cty.UnknownVal(cty.EmptyObject).RefineNotNull(),
			want: cty.UnknownVal(cty.String).Refine().NotNull().StringPrefixFull(`{`).NewValue(),
		},
		{
			name:  "unknown map",
			input: cty.UnknownVal(cty.Map(cty.String)).RefineNotNull(),
			want: cty.UnknownVal(cty.String).Refine().NotNull().StringPrefixFull(`{`).NewValue(),
		},
		{
			name:  "unknown tuple",
			input: cty.UnknownVal(cty.EmptyTuple).RefineNotNull(),
			want: cty.UnknownVal(cty.String).Refine().NotNull().StringPrefixFull(`[`).NewValue(),
		},
		{
			name:  "unknown list",
			input: cty.UnknownVal(cty.List(cty.String)).RefineNotNull(),
			want: cty.UnknownVal(cty.String).Refine().NotNull().StringPrefixFull(`[`).NewValue(),
		},
		{
			name:  "unknown set",
			input: cty.UnknownVal(cty.Set(cty.String)).RefineNotNull(),
			want: cty.UnknownVal(cty.String).Refine().NotNull().StringPrefixFull(`[`).NewValue(),
		},
	}

	runSingleArgCases(t, EncodeExprFunc, cases)
}

func runSingleArgCases(t *testing.T, fn function.Function, cases []struct {
	name    string
	input   cty.Value
	want    cty.Value
	wantErr string
}) {
	t.Helper()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := fn.Call([]cty.Value{tc.input})
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("unexpected success: got %s", got.GoString())
				}
				if err.Error() != tc.wantErr {
					t.Fatalf("wrong error: got %q want %q", err.Error(), tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if diff := cmp.Diff(tc.want, got, ctydebug.CmpOptions); diff != "" {
				t.Fatalf("unexpected value (-want +got):\n%s", diff)
			}
		})
	}
}
