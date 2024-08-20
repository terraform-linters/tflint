package tflint

import (
	"github.com/hashicorp/hcl/v2"
)

// ExprWalker is an interface used with WalkExpressions.
type ExprWalker interface {
	Enter(expr hcl.Expression) hcl.Diagnostics
	Exit(expr hcl.Expression) hcl.Diagnostics
}

// ExprWalkFunc is the callback signature for WalkExpressions.
// This satisfies the ExprWalker interface.
type ExprWalkFunc func(expr hcl.Expression) hcl.Diagnostics

// Enter is a function of ExprWalker that invokes itself on the passed expression.
func (f ExprWalkFunc) Enter(expr hcl.Expression) hcl.Diagnostics {
	return f(expr)
}

// Exit is one of ExprWalker's functions, noop here
func (f ExprWalkFunc) Exit(expr hcl.Expression) hcl.Diagnostics {
	return nil
}
