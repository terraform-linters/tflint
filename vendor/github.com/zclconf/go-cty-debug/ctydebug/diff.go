package ctydebug

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty/cty"
)

// DiffValues returns a human-oriented description of the differences between
// the two given values. It's guaranteed to return an empty string if the
// two values are RawEqual.
//
// Don't depend on the exact formatting of the result. It is likely to change
// in future releases.
func DiffValues(want, got cty.Value) string {
	// want and got are in the order they are here because that's how cmp
	// seems to treat them, and we'd like to be consistent with cmp to
	// minimize confusion in codebases that are using both cmp directly and
	// indirectly via DiffValues.

	if got.RawEquals(want) {
		return "" // just to make sure
	}

	r := &diffValuesReporter{}
	cmp.Equal(want, got, CmpOptions, cmp.Reporter(r))

	return r.Result()
}

// This is a very simple reporter for now. Hopefully one day it can become
// more sophisticated and produce output that looks more like the result
// of ValueString.
type diffValuesReporter struct {
	path cmp.Path
	sb   strings.Builder
}

func (r *diffValuesReporter) PushStep(step cmp.PathStep) {
	r.path = append(r.path, step)
}

func (r *diffValuesReporter) PopStep() {
	r.path = r.path[:len(r.path)-1]
}

func (r *diffValuesReporter) Report(result cmp.Result) {
	if result.Equal() {
		return
	}

	r.sb.WriteString(cmpPathString(r.path))
	r.sb.WriteString("\n")
	want, got := r.path.Last().Values()
	fmt.Fprintf(&r.sb, "  got:  %s\n", resultValueString(got))
	fmt.Fprintf(&r.sb, "  want: %s\n", resultValueString(want))
	r.sb.WriteString("\n")
}

func (r *diffValuesReporter) Result() string {
	return r.sb.String()
}

func resultValueString(rv reflect.Value) string {
	if !rv.IsValid() {
		return "(no value)"
	}
	if v, ok := rv.Interface().(cty.Value); ok && v == cty.NilVal {
		return "cty.NilVal"
	}
	if ty, ok := rv.Interface().(cty.Type); ok && ty == cty.NilType {
		return "cty.NilType"
	}
	return fmt.Sprintf("%#v", rv)
}

// cmpPathString returns the given path serialized using a compact syntax
// that isn't in any language exactly but is hopefully intuitive.
func cmpPathString(path cmp.Path) string {
	var b strings.Builder
	for _, step := range path {
		switch step := step.(type) {
		case cmp.Transform:
			if step.Option() == transformValueOp || step.Option() == transformTypeOp {
				continue // ignore; it's an implementation detail
			}
			b.WriteString(step.String())
		case cmp.TypeAssertion:
			// These show up on the results of the transforms we do to trick
			// cmp into walking into our structural/collection values, but
			// that's an implementation detail so we'll skip it.
			continue
		case cmp.Indirect:
			continue
		case cmp.MapIndex:
			fmt.Fprintf(&b, "[%q]", step.Key())
		case cmp.SliceIndex:
			fmt.Fprintf(&b, "[%d]", step.Key())
		case cmp.StructField:
			// We don't expect to see any struct field traversals in our
			// work because of our transformations, but if one shows up then
			// we'll handle it somewhat gracefully...
			fmt.Fprintf(&b, ".%s", step.Name())
		default:
			b.WriteString(step.String())
		}
	}
	return b.String()
}
