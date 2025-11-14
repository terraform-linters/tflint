package terraform

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

func TestDefaultVariableValues(t *testing.T) {
	tests := []struct {
		name      string
		variables map[string]*Variable
		want      InputValues
	}{
		{
			name: "basic",
			variables: map[string]*Variable{
				"default":      {Name: "default", Type: cty.String, Default: cty.StringVal("default")},
				"no_default":   {Name: "no_default", Type: cty.String},
				"null_default": {Name: "null_default", Type: cty.String, Default: cty.NullVal(cty.String)},
			},
			want: InputValues{
				"default":      {Value: cty.StringVal("default")},
				"no_default":   {Value: cty.UnknownVal(cty.String)},
				"null_default": {Value: cty.NullVal(cty.String)},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := DefaultVariableValues(test.variables)

			opt := cmp.Comparer(func(x, y cty.Value) bool {
				return x.RawEquals(y)
			})
			if diff := cmp.Diff(test.want, got, opt); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestEnvironmentVariableValues(t *testing.T) {
	neverHappend := func(diags hcl.Diagnostics) bool { return diags.HasErrors() }

	tests := []struct {
		name     string
		declared map[string]*Variable
		env      map[string]string
		want     InputValues
		errCheck func(hcl.Diagnostics) bool
	}{
		{
			name:     "undeclared",
			declared: map[string]*Variable{},
			env: map[string]string{
				"TF_VAR_instance_type": "t2.micro",
				"TF_VAR_count":         "5",
				"TF_VAR_list":          "[\"foo\"]",
				"TF_VAR_map":           "{foo=\"bar\"}",
			},
			want: InputValues{
				"instance_type": &InputValue{
					Value: cty.StringVal("t2.micro"),
				},
				"count": &InputValue{
					Value: cty.StringVal("5"),
				},
				"list": &InputValue{
					Value: cty.StringVal("[\"foo\"]"),
				},
				"map": &InputValue{
					Value: cty.StringVal("{foo=\"bar\"}"),
				},
			},
			errCheck: neverHappend,
		},
		{
			name: "declared",
			declared: map[string]*Variable{
				"instance_type": {ParsingMode: VariableParseLiteral},
				"count":         {ParsingMode: VariableParseHCL},
				"list":          {ParsingMode: VariableParseHCL},
				"map":           {ParsingMode: VariableParseHCL},
			},
			env: map[string]string{
				"TF_VAR_instance_type": "t2.micro",
				"TF_VAR_count":         "5",
				"TF_VAR_list":          "[\"foo\"]",
				"TF_VAR_map":           "{foo=\"bar\"}",
			},
			want: InputValues{
				"instance_type": &InputValue{
					Value: cty.StringVal("t2.micro"),
				},
				"count": &InputValue{
					Value: cty.NumberIntVal(5),
				},
				"list": &InputValue{
					Value: cty.TupleVal([]cty.Value{cty.StringVal("foo")}),
				},
				"map": &InputValue{
					Value: cty.ObjectVal(map[string]cty.Value{"foo": cty.StringVal("bar")}),
				},
			},
			errCheck: neverHappend,
		},
		{
			name: "invalid parsing mode",
			declared: map[string]*Variable{
				"foo": {ParsingMode: VariableParseHCL},
			},
			env: map[string]string{
				"TF_VAR_foo": "bar",
			},
			want: InputValues{},
			errCheck: func(diags hcl.Diagnostics) bool {
				return diags.Error() != "<value for var.foo>:1,1-4: Variables not allowed; Variables may not be used here."
			},
		},
		{
			name: "invalid expression",
			declared: map[string]*Variable{
				"foo": {ParsingMode: VariableParseHCL},
			},
			env: map[string]string{
				"TF_VAR_foo": `{"bar": "baz"`,
			},
			want: InputValues{},
			errCheck: func(diags hcl.Diagnostics) bool {
				return diags.Error() != "<value for var.foo>:1,1-2: Unterminated object constructor expression; There is no corresponding closing brace before the end of the file. This may be caused by incorrect brace nesting elsewhere in this file."
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for k, v := range test.env {
				t.Setenv(k, v)
			}

			got, diags := EnvironmentVariableValues(test.declared)
			if test.errCheck(diags) {
				t.Fatal(diags)
			}

			opt := cmp.Comparer(func(x, y cty.Value) bool {
				return x.RawEquals(y)
			})
			if diff := cmp.Diff(test.want, got, opt); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestParseVariableValues(t *testing.T) {
	neverHappend := func(diags hcl.Diagnostics) bool { return diags.HasErrors() }

	tests := []struct {
		name     string
		declared map[string]*Variable
		vars     []string
		want     InputValues
		errCheck func(hcl.Diagnostics) bool
	}{
		{
			name:     "undeclared",
			declared: map[string]*Variable{},
			vars: []string{
				"foo=bar",
			},
			want: InputValues{},
			errCheck: func(diags hcl.Diagnostics) bool {
				return diags.Error() != `<value for var.foo>:1,1-1: Value for undeclared variable; A variable named "foo" was assigned, but the root module does not declare a variable of that name.`
			},
		},
		{
			name: "declared",
			declared: map[string]*Variable{
				"foo": {ParsingMode: VariableParseLiteral},
				"bar": {ParsingMode: VariableParseHCL},
				"baz": {ParsingMode: VariableParseHCL},
			},
			vars: []string{
				"foo=bar",
				"bar=[\"foo\"]",
				"baz={ foo=\"bar\" }",
			},
			want: InputValues{
				"foo": &InputValue{
					Value: cty.StringVal("bar"),
				},
				"bar": &InputValue{
					Value: cty.TupleVal([]cty.Value{cty.StringVal("foo")}),
				},
				"baz": &InputValue{
					Value: cty.ObjectVal(map[string]cty.Value{"foo": cty.StringVal("bar")}),
				},
			},
			errCheck: neverHappend,
		},
		{
			name:     "invalid format",
			declared: map[string]*Variable{},
			vars:     []string{"foo"},
			want:     InputValues{},
			errCheck: func(diags hcl.Diagnostics) bool {
				return diags.Error() != `<input-value>:1,1-1: invalid variable value format; "foo" is invalid. Variables must be "key=value" format`
			},
		},
		{
			name: "invalid parsing mode",
			declared: map[string]*Variable{
				"foo": {ParsingMode: VariableParseHCL},
			},
			vars: []string{"foo=bar"},
			want: InputValues{},
			errCheck: func(diags hcl.Diagnostics) bool {
				return diags.Error() != "<value for var.foo>:1,1-4: Variables not allowed; Variables may not be used here."
			},
		},
		{
			name: "invalid expression",
			declared: map[string]*Variable{
				"foo": {ParsingMode: VariableParseHCL},
			},
			vars: []string{"foo="},
			want: InputValues{},
			errCheck: func(diags hcl.Diagnostics) bool {
				return diags.Error() != "<value for var.foo>:1,1-1: Missing expression; Expected the start of an expression, but found the end of the file."
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, diags := ParseVariableValues(test.vars, test.declared)
			if test.errCheck(diags) {
				t.Fatal(diags)
			}

			opt := cmp.Comparer(func(x, y cty.Value) bool {
				return x.RawEquals(y)
			})
			if diff := cmp.Diff(test.want, got, opt); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestVariableValues(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		env    map[string]string
		inputs []InputValues
		want   map[string]map[string]cty.Value
	}{
		{
			name: "basic",
			config: &Config{
				Path: []string{"child1", "child2"},
				Module: &Module{
					Variables: map[string]*Variable{
						"a": {Name: "a", Type: cty.String, ParsingMode: VariableParseLiteral, Default: cty.StringVal("config")},
						"b": {Name: "b", Type: cty.String, ParsingMode: VariableParseLiteral, Default: cty.StringVal("config")},
						"c": {Name: "c", Type: cty.String, ParsingMode: VariableParseLiteral, Default: cty.StringVal("config")},
						"d": {Name: "d", Type: cty.String, ParsingMode: VariableParseLiteral, Default: cty.StringVal("config")},
					},
				},
			},
			env: map[string]string{
				"TF_VAR_a": "env",
				"TF_VAR_b": "env",
				"TF_VAR_c": "env",
			},
			inputs: []InputValues{
				{
					"a": {Value: cty.StringVal("input1")},
					"b": {Value: cty.StringVal("input1")},
				},
				{
					"a": {Value: cty.StringVal("input2")},
				},
			},
			want: map[string]map[string]cty.Value{
				"module.child1.module.child2": {
					"a": cty.StringVal("input2"),
					"b": cty.StringVal("input1"),
					"c": cty.StringVal("config"), // Environment variables don't apply to child modules
					"d": cty.StringVal("config"),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for k, v := range test.env {
				t.Setenv(k, v)
			}

			got, diags := VariableValues(test.config, test.inputs...)
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			opt := cmp.Comparer(func(x, y cty.Value) bool {
				return x.RawEquals(y)
			})
			if diff := cmp.Diff(test.want, got, opt); diff != "" {
				t.Error(diff)
			}
		})
	}
}
