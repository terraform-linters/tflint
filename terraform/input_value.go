package terraform

import "github.com/zclconf/go-cty/cty"

type InputValue struct {
	Value cty.Value
}

type InputValues map[string]*InputValue

func (vv InputValues) Override(others ...InputValues) InputValues {
	ret := make(InputValues)
	for k, v := range vv {
		ret[k] = v
	}
	for _, other := range others {
		for k, v := range other {
			ret[k] = v
		}
	}
	return ret
}
