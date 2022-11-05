package tfhcl

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type dynamicIteration struct {
	IteratorName string
	Key          cty.Value
	Value        cty.Value
	Inherited    map[string]*dynamicIteration
}

func (i *dynamicIteration) Object() cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"key":   i.Key,
		"value": i.Value,
	})
}

func (i *dynamicIteration) EvalContext(base *hcl.EvalContext) *hcl.EvalContext {
	new := base.NewChild()

	if i != nil {
		new.Variables = map[string]cty.Value{}
		for name, otherIt := range i.Inherited {
			new.Variables[name] = otherIt.Object()
		}
		new.Variables[i.IteratorName] = i.Object()
	}

	return new
}

func (i *dynamicIteration) MakeChild(iteratorName string, key, value cty.Value) *dynamicIteration {
	if i == nil {
		// Create entirely new root iteration, then
		return &dynamicIteration{
			IteratorName: iteratorName,
			Key:          key,
			Value:        value,
		}
	}

	inherited := map[string]*dynamicIteration{}
	for name, otherIt := range i.Inherited {
		inherited[name] = otherIt
	}
	inherited[i.IteratorName] = i
	return &dynamicIteration{
		IteratorName: iteratorName,
		Key:          key,
		Value:        value,
		Inherited:    inherited,
	}
}

type metaArgIteration struct {
	Count bool
	Index cty.Value

	ForEach bool
	Key     cty.Value
	Value   cty.Value
}

func MakeCountIteration(index cty.Value) *metaArgIteration {
	return &metaArgIteration{
		Count: true,
		Index: index,
	}
}

func MakeForEachIteration(key, value cty.Value) *metaArgIteration {
	return &metaArgIteration{
		ForEach: true,
		Key:     key,
		Value:   value,
	}
}

func (i *metaArgIteration) EvalContext(base *hcl.EvalContext) *hcl.EvalContext {
	new := base.NewChild()

	if i != nil {
		new.Variables = map[string]cty.Value{}

		if i.Count {
			new.Variables["count"] = cty.ObjectVal(map[string]cty.Value{
				"index": i.Index,
			})
		}
		if i.ForEach {
			new.Variables["each"] = cty.ObjectVal(map[string]cty.Value{
				"key":   i.Key,
				"value": i.Value,
			})
		}
	}

	return new
}
