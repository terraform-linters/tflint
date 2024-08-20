// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package typeexpr

import (
	"sort"
	"strconv"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

// Defaults represents a type tree which may contain default values for
// optional object attributes at any level. This is used to apply nested
// defaults to a given cty.Value before converting it to a concrete type.
type Defaults struct {
	// Type of the node for which these defaults apply. This is necessary in
	// order to determine how to inspect the Defaults and Children collections.
	Type cty.Type

	// DefaultValues contains the default values for each object attribute,
	// indexed by attribute name.
	DefaultValues map[string]cty.Value

	// Children is a map of Defaults for elements contained in this type. This
	// only applies to structural and collection types.
	//
	// The map is indexed by string instead of cty.Value because cty.Number
	// instances are non-comparable, due to embedding a *big.Float.
	//
	// Collections have a single element type, which is stored at key "".
	Children map[string]*Defaults
}

// Apply walks the given value, applying specified defaults wherever optional
// attributes are missing. The input and output values may have different
// types, and the result may still require type conversion to the final desired
// type.
//
// This function is permissive and does not report errors, assuming that the
// caller will have better context to report useful type conversion failure
// diagnostics.
func (d *Defaults) Apply(val cty.Value) cty.Value {
	return d.apply(val)
}

func (d *Defaults) apply(v cty.Value) cty.Value {
	// We don't apply defaults to null values or unknown values. To be clear,
	// we will overwrite children values with defaults if they are null but not
	// if the actual value is null.
	if !v.IsKnown() || v.IsNull() {
		return v
	}

	// Also, do nothing if we have no defaults to apply.
	if len(d.DefaultValues) == 0 && len(d.Children) == 0 {
		return v
	}

	v, marks := v.Unmark()

	switch {
	case v.Type().IsSetType(), v.Type().IsListType(), v.Type().IsTupleType():
		values := d.applyAsSlice(v)

		if v.Type().IsSetType() {
			if len(values) == 0 {
				v = cty.SetValEmpty(v.Type().ElementType())
				break
			}
			if converts := d.unifyAsSlice(values); len(converts) > 0 {
				v = cty.SetVal(converts).WithMarks(marks)
				break
			}
		} else if v.Type().IsListType() {
			if len(values) == 0 {
				v = cty.ListValEmpty(v.Type().ElementType())
				break
			}
			if converts := d.unifyAsSlice(values); len(converts) > 0 {
				v = cty.ListVal(converts)
				break
			}
		}
		v = cty.TupleVal(values)
	case v.Type().IsObjectType(), v.Type().IsMapType():
		values := d.applyAsMap(v)

		for key, defaultValue := range d.DefaultValues {
			if value, ok := values[key]; !ok || value.IsNull() {
				if defaults, ok := d.Children[key]; ok {
					values[key] = defaults.apply(defaultValue)
					continue
				}
				values[key] = defaultValue
			}
			if defaultRng := defaultValue.Range(); defaultRng.DefinitelyNotNull() && values[key].Type() != cty.DynamicPseudoType {
				values[key] = values[key].RefineNotNull()
			}
		}

		if v.Type().IsMapType() {
			if len(values) == 0 {
				v = cty.MapValEmpty(v.Type().ElementType())
				break
			}
			if converts := d.unifyAsMap(values); len(converts) > 0 {
				v = cty.MapVal(converts)
				break
			}
		}
		v = cty.ObjectVal(values)
	}

	return v.WithMarks(marks)
}

func (d *Defaults) applyAsSlice(value cty.Value) []cty.Value {
	var elements []cty.Value
	for ix, element := range value.AsValueSlice() {
		if childDefaults := d.getChild(ix); childDefaults != nil {
			element = childDefaults.apply(element)
			elements = append(elements, element)
			continue
		}
		elements = append(elements, element)
	}
	return elements
}

func (d *Defaults) applyAsMap(value cty.Value) map[string]cty.Value {
	elements := make(map[string]cty.Value)
	for key, element := range value.AsValueMap() {
		if childDefaults := d.getChild(key); childDefaults != nil {
			elements[key] = childDefaults.apply(element)
			continue
		}
		elements[key] = element
	}
	return elements
}

func (d *Defaults) getChild(key interface{}) *Defaults {
	// Children for tuples are keyed by an int.
	// Children for objects are keyed by a string.
	// Children for maps, lists, and sets are always keyed by the empty string.
	//
	// For maps and objects the supplied key type is a string type.
	// For lists, sets, and tuples the supplied key type is an int type.
	//
	// The callers of the defaults package could, in theory, pass a value in
	// where the types expected based on the defaults do not match the actual
	// type in the value. In this case, we get a mismatch between what the
	// defaults package expects the key to be, and which type it actually is.
	//
	// In the event of such a mismatch, we just won't apply defaults. Instead,
	// relying on the user later calling go-cty.Convert to detect this same
	// error (as  documented). In this case we'd just want to return nil to
	// indicate either there are no defaults or we can't work out how to apply
	// them. Both of  these outcomes are treated the same by the rest of the
	// package.
	//
	// For the above reasons it isn't necessarily safe to just rely on a single
	// metric for working out how we should retrieve the key. If the defaults
	// type is a tuple we can't just assume the supplied key will be an int (as
	// the concrete value actually supplied by the user could be an object or a
	// map). Similarly, if the supplied key is an int we can't just assume we
	// should treat the type as a tuple (as a list would also specify an int,
	// but we should return the children keyed by the empty string rather than
	// the index).

	switch concrete := key.(type) {
	case int:
		if d.Type.IsTupleType() {
			// If the type is an int, and our defaults are expecting a tuple
			// then we return the children for the tuple at the index.
			return d.Children[strconv.Itoa(concrete)]
		}
	case string:
		if d.Type.IsObjectType() {
			// If the type is a string, and our defaults are expecting an object
			// then we return the children for the object at the key.
			return d.Children[concrete]
		}
	}

	// Otherwise, either our defaults are expecting this to be a map, list, or
	// set or the type our defaults expecting didn't line up with something we
	// can convert between. So, either we want to return the child keyed by
	// the empty string (in the first case) or nil (in the second case).
	// Luckily, Golang maps return nil when referencing something that doesn't
	// exist. So, we can just try and retrieve the child at the empty key and
	// if it doesn't exist then that's fine since we'd just return nil anyway.

	return d.Children[""]
}

func (d *Defaults) unifyAsSlice(values []cty.Value) []cty.Value {
	var types []cty.Type
	for _, value := range values {
		types = append(types, value.Type())
	}
	unify, conversions := convert.UnifyUnsafe(types)
	if unify == cty.NilType {
		return nil
	}

	var converts []cty.Value
	for ix := 0; ix < len(conversions); ix++ {
		if conversions[ix] == nil {
			converts = append(converts, values[ix])
			continue
		}

		converted, err := conversions[ix](values[ix])
		if err != nil {
			return nil
		}
		converts = append(converts, converted)
	}
	return converts
}

func (d *Defaults) unifyAsMap(values map[string]cty.Value) map[string]cty.Value {
	var keys []string
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var types []cty.Type
	for _, key := range keys {
		types = append(types, values[key].Type())
	}
	unify, conversions := convert.UnifyUnsafe(types)
	if unify == cty.NilType {
		return nil
	}

	converts := make(map[string]cty.Value)
	for ix, key := range keys {
		if conversions[ix] == nil {
			converts[key] = values[key]
			continue
		}

		var err error
		if converts[key], err = conversions[ix](values[key]); err != nil {
			return nil
		}
	}
	return converts
}
