// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package funcs

import (
	"strconv"

	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang/marks"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
)

// MakeToFunc constructs a "to..." function, like "tostring", which converts
// its argument to a specific type or type kind.
//
// The given type wantTy can be any type constraint that cty's "convert" package
// would accept. In particular, this means that you can pass
// cty.List(cty.DynamicPseudoType) to mean "list of any single type", which
// will then cause cty to attempt to unify all of the element types when given
// a tuple.
func MakeToFunc(wantTy cty.Type) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "v",
				// We use DynamicPseudoType rather than wantTy here so that
				// all values will pass through the function API verbatim and
				// we can handle the conversion logic within the Type and
				// Impl functions. This allows us to customize the error
				// messages to be more appropriate for an explicit type
				// conversion, whereas the cty function system produces
				// messages aimed at _implicit_ type conversions.
				Type:             cty.DynamicPseudoType,
				AllowNull:        true,
				AllowMarked:      true,
				AllowDynamicType: true,
				AllowUnknown:     true,
			},
		},
		Type: func(args []cty.Value) (cty.Type, error) {
			gotTy := args[0].Type()
			if gotTy.Equals(wantTy) {
				return wantTy, nil
			}
			conv := convert.GetConversionUnsafe(args[0].Type(), wantTy)
			if conv == nil {
				// We'll use some specialized errors for some trickier cases,
				// but most we can handle in a simple way.
				switch {
				case gotTy.IsTupleType() && wantTy.IsTupleType():
					return cty.NilType, function.NewArgErrorf(0, "incompatible tuple type for conversion: %s", convert.MismatchMessage(gotTy, wantTy))
				case gotTy.IsObjectType() && wantTy.IsObjectType():
					return cty.NilType, function.NewArgErrorf(0, "incompatible object type for conversion: %s", convert.MismatchMessage(gotTy, wantTy))
				default:
					return cty.NilType, function.NewArgErrorf(0, "cannot convert %s to %s", gotTy.FriendlyName(), wantTy.FriendlyNameForConstraint())
				}
			}
			// If a conversion is available then everything is fine.
			return wantTy, nil
		},
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			if !args[0].IsKnown() {
				return cty.UnknownVal(retType).WithSameMarks(args[0]), nil
			}

			ret, err := convert.Convert(args[0], retType)
			if err != nil {
				val, _ := args[0].UnmarkDeep()
				// Because we used GetConversionUnsafe above, conversion can
				// still potentially fail in here. For example, if the user
				// asks to convert the string "a" to bool then we'll
				// optimistically permit it during type checking but fail here
				// once we note that the value isn't either "true" or "false".
				gotTy := val.Type()
				switch {
				case marks.Contains(args[0], marks.Sensitive):
					// Generic message so we won't inadvertently disclose
					// information about sensitive values.
					return cty.NilVal, function.NewArgErrorf(0, "cannot convert this sensitive %s to %s", gotTy.FriendlyName(), wantTy.FriendlyNameForConstraint())

				case gotTy == cty.String && wantTy == cty.Bool:
					what := "string"
					if !val.IsNull() {
						what = strconv.Quote(val.AsString())
					}
					return cty.NilVal, function.NewArgErrorf(0, `cannot convert %s to bool; only the strings "true" or "false" are allowed`, what)
				case gotTy == cty.String && wantTy == cty.Number:
					what := "string"
					if !val.IsNull() {
						what = strconv.Quote(val.AsString())
					}
					return cty.NilVal, function.NewArgErrorf(0, `cannot convert %s to number; given string must be a decimal representation of a number`, what)
				default:
					return cty.NilVal, function.NewArgErrorf(0, "cannot convert %s to %s", gotTy.FriendlyName(), wantTy.FriendlyNameForConstraint())
				}
			}
			return ret, nil
		},
	})
}

// EphemeralAsNullFunc is a cty function that takes a value of any type and
// returns a similar value with any ephemeral-marked values anywhere in the
// structure replaced with a null value of the same type that is not marked
// as ephemeral.
//
// This is intended as a convenience for returning the non-ephemeral parts of
// a partially-ephemeral data structure through an output value that isn't
// ephemeral itself.
var EphemeralAsNullFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "value",
			Type:             cty.DynamicPseudoType,
			AllowDynamicType: true,
			AllowUnknown:     true,
			AllowNull:        true,
			AllowMarked:      true,
		},
	},
	Type: func(args []cty.Value) (cty.Type, error) {
		// This function always preserves the type of the given argument.
		return args[0].Type(), nil
	},
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		return cty.Transform(args[0], func(p cty.Path, v cty.Value) (cty.Value, error) {
			_, givenMarks := v.Unmark()
			if _, isEphemeral := givenMarks[marks.Ephemeral]; isEphemeral {
				// We'll strip the ephemeral mark but retain any other marks
				// that might be present on the input.
				delete(givenMarks, marks.Ephemeral)
				if !v.IsKnown() {
					// If the source value is unknown then we must leave it
					// unknown because its final type might be more precise
					// than the associated type constraint and returning a
					// typed null could therefore over-promise on what the
					// final result type will be.
					// We're deliberately constructing a fresh unknown value
					// here, rather than returning the one we were given,
					// because we need to discard any refinements that the
					// unknown value might be carrying that definitely won't
					// be honored when we force the final result to be null.
					return cty.UnknownVal(v.Type()).WithMarks(givenMarks), nil
				}
				return cty.NullVal(v.Type()).WithMarks(givenMarks), nil
			}
			return v, nil
		})
	},
})

func EphemeralAsNull(input cty.Value) (cty.Value, error) {
	return EphemeralAsNullFunc.Call([]cty.Value{input})
}
