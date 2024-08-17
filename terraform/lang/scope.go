package lang

import (
	"strings"
	"sync"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty/function"

	"github.com/terraform-linters/tflint/terraform/addrs"
)

// Scope is the main type in this package, allowing dynamic evaluation of
// blocks and expressions based on some contextual information that informs
// which variables and functions will be available.
type Scope struct {
	// Data is used to resolve references in expressions.
	Data Data

	// SelfAddr is the address that the "self" object should be an alias of,
	// or nil if the "self" object should not be available at all.
	SelfAddr addrs.Referenceable

	// BaseDir is the base directory used by any interpolation functions that
	// accept filesystem paths as arguments.
	BaseDir string

	// PureOnly can be set to true to request that any non-pure functions
	// produce unknown value results rather than actually executing. This is
	// important during a plan phase to avoid generating results that could
	// then differ during apply.
	PureOnly bool

	// CallStack is a stack for recording local value references to detect
	// circular references.
	CallStack *CallStack

	funcs     map[string]function.Function
	funcsLock sync.Mutex
}

type CallStack struct {
	addrs map[string]addrs.Reference
	stack []string
}

func NewCallStack() *CallStack {
	return &CallStack{
		addrs: make(map[string]addrs.Reference),
		stack: make([]string, 0),
	}
}

func (g *CallStack) Push(addr addrs.Reference) hcl.Diagnostics {
	g.stack = append(g.stack, addr.Subject.String())

	if _, exists := g.addrs[addr.Subject.String()]; exists {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "circular reference found",
				Detail:   g.String(),
				Subject:  addr.SourceRange.Ptr(),
			},
		}
	}
	g.addrs[addr.Subject.String()] = addr
	return hcl.Diagnostics{}
}

func (g *CallStack) Pop() {
	if g.Empty() {
		panic("cannot pop from empty stack")
	}

	addr := g.stack[len(g.stack)-1]
	g.stack = g.stack[:len(g.stack)-1]
	delete(g.addrs, addr)
}

func (g *CallStack) String() string {
	return strings.Join(g.stack, " -> ")
}

func (g *CallStack) Empty() bool {
	return len(g.stack) == 0
}

func (g *CallStack) Clear() {
	g.addrs = make(map[string]addrs.Reference)
	g.stack = make([]string, 0)
}
