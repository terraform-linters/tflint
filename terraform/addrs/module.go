package addrs

import (
	"strings"
)

// Module is an address for a module call within configuration. This is
// the static counterpart of ModuleInstance, representing a traversal through
// the static module call tree in configuration and does not take into account
// the potentially-multiple instances of a module that might be created by
// "count" and "for_each" arguments within those calls.
//
// This type should be used only in very specialized cases when working with
// the static module call tree. Type ModuleInstance is appropriate in more cases.
//
// Although Module is a slice, it should be treated as immutable after creation.
type Module []string

// RootModule is the module address representing the root of the static module
// call tree, which is also the zero value of Module.
//
// Note that this is not the root of the dynamic module tree, which is instead
// represented by RootModuleInstance.
var RootModule Module

// IsRoot returns true if the receiver is the address of the root module,
// or false otherwise.
func (m Module) IsRoot() bool {
	return len(m) == 0
}

func (m Module) String() string {
	if len(m) == 0 {
		return ""
	}
	// Calculate necessary space.
	l := 0
	for _, step := range m {
		l += len(step)
	}
	buf := strings.Builder{}
	// 8 is len(".module.") which separates entries.
	buf.Grow(l + len(m)*8)
	sep := ""
	for _, step := range m {
		buf.WriteString(sep)
		buf.WriteString("module.")
		buf.WriteString(step)
		sep = "."
	}
	return buf.String()
}
