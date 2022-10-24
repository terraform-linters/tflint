package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/zclconf/go-cty/cty"
)

func TestEvaluateExpr(t *testing.T) {
	// default error check helper
	neverHappend := func(diags hcl.Diagnostics) bool { return diags.HasErrors() }

	expr := func(in string) hcl.Expression {
		expr, diags := hclsyntax.ParseExpression([]byte(in), "", hcl.InitialPos)
		if diags.HasErrors() {
			t.Fatal(diags)
		}
		return expr
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	cwd = filepath.ToSlash(cwd)

	tests := []struct {
		name     string
		config   string
		inputs   []InputValues
		expr     hcl.Expression
		ty       cty.Type
		keyData  InstanceKeyEvalData
		want     string
		errCheck func(hcl.Diagnostics) bool
	}{
		{
			name:     "string literal",
			expr:     expr(`"literal_val"`),
			ty:       cty.String,
			want:     `cty.StringVal("literal_val")`,
			errCheck: neverHappend,
		},
		{
			name: "string interpolation",
			config: `
variable "string_var" {
  default = "string_val"
}`,
			expr:     expr(`var.string_var`),
			ty:       cty.String,
			want:     `cty.StringVal("string_val")`,
			errCheck: neverHappend,
		},
		{
			name: "string interpolation (legacy-style)",
			config: `
variable "string_var" {
  default = "string_val"
}`,
			expr:     expr(`"${var.string_var}"`),
			ty:       cty.String,
			want:     `cty.StringVal("string_val")`,
			errCheck: neverHappend,
		},
		{
			name: "list element",
			config: `
variable "list_var" {
  default = ["one", "two"]
}`,
			expr:     expr(`var.list_var[0]`),
			ty:       cty.String,
			want:     `cty.StringVal("one")`,
			errCheck: neverHappend,
		},
		{
			name: "map element",
			config: `
variable "map_var" {
  default = {
    one = "one"
    two = "two"
  }
}`,
			expr:     expr(`var.map_var["one"]`),
			ty:       cty.String,
			want:     `cty.StringVal("one")`,
			errCheck: neverHappend,
		},
		{
			name: "object item",
			config: `
variable "object" {
  type = object({ foo = string })
  default = { foo = "bar" }
}`,
			expr:     expr(`var.object.foo`),
			ty:       cty.String,
			want:     `cty.StringVal("bar")`,
			errCheck: neverHappend,
		},
		{
			name: "convert to string from integer",
			config: `
variable "string_var" {
  default = 10
}`,
			expr:     expr(`var.string_var`),
			ty:       cty.String,
			want:     `cty.StringVal("10")`,
			errCheck: neverHappend,
		},
		{
			name:     "conditional",
			expr:     expr(`true ? "production" : "development"`),
			ty:       cty.String,
			want:     `cty.StringVal("production")`,
			errCheck: neverHappend,
		},
		{
			name:     "built-in function",
			expr:     expr(`md5("foo")`),
			ty:       cty.String,
			want:     `cty.StringVal("acbd18db4cc2f85cedef654fccc4a4d8")`,
			errCheck: neverHappend,
		},
		{
			name:     "terraform workspace",
			expr:     expr(`terraform.workspace`),
			ty:       cty.String,
			want:     `cty.StringVal("default")`,
			errCheck: neverHappend,
		},
		{
			name: "interpolation in string",
			config: `
variable "string_var" {
  default = "World"
}`,
			expr:     expr(`"Hello ${var.string_var}"`),
			ty:       cty.String,
			want:     `cty.StringVal("Hello World")`,
			errCheck: neverHappend,
		},
		{
			name:     "path.root",
			expr:     expr(`path.root`),
			ty:       cty.String,
			want:     `cty.StringVal(".")`,
			errCheck: neverHappend,
		},
		{
			name:     "path.module",
			expr:     expr(`path.module`),
			ty:       cty.String,
			want:     `cty.StringVal(".")`,
			errCheck: neverHappend,
		},
		{
			name:     "path.cwd",
			expr:     expr(`path.cwd`),
			ty:       cty.String,
			want:     fmt.Sprintf(`cty.StringVal("%s")`, cwd),
			errCheck: neverHappend,
		},
		{
			name: "integer interpolation",
			config: `
variable "integer_var" {
  default = 3
}`,
			expr:     expr(`var.integer_var`),
			ty:       cty.Number,
			want:     `cty.NumberIntVal(3)`,
			errCheck: neverHappend,
		},
		{
			name: "convert to integer from string",
			config: `
variable "integer_var" {
  default = "3"
}`,
			expr:     expr(`var.integer_var`),
			ty:       cty.Number,
			want:     `cty.NumberIntVal(3)`,
			errCheck: neverHappend,
		},
		{
			name:     "string list literal",
			expr:     expr(`["one", "two", "three"]`),
			ty:       cty.List(cty.String),
			want:     `cty.ListVal([]cty.Value{cty.StringVal("one"), cty.StringVal("two"), cty.StringVal("three")})`,
			errCheck: neverHappend,
		},
		{
			name:     "number list literal",
			expr:     expr(`[1, 2, 3]`),
			ty:       cty.List(cty.Number),
			want:     `cty.ListVal([]cty.Value{cty.NumberIntVal(1), cty.NumberIntVal(2), cty.NumberIntVal(3)})`,
			errCheck: neverHappend,
		},
		{
			name:     "string map literal",
			expr:     expr(`{ one = 1, two = "2" }`),
			ty:       cty.Map(cty.String),
			want:     `cty.MapVal(map[string]cty.Value{"one":cty.StringVal("1"), "two":cty.StringVal("2")})`,
			errCheck: neverHappend,
		},
		{
			name:     "number map literal",
			expr:     expr(`{ one = 1, two = "2" }`),
			ty:       cty.Map(cty.Number),
			want:     `cty.MapVal(map[string]cty.Value{"one":cty.NumberIntVal(1), "two":cty.NumberIntVal(2)})`,
			errCheck: neverHappend,
		},
		{
			name:     "map object literal",
			expr:     expr(`{ one = 1, two = "2" }`),
			ty:       cty.DynamicPseudoType,
			want:     `cty.ObjectVal(map[string]cty.Value{"one":cty.NumberIntVal(1), "two":cty.StringVal("2")})`,
			errCheck: neverHappend,
		},
		{
			name: "undefined variable",
			expr: expr(`var.undefined_var`),
			ty:   cty.String,
			want: `cty.UnknownVal(cty.String)`,
			errCheck: func(diags hcl.Diagnostics) bool {
				return diags.Error() != `:1,1-18: Reference to undeclared input variable; An input variable with the name "undefined_var" has not been declared. This variable can be declared with a variable "undefined_var" {} block.`
			},
		},
		{
			name:     "no defualt variable",
			config:   `variable "no_value_var" {}`,
			expr:     expr(`var.no_value_var`),
			ty:       cty.String,
			want:     `cty.UnknownVal(cty.String)`,
			errCheck: neverHappend,
		},
		{
			name: "null value",
			config: `
variable "null_var" {
  type    = string
  default = null
}`,
			expr:     expr(`var.null_var`),
			ty:       cty.String,
			want:     `cty.NullVal(cty.String)`,
			errCheck: neverHappend,
		},
		{
			name: "terraform env",
			expr: expr(`terraform.env`),
			ty:   cty.String,
			want: `cty.UnknownVal(cty.String)`,
			errCheck: func(diags hcl.Diagnostics) bool {
				return diags.Error() != `:1,1-14: Invalid "terraform" attribute; The terraform.env attribute was deprecated in v0.10 and removed in v0.12. The "state environment" concept was renamed to "workspace" in v0.12, and so the workspace name can now be accessed using the terraform.workspace attribute.`
			},
		},
		{
			name: "type mismatch",
			expr: expr(`["one", "two", "three"]`),
			ty:   cty.String,
			want: `cty.UnknownVal(cty.String)`,
			errCheck: func(diags hcl.Diagnostics) bool {
				return diags.Error() != `:1,1-24: Incorrect value type; Invalid expression value: string required.`
			},
		},
		{
			name:     "unevaluable",
			expr:     expr(`module.text`),
			ty:       cty.String,
			want:     `cty.UnknownVal(cty.String)`,
			errCheck: neverHappend,
		},
		{
			name: "undefined variable in map",
			expr: expr(`{ value = var.undefined_var }`),
			ty:   cty.Map(cty.String),
			want: `cty.UnknownVal(cty.Map(cty.String))`,
			errCheck: func(diags hcl.Diagnostics) bool {
				return diags.Error() != `:1,11-28: Reference to undeclared input variable; An input variable with the name "undefined_var" has not been declared. This variable can be declared with a variable "undefined_var" {} block.`
			},
		},
		{
			name:     "no default value in map",
			config:   `variable "no_value_var" {}`,
			expr:     expr(`{ value = var.no_value_var }`),
			ty:       cty.Map(cty.String),
			want:     `cty.MapVal(map[string]cty.Value{"value":cty.UnknownVal(cty.String)})`,
			errCheck: neverHappend,
		},
		{
			name:     "interpolation with no default value",
			config:   `variable "no_value_var" {}`,
			expr:     expr(`"Hello, ${var.no_value_var}"`),
			ty:       cty.String,
			want:     `cty.UnknownVal(cty.String)`,
			errCheck: neverHappend,
		},
		{
			name: "null value in map",
			config: `
variable "null_var" {
  type    = string
  default = null
}`,
			expr:     expr(`{ value = var.null_var }`),
			ty:       cty.Map(cty.String),
			want:     `cty.MapVal(map[string]cty.Value{"value":cty.NullVal(cty.String)})`,
			errCheck: neverHappend,
		},
		{
			name:     "unevalauble in map",
			expr:     expr(`{ value = module.text }`),
			ty:       cty.Map(cty.String),
			want:     `cty.MapVal(map[string]cty.Value{"value":cty.UnknownVal(cty.String)})`,
			errCheck: neverHappend,
		},
		{
			name:     "interpolation with unevalauble value",
			expr:     expr(`"Hello, ${module.text}"`),
			ty:       cty.String,
			want:     `cty.UnknownVal(cty.String)`,
			errCheck: neverHappend,
		},
		{
			name: "simple override",
			config: `
variable "instance_type" {
  default = "t2.micro"
}`,
			inputs: []InputValues{
				{
					"instance_type": {Value: cty.StringVal("m4.large")},
				},
			},
			expr:     expr(`var.instance_type`),
			ty:       cty.String,
			want:     `cty.StringVal("m4.large")`,
			errCheck: neverHappend,
		},
		{
			name: "optional object attributes set to null",
			config: `
variable "optional_object" {
  type = object({
    a = optional(string)
    b = optional(string)
  })
}`,
			inputs: []InputValues{
				{
					"optional_object": {
						Value: cty.ObjectVal(map[string]cty.Value{"a": cty.StringVal("foo")}),
					},
				},
			},
			expr:     expr(`coalesce(var.optional_object.b, "baz")`),
			ty:       cty.String,
			want:     `cty.StringVal("baz")`,
			errCheck: neverHappend,
		},
		{
			name: "nullable value set to null",
			config: `
variable "foo" {
  default = "bar"
}`,
			inputs: []InputValues{
				{
					"foo": {Value: cty.NullVal(cty.String)},
				},
			},
			expr:     expr(`coalesce(var.foo, "baz")`),
			ty:       cty.String,
			want:     `cty.StringVal("baz")`,
			errCheck: neverHappend,
		},
		{
			name: "non-nullable value ignores null",
			config: `
variable "foo" {
  nullable = false
  default  = "bar"
}`,
			inputs: []InputValues{
				{
					"foo": {Value: cty.NullVal(cty.String)},
				},
			},
			expr:     expr(`coalesce(var.foo, "baz")`),
			ty:       cty.String,
			want:     `cty.StringVal("bar")`,
			errCheck: neverHappend,
		},
		{
			name: "module variable optional attributes",
			config: `
variable "foo" {
  type = object({
    required = string
	optional = optional(string)
	default  = optional(bool, true)
  })
  default = {
    required = "boop"
  }
}`,
			expr: expr(`var.foo`),
			ty: cty.Object(map[string]cty.Type{
				"required": cty.String,
				"optional": cty.String,
				"default":  cty.Bool,
			}),
			want:     `cty.ObjectVal(map[string]cty.Value{"default":cty.True, "optional":cty.NullVal(cty.String), "required":cty.StringVal("boop")})`,
			errCheck: neverHappend,
		},
		{
			name: "module variable optional attributes without default",
			config: `
variable "foo" {
  type = object({
    required = string
    optional = optional(string)
    default  = optional(bool, true)
  })
}`,
			expr: expr(`var.foo`),
			ty: cty.Object(map[string]cty.Type{
				"required": cty.String,
				"optional": cty.String,
				"default":  cty.Bool,
			}),
			want:     `cty.UnknownVal(cty.Object(map[string]cty.Type{"default":cty.Bool, "optional":cty.String, "required":cty.String}))`,
			errCheck: neverHappend,
		},
		{
			name: "module variable optional attributes with inputs",
			config: `
variable "foo" {
  type = object({
    required = string
    optional = optional(string)
    default  = optional(bool, true)
  })
}`,
			inputs: []InputValues{
				{
					"foo": {Value: cty.ObjectVal(map[string]cty.Value{"required": cty.StringVal("boop")})},
				},
			},
			expr: expr(`var.foo`),
			ty: cty.Object(map[string]cty.Type{
				"required": cty.String,
				"optional": cty.String,
				"default":  cty.Bool,
			}),
			want:     `cty.ObjectVal(map[string]cty.Value{"default":cty.True, "optional":cty.NullVal(cty.String), "required":cty.StringVal("boop")})`,
			errCheck: neverHappend,
		},
		{
			name: "module variable optional attributes with null",
			config: `
variable "foo" {
  type = object({
    required = string
    optional = optional(string)
    default  = optional(bool, true)
  })
  default = null
}`,
			expr: expr(`var.foo`),
			ty: cty.Object(map[string]cty.Type{
				"required": cty.String,
				"optional": cty.String,
				"default":  cty.Bool,
			}),
			want:     `cty.NullVal(cty.Object(map[string]cty.Type{"default":cty.Bool, "optional":cty.String, "required":cty.String}))`,
			errCheck: neverHappend,
		},
		{
			name: "module variable optional attributes with null inputs",
			config: `
variable "foo" {
  type = object({
    required = string
    optional = optional(string)
    default  = optional(bool, true)
  })
}`,
			inputs: []InputValues{
				{
					"foo": {Value: cty.NullVal(cty.Object(map[string]cty.Type{
						"required": cty.String,
						"optional": cty.String,
						"default":  cty.Bool,
					}))},
				},
			},
			expr: expr(`var.foo`),
			ty: cty.Object(map[string]cty.Type{
				"required": cty.String,
				"optional": cty.String,
				"default":  cty.Bool,
			}),
			want:     `cty.NullVal(cty.Object(map[string]cty.Type{"default":cty.Bool, "optional":cty.String, "required":cty.String}))`,
			errCheck: neverHappend,
		},
		{
			name:     "static local value",
			config:   `locals { foo = "bar" }`,
			expr:     expr(`local.foo`),
			ty:       cty.String,
			want:     `cty.StringVal("bar")`,
			errCheck: neverHappend,
		},
		{
			name: "local value using variables",
			config: `
variable "bar" {
  default = "baz"
}
locals {
  foo = var.bar
}`,
			expr:     expr(`local.foo`),
			ty:       cty.String,
			want:     `cty.StringVal("baz")`,
			errCheck: neverHappend,
		},
		{
			name: "local value using other locals",
			config: `
locals {
  foo = local.bar
  bar = "baz"
}`,
			expr:     expr(`local.foo`),
			ty:       cty.String,
			want:     `cty.StringVal("baz")`,
			errCheck: neverHappend,
		},
		{
			name: "local value using unknown value",
			config: `
locals {
  foo = module.meta.output
}`,
			expr:     expr(`local.foo`),
			ty:       cty.String,
			want:     `cty.UnknownVal(cty.String)`,
			errCheck: neverHappend,
		},
		{
			name:   "self-referencing local value",
			config: `locals { foo = local.foo }`,
			expr:   expr(`local.foo`),
			ty:     cty.String,
			want:   `cty.UnknownVal(cty.String)`,
			errCheck: func(diags hcl.Diagnostics) bool {
				return diags.Error() != `main.tf:1,16-25: circular reference found; local.foo -> local.foo`
			},
		},
		{
			name: "circular-referencing local value",
			config: `
locals {
  foo = local.bar
  bar = local.foo
}`,
			expr: expr(`local.foo`),
			ty:   cty.String,
			want: `cty.UnknownVal(cty.String)`,
			errCheck: func(diags hcl.Diagnostics) bool {
				return diags.Error() != `main.tf:4,9-18: circular reference found; local.foo -> local.bar -> local.foo`
			},
		},
		{
			name: "nested multiple local values",
			config: `
locals {
  foo = "foo"
  bar = [local.foo, local.foo]
}`,
			expr:     expr(`local.bar`),
			ty:       cty.List(cty.String),
			want:     `cty.ListVal([]cty.Value{cty.StringVal("foo"), cty.StringVal("foo")})`,
			errCheck: neverHappend,
		},
		{
			name:     "count.index in non-counted context",
			expr:     expr(`count.index`),
			ty:       cty.Number,
			want:     `cty.UnknownVal(cty.Number)`,
			errCheck: neverHappend,
		},
		{
			name:     "count.index in counted context",
			expr:     expr(`count.index`),
			ty:       cty.Number,
			keyData:  InstanceKeyEvalData{CountIndex: cty.NumberIntVal(1)},
			want:     `cty.NumberIntVal(1)`,
			errCheck: neverHappend,
		},
		{
			name:     "each.key in non-forEach context",
			expr:     expr(`each.key`),
			ty:       cty.String,
			want:     `cty.UnknownVal(cty.String)`,
			errCheck: neverHappend,
		},
		{
			name:     "each.key in forEach context",
			expr:     expr(`each.key`),
			ty:       cty.String,
			keyData:  InstanceKeyEvalData{EachKey: cty.StringVal("foo"), EachValue: cty.StringVal("bar")},
			want:     `cty.StringVal("foo")`,
			errCheck: neverHappend,
		},
		{
			name:     "each.value in non-forEach context",
			expr:     expr(`each.value`),
			ty:       cty.String,
			want:     `cty.UnknownVal(cty.String)`,
			errCheck: neverHappend,
		},
		{
			name:     "each.value in forEach context",
			expr:     expr(`each.value`),
			ty:       cty.String,
			keyData:  InstanceKeyEvalData{EachKey: cty.StringVal("foo"), EachValue: cty.StringVal("bar")},
			want:     `cty.StringVal("bar")`,
			errCheck: neverHappend,
		},
		{
			name:     "bound expr without key data",
			expr:     hclext.BindValue(cty.StringVal("foo"), expr(`each.value`)),
			ty:       cty.String,
			want:     `cty.StringVal("foo")`,
			errCheck: neverHappend,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			if err := fs.WriteFile("main.tf", []byte(test.config), os.ModePerm); err != nil {
				t.Fatal(err)
			}

			parser := NewParser(fs)
			mod, diags := parser.LoadConfigDir(".")
			if diags.HasErrors() {
				t.Fatal(diags)
			}
			config, diags := BuildConfig(mod, ModuleWalkerFunc(func(req *ModuleRequest) (*Module, *version.Version, hcl.Diagnostics) { return nil, nil, nil }))
			if diags.HasErrors() {
				t.Fatal(diags)
			}
			variableValues, diags := VariableValues(config, test.inputs...)
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			evaluator := &Evaluator{
				Meta:           &ContextMeta{Env: Workspace()},
				ModulePath:     config.Path.UnkeyedInstanceShim(),
				Config:         config,
				VariableValues: variableValues,
				CallStack:      NewCallStack(),
			}

			got, diags := evaluator.EvaluateExpr(test.expr, test.ty, test.keyData)
			if test.errCheck(diags) {
				t.Fatal(diags)
			}

			if test.want != got.GoString() {
				t.Errorf("want: %s, got: %s", test.want, got.GoString())
			}
		})
	}
}
