package addrs

import (
	"fmt"
)

// ModuleCall is the address of a call from the current module to a child
// module.
type ModuleCall struct {
	referenceable
	Name string
}

func (c ModuleCall) String() string {
	return "module." + c.Name
}

// ModuleCallInstance is the address of one instance of a module created from
// a module call, which might create multiple instances using "count" or
// "for_each" arguments.
type ModuleCallInstance struct {
	referenceable
	Call ModuleCall
	Key  InstanceKey
}

func (c ModuleCallInstance) String() string {
	if c.Key == NoKey {
		return c.Call.String()
	}
	return fmt.Sprintf("module.%s%s", c.Call.Name, c.Key)
}

// ModuleCallInstanceOutput is the address of a particular named output produced by
// an instance of a module call.
type ModuleCallInstanceOutput struct {
	referenceable
	Call ModuleCallInstance
	Name string
}

func (co ModuleCallInstanceOutput) String() string {
	return fmt.Sprintf("%s.%s", co.Call.String(), co.Name)
}
