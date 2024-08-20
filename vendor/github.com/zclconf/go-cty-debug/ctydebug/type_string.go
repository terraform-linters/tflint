package ctydebug

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zclconf/go-cty/cty"
)

// TypeString returns a string representation of a given type that is
// reminiscent of Go syntax calling into the cty package but is mainly
// intended for easy human inspection of values in tests, debug output, etc.
//
// The resulting string will include newlines and indentation in order to
// increase the readability of complex structures. It always ends with a
// newline, so you can print this result directly to your output.
func TypeString(ty cty.Type) string {
	var b strings.Builder
	writeType(ty, &b, 0)
	b.WriteByte('\n') // always end with a newline for easier printing
	return b.String()
}

func writeType(ty cty.Type, b *strings.Builder, indent int) {
	switch {
	case ty == cty.NilType:
		b.WriteString("cty.NilType")
		return
	case ty.IsObjectType():
		atys := ty.AttributeTypes()
		if len(atys) == 0 {
			b.WriteString("cty.EmptyObject")
			return
		}
		attrNames := make([]string, 0, len(atys))
		for name := range atys {
			attrNames = append(attrNames, name)
		}
		sort.Strings(attrNames)
		b.WriteString("cty.Object(map[string]cty.Type{\n")
		indent++
		for _, name := range attrNames {
			aty := atys[name]
			b.WriteString(indentSpaces(indent))
			fmt.Fprintf(b, "%q: ", name)
			writeType(aty, b, indent)
			b.WriteString(",\n")
		}
		indent--
		b.WriteString(indentSpaces(indent))
		b.WriteString("})")
	case ty.IsTupleType():
		etys := ty.TupleElementTypes()
		if len(etys) == 0 {
			b.WriteString("cty.EmptyTuple")
			return
		}
		b.WriteString("cty.Tuple([]cty.Type{\n")
		indent++
		for _, ety := range etys {
			b.WriteString(indentSpaces(indent))
			writeType(ety, b, indent)
			b.WriteString(",\n")
		}
		indent--
		b.WriteString(indentSpaces(indent))
		b.WriteString("})")
	case ty.IsCollectionType():
		ety := ty.ElementType()
		switch {
		case ty.IsListType():
			b.WriteString("cty.List(")
		case ty.IsMapType():
			b.WriteString("cty.Map(")
		case ty.IsSetType():
			b.WriteString("cty.Set(")
		default:
			// At the time of writing there are no other collection types,
			// but we'll be robust here and just pass through the GoString
			// of anything we don't recognize.
			b.WriteString(ty.GoString())
			return
		}
		// Because object and tuple types render split over multiple
		// lines, a collection type container around them can end up
		// being hard to see when scanning, so we'll generate some extra
		// indentation to make a collection of structural type more visually
		// distinct from the structural type alone.
		complexElem := ety.IsObjectType() || ety.IsTupleType()
		if complexElem {
			indent++
			b.WriteString("\n")
			b.WriteString(indentSpaces(indent))
		}
		writeType(ty.ElementType(), b, indent)
		if complexElem {
			indent--
			b.WriteString(",\n")
			b.WriteString(indentSpaces(indent))
		}
		b.WriteString(")")
	default:
		// For any other type we'll just use its GoString and assume it'll
		// follow the usual GoString conventions.
		b.WriteString(ty.GoString())
	}
}

func indentSpaces(level int) string {
	return strings.Repeat("    ", level)
}
