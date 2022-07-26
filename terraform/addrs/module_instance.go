package addrs

import (
	"strings"
)

// ModuleInstance is an address for a particular module instance within the
// dynamic module tree. This is an extension of the static traversals
// represented by type Module that deals with the possibility of a single
// module call producing multiple instances via the "count" and "for_each"
// arguments.
//
// Although ModuleInstance is a slice, it should be treated as immutable after
// creation.
type ModuleInstance []ModuleInstanceStep

// UnkeyedInstanceShim is a shim method for converting a Module address to the
// equivalent ModuleInstance address that assumes that no modules have
// keyed instances.
//
// This is a temporary allowance for the fact that Terraform does not presently
// support "count" and "for_each" on modules, and thus graph building code that
// derives graph nodes from configuration must just assume unkeyed modules
// in order to construct the graph. At a later time when "count" and "for_each"
// support is added for modules, all callers of this method will need to be
// reworked to allow for keyed module instances.
func (m Module) UnkeyedInstanceShim() ModuleInstance {
	path := make(ModuleInstance, len(m))
	for i, name := range m {
		path[i] = ModuleInstanceStep{Name: name}
	}
	return path
}

// ModuleInstanceStep is a single traversal step through the dynamic module
// tree. It is used only as part of ModuleInstance.
type ModuleInstanceStep struct {
	Name        string
	InstanceKey InstanceKey
}

// RootModuleInstance is the module instance address representing the root
// module, which is also the zero value of ModuleInstance.
var RootModuleInstance ModuleInstance

// IsRoot returns true if the receiver is the address of the root module instance,
// or false otherwise.
func (m ModuleInstance) IsRoot() bool {
	return len(m) == 0
}

// String returns a string representation of the receiver, in the format used
// within e.g. user-provided resource addresses.
//
// The address of the root module has the empty string as its representation.
func (m ModuleInstance) String() string {
	if len(m) == 0 {
		return ""
	}
	// Calculate minimal necessary space (no instance keys).
	l := 0
	for _, step := range m {
		l += len(step.Name)
	}
	buf := strings.Builder{}
	// 8 is len(".module.") which separates entries.
	buf.Grow(l + len(m)*8)
	sep := ""
	for _, step := range m {
		buf.WriteString(sep)
		buf.WriteString("module.")
		buf.WriteString(step.Name)
		if step.InstanceKey != NoKey {
			buf.WriteString(step.InstanceKey.String())
		}
		sep = "."
	}
	return buf.String()
}

func (s ModuleInstanceStep) String() string {
	if s.InstanceKey != NoKey {
		return s.Name + s.InstanceKey.String()
	}
	return s.Name
}
