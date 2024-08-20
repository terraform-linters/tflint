package hclext

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// BoundExpr represents an expression whose a value is bound.
// This is a wrapper for any expression, typically satisfying
// an interface to behave like the wrapped expression.
//
// The difference is that when resolving a value with `Value()`,
// instead of resolving the variables with EvalContext,
// the bound value is returned directly.
type BoundExpr struct {
	Val cty.Value

	original hcl.Expression
}

var _ hcl.Expression = (*BoundExpr)(nil)

// BindValue binds the passed value to an expression.
// This returns the bound expression.
func BindValue(val cty.Value, expr hcl.Expression) hcl.Expression {
	return &BoundExpr{original: expr, Val: val}
}

// Value returns the bound value.
func (e BoundExpr) Value(*hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	return e.Val, nil
}

// Variables delegates to the wrapped expression.
func (e BoundExpr) Variables() []hcl.Traversal {
	return e.original.Variables()
}

// Range delegates to the wrapped expression.
func (e BoundExpr) Range() hcl.Range {
	return e.original.Range()
}

// StartRange delegates to the wrapped expression.
func (e BoundExpr) StartRange() hcl.Range {
	return e.original.StartRange()
}

// UnwrapExpression returns the original expression.
// This satisfies the hcl.unwrapExpression interface.
func (e BoundExpr) UnwrapExpression() hcl.Expression {
	return e.original
}
