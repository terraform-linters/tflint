package ctydebug

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zclconf/go-cty/cty"
)

// ValueString returns a string representation of a given value that is
// reminiscent of Go syntax calling into the cty package but is mainly
// intended for easy human inspection of values in tests, debug output, etc.
//
// The resulting string will include newlines and indentation in order to
// increase the readability of complex structures. It always ends with a
// newline, so you can print this result directly to your output.
func ValueString(v cty.Value) string {
	var b strings.Builder
	writeValue(v, &b, 0)
	b.WriteByte('\n') // always end with a newline for easier printing
	return b.String()
}

func writeValue(v cty.Value, b *strings.Builder, indent int) {
	if v == cty.NilVal {
		b.WriteString("cty.NilVal")
		return
	}

	if v.IsMarked() {
		v, marks := v.Unmark()
		writeValue(v, b, indent)
		if len(marks) == 1 {
			var onlyMark interface{}
			for k := range marks {
				onlyMark = k
			}
			fmt.Fprintf(b, ".Mark(%#v)", onlyMark)
		} else {
			fmt.Fprintf(b, ".WithMarks(%#v)", marks)
		}
		return
	}

	ty := v.Type()
	switch {
	case v.IsNull():
		b.WriteString("cty.NullVal(")
		writeType(ty, b, indent)
		b.WriteString(")")
	case !v.IsKnown():
		b.WriteString("cty.UnknownVal(")
		writeType(ty, b, indent)
		b.WriteString(")")
	case ty.IsObjectType() || ty.IsMapType():
		attrs := v.AsValueMap()
		if len(attrs) == 0 {
			switch {
			case ty.IsObjectType():
				b.WriteString("cty.EmptyObjectVal")
			case ty.IsMapType():
				b.WriteString("cty.MapValEmpty(")
				writeType(ty.ElementType(), b, indent)
				b.WriteString(")")
			default:
				b.WriteString(v.GoString())
			}
			return
		}
		attrNames := make([]string, 0, len(attrs))
		for name := range attrs {
			attrNames = append(attrNames, name)
		}
		sort.Strings(attrNames)
		switch {
		case ty.IsObjectType():
			b.WriteString("cty.ObjectVal(map[string]cty.Value{\n")
		case ty.IsMapType():
			b.WriteString("cty.MapVal(map[string]cty.Value{\n")
		default:
			b.WriteString(v.GoString())
			return
		}
		indent++
		for _, name := range attrNames {
			av := attrs[name]
			b.WriteString(indentSpaces(indent))
			fmt.Fprintf(b, "%q: ", name)
			writeValue(av, b, indent)
			b.WriteString(",\n")
		}
		indent--
		b.WriteString(indentSpaces(indent))
		b.WriteString("})")
	case ty.IsTupleType() || ty.IsListType() || ty.IsSetType():
		elems := v.AsValueSlice()
		if len(elems) == 0 {
			switch {
			case ty.IsTupleType():
				b.WriteString("cty.EmptyTupleVal")
			case ty.IsListType():
				b.WriteString("cty.ListValEmpty(")
				writeType(ty.ElementType(), b, indent)
				b.WriteString(")")
			case ty.IsSetType():
				b.WriteString("cty.SetValEmpty(")
				writeType(ty.ElementType(), b, indent)
				b.WriteString(")")
			default:
				b.WriteString(v.GoString())
			}
			return
		}
		switch {
		case ty.IsTupleType():
			b.WriteString("cty.TupleVal([]cty.Value{\n")
		case ty.IsListType():
			b.WriteString("cty.ListVal([]cty.Value{\n")
		case ty.IsSetType():
			b.WriteString("cty.SetVal([]cty.Value{\n")
		default:
			b.WriteString(v.GoString())
			return
		}
		indent++
		for _, ev := range elems {
			b.WriteString(indentSpaces(indent))
			writeValue(ev, b, indent)
			b.WriteString(",\n")
		}
		indent--
		b.WriteString(indentSpaces(indent))
		b.WriteString("})")
	default:
		// For any other type we'll just use its GoString and assume it'll
		// follow the usual GoString conventions.
		b.WriteString(v.GoString())
	}
}
