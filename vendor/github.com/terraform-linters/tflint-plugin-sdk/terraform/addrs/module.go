package addrs

import "strings"

// Module represents the structure of the module tree.
type Module []string

// IsRoot returns true if the receiver is the address of the root module,
// or false otherwise.
func (m Module) IsRoot() bool {
	return len(m) == 0
}

// String returns a string representation.
func (m Module) String() string {
	if len(m) == 0 {
		return ""
	}
	var steps []string
	for _, s := range m {
		steps = append(steps, "module", s)
	}
	return strings.Join(steps, ".")
}
