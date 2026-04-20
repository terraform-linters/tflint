// SPDX-License-Identifier: MPL-2.0

package terraform

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

var EncodeTfvarsFunc = function.New(&function.Spec{
	Params: []function.Parameter{{
		Name:             "value",
		Type:             cty.DynamicPseudoType,
		AllowNull:        true,
		AllowDynamicType: true,
		AllowUnknown:     true,
	}},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, _ cty.Type) (cty.Value, error) {
		value, err := requireSingleDynamicArg(args, "value")
		if err != nil {
			return cty.NilVal, err
		}
		if value.IsNull() {
			return cty.NilVal, function.NewArgErrorf(1, "cannot encode a null value in tfvars syntax")
		}
		if !value.IsWhollyKnown() {
			return cty.UnknownVal(cty.String).RefineNotNull(), nil
		}

		names, err := tfvarsAttributeNames(value)
		if err != nil {
			return cty.NilVal, err
		}

		file := hclwrite.NewEmptyFile()
		body := file.Body()
		for _, name := range names {
			attrValue, _ := hcl.Index(value, cty.StringVal(name), nil)
			body.SetAttributeValue(name, attrValue)
		}
		return cty.StringVal(string(file.Bytes())), nil
	},
})

var DecodeTfvarsFunc = function.New(&function.Spec{
	Params: []function.Parameter{{
		Name:      "src",
		Type:      cty.String,
		AllowNull: true,
	}},
	Type: function.StaticReturnType(cty.DynamicPseudoType),
	Impl: func(args []cty.Value, _ cty.Type) (cty.Value, error) {
		if err := requireExactlyOneArg(args); err != nil {
			return cty.NilVal, err
		}
		src := args[0]
		if src.Type() != cty.String {
			return cty.NilVal, fmt.Errorf("argument must be a string")
		}
		if src.IsNull() {
			return cty.NilVal, fmt.Errorf("cannot decode tfvars from a null value")
		}
		if !src.IsKnown() {
			return cty.DynamicVal, nil
		}

		parsed, diags := hclsyntax.ParseConfig([]byte(src.AsString()), "<decode_tfvars argument>", hcl.InitialPos)
		if diags.HasErrors() {
			return cty.NilVal, fmt.Errorf("invalid tfvars syntax: %s", diags.Error())
		}
		attrs, diags := parsed.Body.JustAttributes()
		if diags.HasErrors() {
			return cty.NilVal, fmt.Errorf("invalid tfvars content: %s", diags.Error())
		}

		values := make(map[string]cty.Value, len(attrs))
		for name, attr := range attrs {
			value, valueDiags := attr.Expr.Value(nil)
			if valueDiags.HasErrors() {
				return cty.NilVal, fmt.Errorf("invalid expression for variable %q: %s", name, valueDiags.Error())
			}
			values[name] = value
		}
		return cty.ObjectVal(values), nil
	},
})

var EncodeExprFunc = function.New(&function.Spec{
	Params: []function.Parameter{{
		Name:             "value",
		Type:             cty.DynamicPseudoType,
		AllowNull:        true,
		AllowDynamicType: true,
		AllowUnknown:     true,
	}},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, _ cty.Type) (cty.Value, error) {
		value, err := requireSingleDynamicArg(args, "value")
		if err != nil {
			return cty.NilVal, err
		}
		if !value.IsWhollyKnown() {
			return refinedUnknownExpression(value), nil
		}

		encoded := bytes.TrimSpace(hclwrite.TokensForValue(value).Bytes())
		return cty.StringVal(string(encoded)), nil
	},
})

func requireExactlyOneArg(args []cty.Value) error {
	if len(args) > 1 {
		return function.NewArgErrorf(1, "too many arguments; only one expected")
	}
	if len(args) == 0 {
		return fmt.Errorf("exactly one argument is required")
	}
	return nil
}

func requireSingleDynamicArg(args []cty.Value, _ string) (cty.Value, error) {
	if err := requireExactlyOneArg(args); err != nil {
		return cty.NilVal, err
	}
	return args[0], nil
}

func tfvarsAttributeNames(value cty.Value) ([]string, error) {
	var names []string

	switch {
	case value.Type().IsObjectType():
		for name := range value.Type().AttributeTypes() {
			names = append(names, name)
		}
	case value.Type().IsMapType():
		names = make([]string, 0, value.LengthInt())
		for it := value.ElementIterator(); it.Next(); {
			key, _ := it.Element()
			names = append(names, key.AsString())
		}
	default:
		return nil, function.NewArgErrorf(1, "invalid value to encode: must be an object whose attribute names will become the encoded variable names")
	}

	sort.Strings(names)
	for _, name := range names {
		if !hclsyntax.ValidIdentifier(name) {
			return nil, function.NewArgErrorf(1, "invalid variable name %q: must be a valid identifier, per Terraform's rules for input variable declarations", name)
		}
	}
	return names, nil
}

func refinedUnknownExpression(value cty.Value) cty.Value {
	refined := cty.UnknownVal(cty.String).RefineNotNull()
	if value.Range().CouldBeNull() {
		return refined
	}

	switch ty := value.Type(); {
	case ty.IsObjectType() || ty.IsMapType():
		return refined.Refine().StringPrefixFull("{").NewValue()
	case ty.IsTupleType() || ty.IsListType() || ty.IsSetType():
		return refined.Refine().StringPrefixFull("[").NewValue()
	case ty == cty.String:
		return refined.Refine().StringPrefixFull(`"`).NewValue()
	default:
		return refined
	}
}
