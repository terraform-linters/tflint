package tfhcl

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type exprWrap struct {
	hcl.Expression
	di *dynamicIteration
	mi *metaArgIteration
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
	extCtx := e.di.EvalContext(e.mi.EvalContext(ctx))
	return e.Expression.Value(extCtx)
}

// UnwrapExpression returns the expression being wrapped by this instance.
// This allows the original expression to be recovered by hcl.UnwrapExpression.
func (e exprWrap) UnwrapExpression() hcl.Expression {
	return e.Expression
}
