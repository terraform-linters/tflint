// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package tryfunc contains some optional functions that can be exposed in
// HCL-based languages to allow authors to test whether a particular expression
// can succeed and take dynamic action based on that result.
//
// These functions are implemented in terms of the customdecode extension from
// the sibling directory "customdecode", and so they are only useful when
// used within an HCL EvalContext. Other systems using cty functions are
// unlikely to support the HCL-specific "customdecode" extension.
package tryfunc

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/customdecode"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// TryFunc is a variadic function that tries to evaluate all of is arguments
// in sequence until one succeeds, in which case it returns that result, or
// returns an error if none of them succeed.
var TryFunc function.Function

// CanFunc tries to evaluate the expression given in its first argument.
var CanFunc function.Function

func init() {
	TryFunc = function.New(&function.Spec{
		VarParam: &function.Parameter{
			Name: "expressions",
			Type: customdecode.ExpressionClosureType,
		},
		Type: func(args []cty.Value) (cty.Type, error) {
			v, err := try(args)
			if err != nil {
				return cty.NilType, err
			}
			return v.Type(), nil
		},
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			return try(args)
		},
	})
	CanFunc = function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "expression",
				Type: customdecode.ExpressionClosureType,
			},
		},
		Type: function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			return can(args[0])
		},
	})
}

func try(args []cty.Value) (cty.Value, error) {
	if len(args) == 0 {
		return cty.NilVal, errors.New("at least one argument is required")
	}

	// We'll collect up all of the diagnostics we encounter along the way
	// and report them all if none of the expressions succeed, so that the
	// user might get some hints on how to make at least one succeed.
	var diags hcl.Diagnostics
	for _, arg := range args {
		closure := customdecode.ExpressionClosureFromVal(arg)

		v, moreDiags := closure.Value()
		diags = append(diags, moreDiags...)

		if moreDiags.HasErrors() {
			// If there's an error we know it will always fail and can
			// continue. A more refined value will not remove an error from
			// the expression.
			continue
		}

		if !v.IsWhollyKnown() {
			// If there are any unknowns in the value at all, we cannot be
			// certain that the final value will be consistent or have the same
			// type, so wee need to be conservative and return a dynamic value.

			// There are two different classes of failure that can happen when
			// an expression transitions from unknown to known; an operation on
			// a dynamic value becomes invalid for the type once the type is
			// known, or an index expression on a collection fails once the
			// collection value is known. These changes from a
			// valid-partially-unknown expression to an invalid-known
			// expression can produce inconsistent results by changing which
			// "try" argument is returned, which may be a collection with
			// different previously known values, or a different type entirely
			// ("try" does not require consistent argument types)
			return cty.DynamicVal, nil
		}

		return v, nil // ignore any accumulated diagnostics if one succeeds
	}

	// If we fall out here then none of the expressions succeeded, and so
	// we must have at least one diagnostic and we'll return all of them
	// so that the user can see the errors related to whichever one they
	// were expecting to have succeeded in this case.
	//
	// Because our function must return a single error value rather than
	// diagnostics, we'll construct a suitable error message string
	// that will make sense in the context of the function call failure
	// diagnostic HCL will eventually wrap this in.
	var buf strings.Builder
	buf.WriteString("no expression succeeded:\n")
	for _, diag := range diags {
		if diag.Subject != nil {
			buf.WriteString(fmt.Sprintf("- %s (at %s)\n  %s\n", diag.Summary, diag.Subject, diag.Detail))
		} else {
			buf.WriteString(fmt.Sprintf("- %s\n  %s\n", diag.Summary, diag.Detail))
		}
	}
	buf.WriteString("\nAt least one expression must produce a successful result")
	return cty.NilVal, errors.New(buf.String())
}

func can(arg cty.Value) (cty.Value, error) {
	closure := customdecode.ExpressionClosureFromVal(arg)
	v, diags := closure.Value()
	if diags.HasErrors() {
		return cty.False, nil
	}

	if !v.IsWhollyKnown() {
		// If the value is not wholly known, we still cannot be certain that
		// the expression was valid. There may be yet index expressions which
		// will fail once values are completely known.
		return cty.UnknownVal(cty.Bool), nil
	}

	return cty.True, nil
}
