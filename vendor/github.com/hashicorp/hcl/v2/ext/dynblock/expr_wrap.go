// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynblock

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type exprWrap struct {
	hcl.Expression
	i *iteration

	// resultMarks is a set of marks that must be applied to whatever
	// value results from this expression. We do this whenever a
	// dynamic block's for_each expression produced a marked result,
	// since in that case any nested expressions inside are treated
	// as being derived from that for_each expression.
	//
	// (calling applications might choose to reject marks by passing
	// an [OptCheckForEach] to [Expand] and returning an error when
	// marks are present, but this mechanism is here to help achieve
	// reasonable behavior for situations where marks are permitted,
	// which is the default.)
	resultMarks cty.ValueMarks
}

func (e exprWrap) Variables() []hcl.Traversal {
	raw := e.Expression.Variables()
	ret := make([]hcl.Traversal, 0, len(raw))

	// Filter out traversals that refer to our iterator name or any
	// iterator we've inherited; we're going to provide those in
	// our Value wrapper, so the caller doesn't need to know about them.
	for _, traversal := range raw {
		rootName := traversal.RootName()
		if rootName == e.i.IteratorName {
			continue
		}
		if _, inherited := e.i.Inherited[rootName]; inherited {
			continue
		}
		ret = append(ret, traversal)
	}
	return ret
}

func (e exprWrap) Value(ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	if e.i == nil {
		// If we don't have an active iteration then we can just use the
		// given EvalContext directly.
		return e.prepareValue(e.Expression.Value(ctx))
	}
	extCtx := e.i.EvalContext(ctx)
	return e.prepareValue(e.Expression.Value(extCtx))
}

// UnwrapExpression returns the expression being wrapped by this instance.
// This allows the original expression to be recovered by hcl.UnwrapExpression.
func (e exprWrap) UnwrapExpression() hcl.Expression {
	return e.Expression
}

func (e exprWrap) prepareValue(val cty.Value, diags hcl.Diagnostics) (cty.Value, hcl.Diagnostics) {
	return val.WithMarks(e.resultMarks), diags
}
