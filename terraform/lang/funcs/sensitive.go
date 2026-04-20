// SPDX-License-Identifier: MPL-2.0

package funcs

import (
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang/marks"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

var SensitiveFunc = function.New(&function.Spec{
	Params: []function.Parameter{dynamicMarkedValueParam("value")},
	Type: func(args []cty.Value) (cty.Type, error) {
		return args[0].Type(), nil
	},
	Impl: func(args []cty.Value, _ cty.Type) (cty.Value, error) {
		return args[0].Mark(marks.Sensitive), nil
	},
})

var NonsensitiveFunc = function.New(&function.Spec{
	Params: []function.Parameter{dynamicMarkedValueParam("value")},
	Type: func(args []cty.Value) (cty.Type, error) {
		return args[0].Type(), nil
	},
	Impl: func(args []cty.Value, _ cty.Type) (cty.Value, error) {
		value, existingMarks := args[0].Unmark()
		delete(existingMarks, marks.Sensitive)
		return value.WithMarks(existingMarks), nil
	},
})

var IssensitiveFunc = function.New(&function.Spec{
	Params: []function.Parameter{dynamicMarkedValueParam("value")},
	Type: func([]cty.Value) (cty.Type, error) {
		return cty.Bool, nil
	},
	Impl: func(args []cty.Value, _ cty.Type) (cty.Value, error) {
		value := args[0]
		switch {
		case value.HasMark(marks.Sensitive):
			return cty.True, nil
		case !value.IsKnown():
			return cty.UnknownVal(cty.Bool), nil
		default:
			return cty.False, nil
		}
	},
})

func Sensitive(value cty.Value) (cty.Value, error) {
	return SensitiveFunc.Call([]cty.Value{value})
}

func Nonsensitive(value cty.Value) (cty.Value, error) {
	return NonsensitiveFunc.Call([]cty.Value{value})
}

func Issensitive(value cty.Value) (cty.Value, error) {
	return IssensitiveFunc.Call([]cty.Value{value})
}

func dynamicMarkedValueParam(name string) function.Parameter {
	return function.Parameter{
		Name:             name,
		Type:             cty.DynamicPseudoType,
		AllowUnknown:     true,
		AllowNull:        true,
		AllowMarked:      true,
		AllowDynamicType: true,
	}
}
