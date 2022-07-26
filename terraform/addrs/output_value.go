package addrs

// OutputValue is the address of an output value, in the context of the module
// that is defining it.
//
// This is related to but separate from ModuleCallOutput, which represents
// a module output from the perspective of its parent module. Since output
// values cannot be represented from the module where they are defined,
// OutputValue is not Referenceable, while ModuleCallOutput is.
type OutputValue struct {
	Name string
}

func (v OutputValue) String() string {
	return "output." + v.Name
}
