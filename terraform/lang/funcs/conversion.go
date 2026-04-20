// SPDX-License-Identifier: MPL-2.0

package funcs

import (
	"strconv"

	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang/marks"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
)

func MakeToFunc(wantTy cty.Type) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{dynamicConversionParam("v")},
		Type: func(args []cty.Value) (cty.Type, error) {
			return conversionResultType(args[0].Type(), wantTy)
		},
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			return convertValue(args[0], retType, wantTy)
		},
	})
}

var EphemeralAsNullFunc = function.New(&function.Spec{
	Params: []function.Parameter{dynamicConversionParam("value")},
	Type: func(args []cty.Value) (cty.Type, error) {
		return args[0].Type(), nil
	},
	Impl: func(args []cty.Value, _ cty.Type) (cty.Value, error) {
		return cty.Transform(args[0], stripEphemeralValues)
	},
})

func EphemeralAsNull(input cty.Value) (cty.Value, error) {
	return EphemeralAsNullFunc.Call([]cty.Value{input})
}

func dynamicConversionParam(name string) function.Parameter {
	return function.Parameter{
		Name:             name,
		Type:             cty.DynamicPseudoType,
		AllowNull:        true,
		AllowMarked:      true,
		AllowDynamicType: true,
		AllowUnknown:     true,
	}
}

func conversionResultType(gotTy, wantTy cty.Type) (cty.Type, error) {
	if gotTy.Equals(wantTy) {
		return wantTy, nil
	}
	if convert.GetConversionUnsafe(gotTy, wantTy) != nil {
		return wantTy, nil
	}

	switch {
	case gotTy.IsTupleType() && wantTy.IsTupleType():
		return cty.NilType, function.NewArgErrorf(0, "incompatible tuple type for conversion: %s", convert.MismatchMessage(gotTy, wantTy))
	case gotTy.IsObjectType() && wantTy.IsObjectType():
		return cty.NilType, function.NewArgErrorf(0, "incompatible object type for conversion: %s", convert.MismatchMessage(gotTy, wantTy))
	default:
		return cty.NilType, function.NewArgErrorf(0, "cannot convert %s to %s", gotTy.FriendlyName(), wantTy.FriendlyNameForConstraint())
	}
}

func convertValue(input cty.Value, retType cty.Type, wantTy cty.Type) (cty.Value, error) {
	if !input.IsKnown() {
		return cty.UnknownVal(retType).WithSameMarks(input), nil
	}

	converted, err := convert.Convert(input, retType)
	if err == nil {
		return converted, nil
	}

	unmarked, _ := input.UnmarkDeep()
	gotTy := unmarked.Type()

	switch {
	case marks.Contains(input, marks.Sensitive):
		return cty.NilVal, function.NewArgErrorf(0, "cannot convert this sensitive %s to %s", gotTy.FriendlyName(), wantTy.FriendlyNameForConstraint())
	case gotTy == cty.String && wantTy == cty.Bool:
		return cty.NilVal, function.NewArgErrorf(0, `cannot convert %s to bool; only the strings "true" or "false" are allowed`, quotedStringValue(unmarked))
	case gotTy == cty.String && wantTy == cty.Number:
		return cty.NilVal, function.NewArgErrorf(0, `cannot convert %s to number; given string must be a decimal representation of a number`, quotedStringValue(unmarked))
	default:
		return cty.NilVal, function.NewArgErrorf(0, "cannot convert %s to %s", gotTy.FriendlyName(), wantTy.FriendlyNameForConstraint())
	}
}

func quotedStringValue(value cty.Value) string {
	if value.IsNull() {
		return "string"
	}
	return strconv.Quote(value.AsString())
}

func stripEphemeralValues(_ cty.Path, value cty.Value) (cty.Value, error) {
	_, valueMarks := value.Unmark()
	if _, ok := valueMarks[marks.Ephemeral]; !ok {
		return value, nil
	}

	delete(valueMarks, marks.Ephemeral)
	if !value.IsKnown() {
		return cty.UnknownVal(value.Type()).WithMarks(valueMarks), nil
	}
	return cty.NullVal(value.Type()).WithMarks(valueMarks), nil
}
