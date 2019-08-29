package tflint

import "github.com/hashicorp/hcl2/hcl"

type moduleVariable struct {
	Root      bool
	Parents   []*moduleVariable
	Callers   []*moduleVariable
	DeclRange hcl.Range
}

func (m *moduleVariable) roots() []*moduleVariable {
	if m.Root {
		return []*moduleVariable{m}
	}

	ret := []*moduleVariable{}
	for _, parent := range m.Parents {
		for _, parentRoot := range parent.roots() {
			parentRoot.Callers = append(parentRoot.Callers, m)
			ret = append(ret, parentRoot)
		}
	}
	return ret
}

func (m *moduleVariable) callers() []hcl.Range {
	ret := make([]hcl.Range, len(m.Callers)+1)
	ret[0] = m.DeclRange

	for idx, caller := range m.Callers {
		ret[idx+1] = caller.DeclRange
	}
	return ret
}
