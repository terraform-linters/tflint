// SPDX-License-Identifier: MPL-2.0

package funcs

import (
	"math"
	"math/big"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/gocty"
)

var LogFunc = newFloatBinaryFunc("num", "base", func(num, base float64) float64 {
	return math.Log(num) / math.Log(base)
})

var PowFunc = newFloatBinaryFunc("num", "power", math.Pow)

var SignumFunc = function.New(&function.Spec{
	Params: []function.Parameter{{
		Name: "num",
		Type: cty.Number,
	}},
	Type:         function.StaticReturnType(cty.Number),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, _ cty.Type) (cty.Value, error) {
		num, err := decodeInt(args[0])
		if err != nil {
			return cty.UnknownVal(cty.String), err
		}
		switch {
		case num < 0:
			return cty.NumberIntVal(-1), nil
		case num > 0:
			return cty.NumberIntVal(1), nil
		default:
			return cty.NumberIntVal(0), nil
		}
	},
})

var ParseIntFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:         "number",
			Type:         cty.DynamicPseudoType,
			AllowMarked:  true,
			AllowUnknown: true,
		},
		{
			Name:         "base",
			Type:         cty.Number,
			AllowMarked:  true,
			AllowUnknown: true,
		},
	},
	Type: func(args []cty.Value) (cty.Type, error) {
		if !args[0].Type().Equals(cty.String) {
			return cty.Number, function.NewArgErrorf(0, "first argument must be a string, not %s", args[0].Type().FriendlyName())
		}
		return cty.Number, nil
	},
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		numberArg, numberMarks := args[0].Unmark()
		baseArg, baseMarks := args[1].Unmark()
		if !numberArg.IsKnown() || !baseArg.IsKnown() {
			return cty.UnknownVal(retType).WithMarks(numberMarks, baseMarks), nil
		}

		numberString, err := decodeString(numberArg)
		if err != nil {
			return cty.UnknownVal(cty.String), function.NewArgError(0, err)
		}
		base, err := decodeInt(baseArg)
		if err != nil {
			return cty.UnknownVal(cty.Number), function.NewArgError(1, err)
		}
		if base < 2 || base > 62 {
			return cty.UnknownVal(cty.Number), function.NewArgErrorf(1, "base must be a whole number between 2 and 62 inclusive")
		}

		parsed, ok := (&big.Int{}).SetString(numberString, base)
		if !ok {
			return cty.UnknownVal(cty.Number), function.NewArgErrorf(
				0,
				"cannot parse %s as a base %s integer",
				redactIfSensitive(numberString, numberMarks),
				redactIfSensitive(base, baseMarks),
			)
		}

		return cty.NumberVal((&big.Float{}).SetInt(parsed)).WithMarks(numberMarks, baseMarks), nil
	},
})

func Log(num, base cty.Value) (cty.Value, error) {
	return LogFunc.Call([]cty.Value{num, base})
}

func Pow(num, power cty.Value) (cty.Value, error) {
	return PowFunc.Call([]cty.Value{num, power})
}

func Signum(num cty.Value) (cty.Value, error) {
	return SignumFunc.Call([]cty.Value{num})
}

func ParseInt(num cty.Value, base cty.Value) (cty.Value, error) {
	return ParseIntFunc.Call([]cty.Value{num, base})
}

func newFloatBinaryFunc(leftName, rightName string, op func(float64, float64) float64) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{Name: leftName, Type: cty.Number},
			{Name: rightName, Type: cty.Number},
		},
		Type:         function.StaticReturnType(cty.Number),
		RefineResult: refineNotNull,
		Impl: func(args []cty.Value, _ cty.Type) (cty.Value, error) {
			left, err := decodeFloat(args[0])
			if err != nil {
				return cty.UnknownVal(cty.String), err
			}
			right, err := decodeFloat(args[1])
			if err != nil {
				return cty.UnknownVal(cty.String), err
			}
			return cty.NumberFloatVal(op(left, right)), nil
		},
	})
}

func decodeFloat(value cty.Value) (float64, error) {
	var out float64
	return out, gocty.FromCtyValue(value, &out)
}

func decodeInt(value cty.Value) (int, error) {
	var out int
	return out, gocty.FromCtyValue(value, &out)
}

func decodeString(value cty.Value) (string, error) {
	var out string
	return out, gocty.FromCtyValue(value, &out)
}
