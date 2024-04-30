package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang/marks"
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

	originalWd, err := filepath.Abs("/foo/bar/baz")
	if err != nil {
		t.Fatal(err)
	}
	originalWd = filepath.ToSlash(originalWd)

	tests := []struct {
		name     string
		config   string
		inputs   []InputValues
		context  *ContextMeta
		expr     hcl.Expression
		ty       cty.Type
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
			name:     "built-in function with namespace",
			expr:     expr(`core::md5("foo")`),
			ty:       cty.String,
			want:     `cty.StringVal("acbd18db4cc2f85cedef654fccc4a4d8")`,
			errCheck: neverHappend,
		},
		{
			name:     "provider-defined functions",
			expr:     expr(`provider::tflint::echo("Hello", "World!")`),
			ty:       cty.String,
			want:     `cty.UnknownVal(cty.String)`,
			errCheck: neverHappend,
		},
		{
			name:     "built-in provider-defined functions",
			expr:     expr(`provider::terraform::tfvarsdecode("a = 1")`),
			ty:       cty.Object(map[string]cty.Type{"a": cty.Number}),
			want:     `cty.ObjectVal(map[string]cty.Value{"a":cty.NumberIntVal(1)})`,
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
			name:     "path.cwd with original working dir",
			context:  &ContextMeta{Env: Workspace(), OriginalWorkingDir: originalWd},
			expr:     expr(`path.cwd`),
			ty:       cty.String,
			want:     fmt.Sprintf(`cty.StringVal("%s")`, originalWd),
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
			want:     `cty.UnknownVal(cty.String).Refine().NotNull().StringPrefixFull("Hello, ").NewValue()`,
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
			want:     `cty.UnknownVal(cty.String).Refine().NotNull().StringPrefixFull("Hello, ").NewValue()`,
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
			name: "module variable optional attributes with nested default optional",
			config: `
variable "foo" {
  type = set(object({
    name      = string
    schedules = set(object({
      name               = string
      cold_storage_after = optional(number, 10)
    }))
  }))
}`,
			expr: expr(`var.foo`),
			inputs: []InputValues{
				{
					"foo": {
						Value: cty.SetVal([]cty.Value{
							cty.ObjectVal(map[string]cty.Value{
								"name": cty.StringVal("test1"),
								"schedules": cty.SetVal([]cty.Value{
									cty.MapVal(map[string]cty.Value{
										"name": cty.StringVal("daily"),
									}),
								}),
							}),
							cty.ObjectVal(map[string]cty.Value{
								"name": cty.StringVal("test2"),
								"schedules": cty.SetVal([]cty.Value{
									cty.MapVal(map[string]cty.Value{
										"name": cty.StringVal("daily"),
									}),
									cty.MapVal(map[string]cty.Value{
										"name":               cty.StringVal("weekly"),
										"cold_storage_after": cty.StringVal("0"),
									}),
								}),
							}),
						}),
					},
				},
			},
			ty: cty.Set(cty.Object(map[string]cty.Type{
				"name": cty.String,
				"schedules": cty.Set(cty.Object(map[string]cty.Type{
					"name":               cty.String,
					"cold_storage_after": cty.Number,
				})),
			})),
			want:     `cty.SetVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"name":cty.StringVal("test1"), "schedules":cty.SetVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"cold_storage_after":cty.NumberIntVal(10), "name":cty.StringVal("daily")})})}), cty.ObjectVal(map[string]cty.Value{"name":cty.StringVal("test2"), "schedules":cty.SetVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{"cold_storage_after":cty.NumberIntVal(0), "name":cty.StringVal("weekly")}), cty.ObjectVal(map[string]cty.Value{"cold_storage_after":cty.NumberIntVal(10), "name":cty.StringVal("daily")})})})})`,
			errCheck: neverHappend,
		},
		{
			name: "module variable optional attributes with nested complex types",
			config: `
variable "foo" {
  type = object({
    name                       = string
    nested_object              = object({
      name  = string
      value = optional(string, "foo")
    })
    nested_object_with_default = optional(object({
      name  = string
      value = optional(string, "bar")
    }), {
      name = "nested_object_with_default"
    })
  })
}`,
			expr: expr(`var.foo`),
			inputs: []InputValues{
				{
					"foo": {
						Value: cty.ObjectVal(map[string]cty.Value{
							"name": cty.StringVal("object"),
							"nested_object": cty.ObjectVal(map[string]cty.Value{
								"name": cty.StringVal("nested_object"),
							}),
						}),
					},
				},
			},
			ty: cty.Object(map[string]cty.Type{
				"name": cty.String,
				"nested_object": cty.Object(map[string]cty.Type{
					"name":  cty.String,
					"value": cty.String,
				}),
				"nested_object_with_default": cty.Object(map[string]cty.Type{
					"name":  cty.String,
					"value": cty.String,
				}),
			}),
			want:     `cty.ObjectVal(map[string]cty.Value{"name":cty.StringVal("object"), "nested_object":cty.ObjectVal(map[string]cty.Value{"name":cty.StringVal("nested_object"), "value":cty.StringVal("foo")}), "nested_object_with_default":cty.ObjectVal(map[string]cty.Value{"name":cty.StringVal("nested_object_with_default"), "value":cty.StringVal("bar")})})`,
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
			name:     "each.key in non-forEach context",
			expr:     expr(`each.key`),
			ty:       cty.String,
			want:     `cty.UnknownVal(cty.String)`,
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
			name:     "bound expr",
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
			mod, diags := parser.LoadConfigDir(".", ".")
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
				Meta:           test.context,
				ModulePath:     config.Path.UnkeyedInstanceShim(),
				Config:         config,
				VariableValues: variableValues,
				CallStack:      NewCallStack(),
			}
			if evaluator.Meta == nil {
				evaluator.Meta = &ContextMeta{Env: Workspace()}
			}

			got, diags := evaluator.EvaluateExpr(test.expr, test.ty)
			if test.errCheck(diags) {
				t.Fatal(diags)
			}

			if test.want != got.GoString() {
				t.Errorf("want: %s, got: %s", test.want, got.GoString())
			}
		})
	}
}

func TestExpandBlock(t *testing.T) {
	tests := []struct {
		name   string
		config string
		schema *hclext.BodySchema
		want   *hclext.BodyContent
	}{
		{
			name: "no meta-arguments",
			config: `
resource "aws_instance" "main" {}
`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{Type: "resource", Labels: []string{"aws_instance", "main"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}, Blocks: hclext.Blocks{}}},
				},
			},
		},
		{
			name: "count is not zero (literal)",
			config: `
resource "aws_instance" "main" {
  count = 1
  value = count.index
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}}},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body:   &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.NumberIntVal(0), hcl.Range{})}}, Blocks: hclext.Blocks{}},
					},
				},
			},
		},
		{
			name: "count is not zero (variable)",
			config: `
variable "count" {
  default = 1
}
resource "aws_instance" "main" {
  count = var.count
  value = count.index
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}}},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body:   &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.NumberIntVal(0), hcl.Range{})}}, Blocks: hclext.Blocks{}},
					},
				},
			},
		},
		{
			name: "count is greater than 1",
			config: `
resource "aws_instance" "main" {
  count = 2
  value = count.index
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}}},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body:   &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.NumberIntVal(0), hcl.Range{})}}, Blocks: hclext.Blocks{}},
					},
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body:   &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.NumberIntVal(1), hcl.Range{})}}, Blocks: hclext.Blocks{}},
					},
				},
			},
		},
		{
			name: "count is unknown",
			config: `
variable "count" {}

resource "aws_instance" "main" {
  count = var.count
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{Attributes: hclext.Attributes{}, Blocks: hclext.Blocks{}},
		},
		{
			name: "count is sensitive",
			config: `
variable "count" {
  sensitive = true
  default   = 1
}
resource "aws_instance" "main" {
  count = var.count
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{Type: "resource", Labels: []string{"aws_instance", "main"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}, Blocks: hclext.Blocks{}}},
				},
			},
		},
		{
			name: "count is unevaluable",
			config: `
resource "aws_instance" "main" {
  count = module.meta.count
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{Attributes: hclext.Attributes{}, Blocks: hclext.Blocks{}},
		},
		{
			name: "count is using provider-defined functions",
			config: `
resource "aws_instance" "main" {
  count = provider::tflint::count()
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{Attributes: hclext.Attributes{}, Blocks: hclext.Blocks{}},
		},
		{
			name: "count is zero",
			config: `
resource "aws_instance" "main" {
  count = 0
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{Attributes: hclext.Attributes{}, Blocks: hclext.Blocks{}},
		},
		{
			name: "count is string",
			config: `
resource "aws_instance" "main" {
  count = "1"
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{Type: "resource", Labels: []string{"aws_instance", "main"}, Body: &hclext.BodyContent{Attributes: hclext.Attributes{}, Blocks: hclext.Blocks{}}},
				},
			},
		},
		{
			name: "count.index and sensitive value",
			config: `
variable "sensitive" {
  sensitive = true
  default   = "foo"
}
resource "aws_instance" "main" {
  count = 1
  value = "${count.index}-${var.sensitive}"
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}}},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body:   &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.UnknownVal(cty.String).RefineNotNull().Mark(marks.Sensitive), hcl.Range{})}}, Blocks: hclext.Blocks{}},
					},
				},
			},
		},
		{
			name: "count.index and nested sensitive value",
			config: `
variable "sensitive" {
  sensitive = true
  default   = "foo"
}
resource "aws_instance" "main" {
  count = 1
  value = [count.index, var.sensitive]
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}}},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body:   &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.TupleVal([]cty.Value{cty.UnknownVal(cty.Number), cty.StringVal("foo").Mark(marks.Sensitive)}), hcl.Range{})}}, Blocks: hclext.Blocks{}},
					},
				},
			},
		},
		{
			name: "count.index and provider-defined functions",
			config: `
resource "aws_instance" "main" {
  count = 1
  value = [count.index, provider::tflint::sum(1, 2, 3)]
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}}},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body:   &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.TupleVal([]cty.Value{cty.NumberIntVal(0), cty.DynamicVal}), hcl.Range{})}}, Blocks: hclext.Blocks{}},
					},
				},
			},
		},
		{
			name: "for_each is not empty (literal)",
			config: `
resource "aws_instance" "main" {
  for_each = { foo = "bar" }
  value    = "${each.key}-${each.value}"
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}}},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body:   &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("foo-bar"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
					},
				},
			},
		},
		{
			name: "for_each is not empty (variable)",
			config: `
variable "for_each" {
  default = { foo = "bar" }
}
resource "aws_instance" "main" {
  for_each = var.for_each
  value    = "${each.key}-${each.value}"
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}}},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body:   &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("foo-bar"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
					},
				},
			},
		},
		{
			name: "for_each is unknown",
			config: `
variable "for_each" {}

resource "aws_instance" "main" {
  for_each = var.for_each
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{Attributes: hclext.Attributes{}, Blocks: hclext.Blocks{}},
		},
		{
			name: "for_each is unevaluable",
			config: `
resource "aws_instance" "main" {
  for_each = module.meta.for_each
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{Attributes: hclext.Attributes{}, Blocks: hclext.Blocks{}},
		},
		{
			name: "for_each is using provider-defined functions",
			config: `
resource "aws_instance" "main" {
  for_each = provider::tflint::for_each()
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{Attributes: hclext.Attributes{}, Blocks: hclext.Blocks{}},
		},
		{
			name: "for_each contains unevaluable",
			config: `
resource "aws_instance" "main" {
  for_each = {
    known   = "known"
    unknown = module.meta.unknown
  }
  value = [each.key, each.value]
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}}},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{
								"value": {
									Name: "value",
									Expr: hcl.StaticExpr(cty.TupleVal([]cty.Value{cty.StringVal("known"), cty.StringVal("known")}), hcl.Range{}),
								},
							},
							Blocks: hclext.Blocks{},
						},
					},
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{
								"value": {
									Name: "value",
									Expr: hcl.StaticExpr(cty.TupleVal([]cty.Value{cty.StringVal("unknown"), cty.DynamicVal}), hcl.Range{}),
								},
							},
							Blocks: hclext.Blocks{},
						},
					},
				},
			},
		},
		{
			name: "for_each is empty",
			config: `
resource "aws_instance" "main" {
  for_each = {}
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{Attributes: hclext.Attributes{}, Blocks: hclext.Blocks{}},
		},
		{
			name: "for_each is not empty set",
			config: `
resource "aws_instance" "main" {
  for_each = toset(["foo", "bar"])
  value    = each.key
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}}},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body:   &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("bar"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
					},
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body:   &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("foo"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
					},
				},
			},
		},
		{
			name: "for_each is empty set",
			config: `
resource "aws_instance" "main" {
  for_each = toset([])
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
				},
			},
			want: &hclext.BodyContent{Attributes: hclext.Attributes{}, Blocks: hclext.Blocks{}},
		},
		{
			name: "each.key/each.value and sensitive value",
			config: `
variable "sensitive" {
  sensitive = true
  default   = "foo"
}
resource "aws_instance" "main" {
  for_each = { foo = "bar" }
  value    = "${each.key}-${each.value}-${var.sensitive}"
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}}},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body:   &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.UnknownVal(cty.String).RefineNotNull().Mark(marks.Sensitive), hcl.Range{})}}, Blocks: hclext.Blocks{}},
					},
				},
			},
		},
		{
			name: "each.key/each.value and nested sensitive value",
			config: `
variable "sensitive" {
  sensitive = true
  default   = "foo"
}
resource "aws_instance" "main" {
  for_each = { foo = "bar" }
  value    = [each.key, var.sensitive]
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}}},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body:   &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.TupleVal([]cty.Value{cty.DynamicVal, cty.StringVal("foo").Mark(marks.Sensitive)}), hcl.Range{})}}, Blocks: hclext.Blocks{}},
					},
				},
			},
		},
		{
			name: "non-empty object dynamic blocks",
			config: `
resource "aws_instance" "main" {
  dynamic "ebs_block_device" {
    for_each = {
      foo = "bar"
      baz = "qux"
    }
    content {
      value = "${ebs_block_device.key}-${ebs_block_device.value}"
    }
  }
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body: &hclext.BodySchema{
							Blocks: []hclext.BlockSchema{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}},
								},
							},
						},
					},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks: hclext.Blocks{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("baz-qux"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
								},
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("foo-bar"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "non-empty object variable dynamic blocks",
			config: `
variable "for_each" {
  default = {
    foo = "bar"
    baz = "qux"
  }
}
resource "aws_instance" "main" {
  dynamic "ebs_block_device" {
    for_each = var.for_each
    content {
      value = "${ebs_block_device.key}-${ebs_block_device.value}"
    }
  }
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body: &hclext.BodySchema{
							Blocks: []hclext.BlockSchema{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}},
								},
							},
						},
					},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks: hclext.Blocks{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("baz-qux"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
								},
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("foo-bar"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "unknown variable dynamic blocks",
			config: `
variable "for_each" {}

resource "aws_instance" "main" {
  dynamic "ebs_block_device" {
    for_each = var.for_each
    content {
      value = "${ebs_block_device.key}-${ebs_block_device.value}"
    }
  }
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body: &hclext.BodySchema{
							Blocks: []hclext.BlockSchema{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}},
								},
							},
						},
					},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks:     hclext.Blocks{},
						},
					},
				},
			},
		},
		{
			name: "unevaluable variable dynamic blocks",
			config: `
resource "aws_instance" "main" {
  dynamic "ebs_block_device" {
    for_each = module.meta.for_each
    content {
      value = "${ebs_block_device.key}-${ebs_block_device.value}"
    }
  }
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body: &hclext.BodySchema{
							Blocks: []hclext.BlockSchema{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}},
								},
							},
						},
					},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks:     hclext.Blocks{},
						},
					},
				},
			},
		},
		{
			name: "dynamic blocks with provider-defined functions",
			config: `
resource "aws_instance" "main" {
  dynamic "ebs_block_device" {
    for_each = provider::tflint::for_each()
    content {
      value = "${ebs_block_device.key}-${ebs_block_device.value}"
    }
  }
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body: &hclext.BodySchema{
							Blocks: []hclext.BlockSchema{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}},
								},
							},
						},
					},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks:     hclext.Blocks{},
						},
					},
				},
			},
		},
		{
			name: "object contains unevaluable dynamic blocks",
			config: `
resource "aws_instance" "main" {
  dynamic "ebs_block_device" {
    for_each = {
      known   = "known"
      unknown = module.meta.unknown
    }
    content {
      value = "${ebs_block_device.key}-${ebs_block_device.value}"
    }
  }
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body: &hclext.BodySchema{
							Blocks: []hclext.BlockSchema{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}},
								},
							},
						},
					},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks: hclext.Blocks{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("known-known"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
								},
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.UnknownVal(cty.String).Refine().NotNull().StringPrefixFull("unknown-").NewValue(), hcl.Range{})}}, Blocks: hclext.Blocks{}},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "empty object dynamic blocks",
			config: `
resource "aws_instance" "main" {
  dynamic "ebs_block_device" {
    for_each = {}
    content {
      value = "${ebs_block_device.key}-${ebs_block_device.value}"
    }
  }
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body: &hclext.BodySchema{
							Blocks: []hclext.BlockSchema{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}},
								},
							},
						},
					},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks:     hclext.Blocks{},
						},
					},
				},
			},
		},
		{
			name: "non-empty set dynamic blocks",
			config: `
resource "aws_instance" "main" {
  dynamic "ebs_block_device" {
    for_each = toset(["foo", "bar"])
    content {
      value = "${ebs_block_device.key}-${ebs_block_device.value}"
    }
  }
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body: &hclext.BodySchema{
							Blocks: []hclext.BlockSchema{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}},
								},
							},
						},
					},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks: hclext.Blocks{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("bar-bar"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
								},
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("foo-foo"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "empty set dynamic blocks",
			config: `
resource "aws_instance" "main" {
  dynamic "ebs_block_device" {
    for_each = toset([])
    content {
      value = "${ebs_block_device.key}-${ebs_block_device.value}"
    }
  }
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body: &hclext.BodySchema{
							Blocks: []hclext.BlockSchema{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}},
								},
							},
						},
					},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks:     hclext.Blocks{},
						},
					},
				},
			},
		},
		{
			name: "iterator with sensitive value",
			config: `
variable "sensitive" {
  sensitive = true
  default   = "foo"
}
resource "aws_instance" "main" {
  dynamic "ebs_block_device" {
    for_each = { foo = "bar" }
    content {
      value = "${ebs_block_device.key}-${ebs_block_device.value}-${var.sensitive}"
    }
  }
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body: &hclext.BodySchema{
							Blocks: []hclext.BlockSchema{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}},
								},
							},
						},
					},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks: hclext.Blocks{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.UnknownVal(cty.String).RefineNotNull().Mark(marks.Sensitive), hcl.Range{})}}, Blocks: hclext.Blocks{}},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "iterator with nested sensitive value",
			config: `
variable "sensitive" {
  sensitive = true
  default   = "foo"
}
resource "aws_instance" "main" {
  dynamic "ebs_block_device" {
    for_each = { foo = "bar" }
    content {
      value = [ebs_block_device.key, var.sensitive]
    }
  }
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body: &hclext.BodySchema{
							Blocks: []hclext.BlockSchema{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}},
								},
							},
						},
					},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks: hclext.Blocks{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.TupleVal([]cty.Value{cty.DynamicVal, cty.StringVal("foo").Mark(marks.Sensitive)}), hcl.Range{})}}, Blocks: hclext.Blocks{}},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "nested dynamic blocks",
			config: `
resource "aws_instance" "main" {
  dynamic "ebs_block_device" {
    for_each = toset(["foo", "bar"])
    content {
      dynamic "nested" {
        for_each = toset(["baz", "qux"])
        content {
          value = "${ebs_block_device.key}-${nested.value}"
        }
      }
    }
  }
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body: &hclext.BodySchema{
							Blocks: []hclext.BlockSchema{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodySchema{
										Blocks: []hclext.BlockSchema{
											{
												Type: "nested",
												Body: &hclext.BodySchema{
													Attributes: []hclext.AttributeSchema{{Name: "value"}},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks: hclext.Blocks{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{
										Attributes: hclext.Attributes{},
										Blocks: hclext.Blocks{
											{
												Type: "nested",
												Body: &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("bar-baz"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
											},
											{
												Type: "nested",
												Body: &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("bar-qux"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
											},
										},
									},
								},
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{
										Attributes: hclext.Attributes{},
										Blocks: hclext.Blocks{
											{
												Type: "nested",
												Body: &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("foo-baz"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
											},
											{
												Type: "nested",
												Body: &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("foo-qux"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "custom iterator",
			config: `
resource "aws_instance" "main" {
  dynamic "ebs_block_device" {
    for_each = { foo = "bar" }
    iterator = it
    content {
      value = "${it.key}-${it.value}"
    }
  }
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body: &hclext.BodySchema{
							Blocks: []hclext.BlockSchema{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}},
								},
							},
						},
					},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks: hclext.Blocks{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("foo-bar"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "dynamic labels",
			config: `
resource "aws_instance" "main" {
  dynamic "ebs_block_device" {
    for_each = { foo = "bar" }
    labels   = [ebs_block_device.key]
    content {
      value = ebs_block_device.value
    }
  }
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body: &hclext.BodySchema{
							Blocks: []hclext.BlockSchema{
								{
									Type:       "ebs_block_device",
									LabelNames: []string{"name"},
									Body:       &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}},
								},
							},
						},
					},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks: hclext.Blocks{
								{
									Type:   "ebs_block_device",
									Labels: []string{"foo"},
									Body:   &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("bar"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
								},
							},
						},
					},
				},
			},
		},
		{
			// Terraform does not allow variables to be used in dynamic labels.
			// @see https://github.com/hashicorp/terraform/issues/32180
			name: "dynamic labels with variable",
			config: `
variable "label" {
  default = "baz"
}
resource "aws_instance" "main" {
  dynamic "ebs_block_device" {
    for_each = { foo = "bar" }
    labels   = [var.label]
    content {
      value = ebs_block_device.value
    }
  }
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body: &hclext.BodySchema{
							Blocks: []hclext.BlockSchema{
								{
									Type:       "ebs_block_device",
									LabelNames: []string{"name"},
									Body:       &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}},
								},
							},
						},
					},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks: hclext.Blocks{
								{
									Type:   "ebs_block_device",
									Labels: []string{"baz"},
									Body:   &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("bar"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "meta-aruguments and dynamic blocks",
			config: `
resource "aws_instance" "main" {
  count = 2

  dynamic "ebs_block_device" {
    for_each = toset(["foo", "bar"])
    content {
      value = "${count.index}-${ebs_block_device.value}"
    }
  }
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body: &hclext.BodySchema{
							Blocks: []hclext.BlockSchema{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "value"}}},
								},
							},
						},
					},
				},
			},
			want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks: hclext.Blocks{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("0-bar"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
								},
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("0-foo"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
								},
							},
						},
					},
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "main"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks: hclext.Blocks{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("1-bar"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
								},
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{Attributes: hclext.Attributes{"value": {Name: "value", Expr: hcl.StaticExpr(cty.StringVal("1-foo"), hcl.Range{})}}, Blocks: hclext.Blocks{}},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			if err := fs.WriteFile("main.tf", []byte(test.config), os.ModePerm); err != nil {
				t.Fatal(err)
			}
			file, diags := hclsyntax.ParseConfig([]byte(test.config), "main.tf", hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			parser := NewParser(fs)
			mod, diags := parser.LoadConfigDir(".", ".")
			if diags.HasErrors() {
				t.Fatal(diags)
			}
			config, diags := BuildConfig(mod, ModuleWalkerFunc(func(req *ModuleRequest) (*Module, *version.Version, hcl.Diagnostics) { return nil, nil, nil }))
			if diags.HasErrors() {
				t.Fatal(diags)
			}
			variableValues, diags := VariableValues(config)
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

			expanded, diags := evaluator.ExpandBlock(file.Body, test.schema)
			if diags.HasErrors() {
				t.Fatal(diags)
			}
			got, diags := hclext.PartialContent(expanded, test.schema)
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			opts := cmp.Options{
				cmpopts.IgnoreFields(hclext.Block{}, "TypeRange", "LabelRanges"),
				cmpopts.IgnoreFields(hclext.Attribute{}, "NameRange"),
				cmpopts.IgnoreFields(hcl.Range{}, "Start", "End", "Filename"),
				cmp.Comparer(func(x, y hcl.Expression) bool {
					xv, diags := evaluator.EvaluateExpr(x, cty.DynamicPseudoType)
					if diags.HasErrors() {
						t.Fatal(diags)
					}
					yv, diags := evaluator.EvaluateExpr(y, cty.DynamicPseudoType)
					if diags.HasErrors() {
						t.Fatal(diags)
					}
					return xv.RawEquals(yv)
				}),
			}
			if diff := cmp.Diff(got, test.want, opts); diff != "" {
				t.Error(diff)
			}
		})
	}
}
