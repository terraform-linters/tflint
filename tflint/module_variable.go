package tflint

import "github.com/hashicorp/hcl2/hcl"

type moduleVariable struct {
	Root      bool
	Parents   []*moduleVariable
	DeclRange hcl.Range
}

func (m *moduleVariable) roots() []*moduleVariable {
	if m.Root {
		return []*moduleVariable{m}
	}

	ret := []*moduleVariable{}
	for _, parent := range m.Parents {
		ret = append(ret, parent.roots()...)
	}
	return ret
}
