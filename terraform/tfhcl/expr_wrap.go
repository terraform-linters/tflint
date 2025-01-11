// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfhcl

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type exprWrap struct {
	hcl.Expression
	di *dynamicIteration
	mi *metaArgIteration

	// resultMarks is a set of marks that must be applied to whatever
	// value results from this expression. We do this whenever a
	// dynamic block's for_each expression produced a marked result,
	// since in that case any nested expressions inside are treated
	// as being derived from that for_each expression.
	resultMarks cty.ValueMarks
}

func (e exprWrap) Variables() []hcl.Traversal {
	raw := e.Expression.Variables()
	if e.di == nil {
		return raw
	}
	ret := make([]hcl.Traversal, 0, len(raw))

	// Filter out traversals that refer to our dynamic iterator name or any
	// iterator we've inherited; we're going to provide those in
	// our Value wrapper, so the caller doesn't need to know about them.
	for _, traversal := range raw {
		rootName := traversal.RootName()
		if rootName == e.di.IteratorName {
			continue
		}
		if _, inherited := e.di.Inherited[rootName]; inherited {
			continue
		}
		ret = append(ret, traversal)
	}
	return ret
}

func (e exprWrap) Value(ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	if e.di == nil && e.mi == nil {
		// If we don't have an active iteration then we can just use the
		// given EvalContext directly.
		return e.prepareValue(e.Expression.Value(ctx))
	}

	extCtx := e.di.EvalContext(e.mi.EvalContext(ctx))
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
