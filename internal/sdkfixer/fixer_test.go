// Derived from tflint-plugin-sdk v0.25.0/internal/fixer_test.go.
// Licensed under MPL-2.0; see LICENSE.
package sdkfixer

import (
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

func TestReplaceText(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	tests := []struct {
		name     string
		sources  map[string]string
		fix      func(*Fixer) error
		want     map[string]string
		errCheck func(error) bool
	}{
		{
			name: "no change",
			sources: map[string]string{
				"main.tf": "// comment",
			},
			fix: func(fixer *Fixer) error {
				return nil
			},
			want:     map[string]string{},
			errCheck: neverHappend,
		},
		{
			name: "no shift",
			sources: map[string]string{
				"main.tf": "// comment",
			},
			fix: func(fixer *Fixer) error {
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 0}, End: hcl.Pos{Byte: 2}}, "##")
			},
			want: map[string]string{
				"main.tf": "## comment",
			},
			errCheck: neverHappend,
		},
		{
			name: "shift left",
			sources: map[string]string{
				"main.tf": "// comment",
			},
			fix: func(fixer *Fixer) error {
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 0}, End: hcl.Pos{Byte: 2}}, "#")
			},
			want: map[string]string{
				"main.tf": "# comment",
			},
			errCheck: neverHappend,
		},
		{
			name: "shift right",
			sources: map[string]string{
				"main.tf": "# comment",
			},
			fix: func(fixer *Fixer) error {
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 0}, End: hcl.Pos{Byte: 1}}, "//")
			},
			want: map[string]string{
				"main.tf": "// comment",
			},
			errCheck: neverHappend,
		},
		{
			name: "no shift + shift left",
			sources: map[string]string{
				"main.tf": `
// comment
// comment2`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 3}}, "##")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 12}, End: hcl.Pos{Byte: 14}}, "#")
			},
			want: map[string]string{
				"main.tf": `
## comment
# comment2`,
			},
			errCheck: neverHappend,
		},
		{
			name: "no shift + shift right",
			sources: map[string]string{
				"main.tf": `
## comment
# comment2`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 3}}, "//")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 12}, End: hcl.Pos{Byte: 13}}, "//")
			},
			want: map[string]string{
				"main.tf": `
// comment
// comment2`,
			},
			errCheck: neverHappend,
		},
		{
			name: "shift left + shift left",
			sources: map[string]string{
				"main.tf": `
// comment
// comment2`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 3}}, "#")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 12}, End: hcl.Pos{Byte: 14}}, "#")
			},
			want: map[string]string{
				"main.tf": `
# comment
# comment2`,
			},
			errCheck: neverHappend,
		},
		{
			name: "shift left + shift right",
			sources: map[string]string{
				"main.tf": `
// comment
# comment2`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 3}}, "#")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 12}, End: hcl.Pos{Byte: 13}}, "//")
			},
			want: map[string]string{
				"main.tf": `
# comment
// comment2`,
			},
			errCheck: neverHappend,
		},
		{
			name: "shift right + shift left",
			sources: map[string]string{
				"main.tf": `
# comment
// comment2`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 2}}, "//")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 11}, End: hcl.Pos{Byte: 13}}, "#")
			},
			want: map[string]string{
				"main.tf": `
// comment
# comment2`,
			},
			errCheck: neverHappend,
		},
		{
			name: "shift right + shift right",
			sources: map[string]string{
				"main.tf": `
# comment
# comment2`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 2}}, "//")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 11}, End: hcl.Pos{Byte: 12}}, "//")
			},
			want: map[string]string{
				"main.tf": `
// comment
// comment2`,
			},
			errCheck: neverHappend,
		},
		{
			name: "shift left + shift left + shift left",
			sources: map[string]string{
				"main.tf": `
// comment
// comment2
// comment3`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 3}}, "#")
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 12}, End: hcl.Pos{Byte: 14}}, "#")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 24}, End: hcl.Pos{Byte: 26}}, "#")
			},
			want: map[string]string{
				"main.tf": `
# comment
# comment2
# comment3`,
			},
			errCheck: neverHappend,
		},
		{
			name: "shift left + shift left + shift right",
			sources: map[string]string{
				"main.tf": `
// comment
// comment2
# comment3`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 3}}, "#")
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 12}, End: hcl.Pos{Byte: 14}}, "#")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 24}, End: hcl.Pos{Byte: 25}}, "//")
			},
			want: map[string]string{
				"main.tf": `
# comment
# comment2
// comment3`,
			},
			errCheck: neverHappend,
		},
		{
			name: "shift left + shift right + shift left",
			sources: map[string]string{
				"main.tf": `
// comment
# comment2
// comment3`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 3}}, "#")
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 12}, End: hcl.Pos{Byte: 13}}, "//")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 23}, End: hcl.Pos{Byte: 25}}, "#")
			},
			want: map[string]string{
				"main.tf": `
# comment
// comment2
# comment3`,
			},
			errCheck: neverHappend,
		},
		{
			name: "shift left + shift right + shift right",
			sources: map[string]string{
				"main.tf": `
// comment
# comment2
# comment3`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 3}}, "#")
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 12}, End: hcl.Pos{Byte: 13}}, "//")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 23}, End: hcl.Pos{Byte: 24}}, "//")
			},
			want: map[string]string{
				"main.tf": `
# comment
// comment2
// comment3`,
			},
			errCheck: neverHappend,
		},
		{
			name: "shift right + shift left + shift left",
			sources: map[string]string{
				"main.tf": `
# comment
// comment2
// comment3`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 2}}, "//")
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 11}, End: hcl.Pos{Byte: 13}}, "#")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 23}, End: hcl.Pos{Byte: 25}}, "#")
			},
			want: map[string]string{
				"main.tf": `
// comment
# comment2
# comment3`,
			},
			errCheck: neverHappend,
		},
		{
			name: "shift right + shift left + shift right",
			sources: map[string]string{
				"main.tf": `
# comment
// comment2
# comment3`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 2}}, "//")
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 11}, End: hcl.Pos{Byte: 13}}, "#")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 23}, End: hcl.Pos{Byte: 24}}, "//")
			},
			want: map[string]string{
				"main.tf": `
// comment
# comment2
// comment3`,
			},
			errCheck: neverHappend,
		},
		{
			name: "shift right + shift right + shift left",
			sources: map[string]string{
				"main.tf": `
# comment
# comment2
// comment3`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 2}}, "//")
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 11}, End: hcl.Pos{Byte: 12}}, "//")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 22}, End: hcl.Pos{Byte: 24}}, "#")
			},
			want: map[string]string{
				"main.tf": `
// comment
// comment2
# comment3`,
			},
			errCheck: neverHappend,
		},
		{
			name: "shift right + shift right + shift right",
			sources: map[string]string{
				"main.tf": `
# comment
# comment2
# comment3`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 2}}, "//")
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 11}, End: hcl.Pos{Byte: 12}}, "//")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 22}, End: hcl.Pos{Byte: 23}}, "//")
			},
			want: map[string]string{
				"main.tf": `
// comment
// comment2
// comment3`,
			},
			errCheck: neverHappend,
		},
		{
			name: "change order",
			sources: map[string]string{
				"main.tf": `
# comment
# comment2
# comment3`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 11}, End: hcl.Pos{Byte: 12}}, "//")
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 2}}, "//")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 22}, End: hcl.Pos{Byte: 23}}, "//")
			},
			want: map[string]string{
				"main.tf": `
// comment
// comment2
// comment3`,
			},
			errCheck: neverHappend,
		},
		{
			name: "shift left (boundary)",
			sources: map[string]string{
				"main.tf": `"Hellooo, world!"`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 8}}, "Hello")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 8}, End: hcl.Pos{Byte: 16}}, ", you and world!")
			},
			want: map[string]string{
				"main.tf": `"Hello, you and world!"`,
			},
			errCheck: neverHappend,
		},
		{
			name: "shift right (boundary)",
			sources: map[string]string{
				"main.tf": `"Hello, world!"`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 6}}, "Hellooo")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 6}, End: hcl.Pos{Byte: 14}}, ", you and world!")
			},
			want: map[string]string{
				"main.tf": `"Hellooo, you and world!"`,
			},
			errCheck: neverHappend,
		},
		{
			name: "overlapping",
			sources: map[string]string{
				"main.tf": `"Hello, world!"`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 8}}, "Hellooo, ")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 6}, End: hcl.Pos{Byte: 14}}, ", you and world!")
			},
			want: map[string]string{
				"main.tf": `"Hellooo, world!"`,
			},
			errCheck: func(err error) bool {
				return err == nil || err.Error() != "range overlaps with a previous rewrite range: main.tf:0,0-0"
			},
		},
		{
			name: "same range",
			sources: map[string]string{
				"main.tf": `"Hello, world!"`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 6}}, "hello")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 6}}, "HELLO")
			},
			want: map[string]string{
				"main.tf": `"HELLO, world!"`,
			},
			errCheck: neverHappend,
		},
		{
			name: "same range (shift left)",
			sources: map[string]string{
				"main.tf": `"Hellooo, world!"`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 8}}, "hello")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 8}}, "HELLO")
			},
			want: map[string]string{
				"main.tf": `"HELLO, world!"`,
			},
			errCheck: neverHappend,
		},
		{
			name: "same range (shift right)",
			sources: map[string]string{
				"main.tf": `"Hello, world!"`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 6}}, "hellooo")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 6}}, "HELLOOO")
			},
			want: map[string]string{
				"main.tf": `"HELLOOO, world!"`,
			},
			errCheck: neverHappend,
		},
		{
			name: "shift after same range",
			sources: map[string]string{
				"main.tf": `"Hello, world!"`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 6}}, "hellooo")
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 6}}, "HELLOOO")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 6}, End: hcl.Pos{Byte: 14}}, ", you and world!")
			},
			want: map[string]string{
				"main.tf": `"HELLOOO, you and world!"`,
			},
			errCheck: neverHappend,
		},
		{
			name: "same range after shift",
			sources: map[string]string{
				"main.tf": `"Hello, world!"`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 6}}, "hellooo")
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 8}, End: hcl.Pos{Byte: 13}}, "wooorld")
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 6}}, "Hellooo")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 8}, End: hcl.Pos{Byte: 13}}, "Wooorld")
			},
			want: map[string]string{
				"main.tf": `"Hellooo, Wooorld!"`,
			},
			errCheck: neverHappend,
		},
		{
			name: "multibyte",
			sources: map[string]string{
				"main.tf": `"Hello, world!"`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 6}}, "こんにちは")
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 8}, End: hcl.Pos{Byte: 13}}, "世界")
			},
			want: map[string]string{
				"main.tf": `"こんにちは, 世界!"`,
			},
			errCheck: neverHappend,
		},
		{
			name: "file not found",
			sources: map[string]string{
				"main.tf": `"Hello, world!"`,
			},
			fix: func(fixer *Fixer) error {
				return fixer.ReplaceText(hcl.Range{Filename: "template.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 6}}, "hello")
			},
			want: map[string]string{},
			errCheck: func(err error) bool {
				return err == nil || err.Error() != "file not found: template.tf"
			},
		},
		{
			name: "multiple string literals",
			sources: map[string]string{
				"main.tf": `(foo)(bar)`,
			},
			fix: func(fixer *Fixer) error {
				return fixer.ReplaceText(
					hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 0}, End: hcl.Pos{Byte: 10}},
					"[",
					"foo",
					"]",
					"[",
					"bar",
					"]",
				)
			},
			want: map[string]string{
				"main.tf": `[foo][bar]`,
			},
			errCheck: neverHappend,
		},
		{
			name: "literals with text nodes",
			sources: map[string]string{
				"main.tf": `(foo)(bar)`,
			},
			fix: func(fixer *Fixer) error {
				if err := fixer.ReplaceText(
					hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 0}, End: hcl.Pos{Byte: 10}},
					"[",
					tflint.TextNode{Bytes: []byte("foo"), Range: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 4}}},
					"]",
					"[",
					tflint.TextNode{Bytes: []byte("bar"), Range: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 6}, End: hcl.Pos{Byte: 9}}},
					"]",
				); err != nil {
					return err
				}
				// The replacement is not overlapped because the "foo" is not replaced in the previous replacement.
				if err := fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 4}}, "bar"); err != nil {
					return err
				}
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 6}, End: hcl.Pos{Byte: 9}}, "baz")
			},
			want: map[string]string{
				"main.tf": `[bar][baz]`,
			},
			errCheck: neverHappend,
		},
		{
			name: "only text nodes",
			sources: map[string]string{
				"main.tf": `(foo)(bar)`,
			},
			fix: func(fixer *Fixer) error {
				if err := fixer.ReplaceText(
					hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 0}, End: hcl.Pos{Byte: 10}},
					tflint.TextNode{Bytes: []byte("foo"), Range: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 4}}},
					tflint.TextNode{Bytes: []byte("bar"), Range: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 6}, End: hcl.Pos{Byte: 9}}},
				); err != nil {
					return err
				}
				// The replacement is not overlapped because the "foo" is not replaced in the previous replacement.
				if err := fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 4}}, "bar"); err != nil {
					return err
				}
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 6}, End: hcl.Pos{Byte: 9}}, "baz")
			},
			want: map[string]string{
				"main.tf": `barbaz`,
			},
			errCheck: neverHappend,
		},
		{
			name: "unordered text nodes",
			sources: map[string]string{
				"main.tf": `(foo)(bar)`,
			},
			fix: func(fixer *Fixer) error {
				if err := fixer.ReplaceText(
					hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 0}, End: hcl.Pos{Byte: 10}},
					"[",
					tflint.TextNode{Bytes: []byte("bar"), Range: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 6}, End: hcl.Pos{Byte: 9}}},
					"]",
					"[",
					tflint.TextNode{Bytes: []byte("foo"), Range: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 4}}},
					"]",
				); err != nil {
					return err
				}
				// The replacement is not overlapped because the "bar" is not replaced in the previous replacement.
				if err := fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 6}, End: hcl.Pos{Byte: 9}}, "baz"); err != nil {
					return err
				}
				// The replacement is overlapped because the "foo" is replaced in the previous replacement.
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 4}}, "bar")
			},
			want: map[string]string{
				"main.tf": `[baz][foo]`,
			},
			errCheck: func(err error) bool {
				return err == nil || err.Error() != "range overlaps with a previous rewrite range: main.tf:0,0-0"
			},
		},
		{
			name: "out of range text node",
			sources: map[string]string{
				"main.tf": `foo`,
			},
			fix: func(fixer *Fixer) error {
				return fixer.ReplaceText(
					hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 0}, End: hcl.Pos{Byte: 3}},
					tflint.TextNode{Bytes: []byte("baz"), Range: hcl.Range{Filename: "template.tf", Start: hcl.Pos{Byte: 0}, End: hcl.Pos{Byte: 3}}},
				)
			},
			want: map[string]string{
				"main.tf": `baz`,
			},
			errCheck: neverHappend,
		},
		{
			name: "text node with the same range",
			sources: map[string]string{
				"main.tf": `foo`,
			},
			fix: func(fixer *Fixer) error {
				return fixer.ReplaceText(
					hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 0}, End: hcl.Pos{Byte: 3}},
					tflint.TextNode{Bytes: []byte("foo"), Range: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 0}, End: hcl.Pos{Byte: 3}}},
				)
			},
			want:     map[string]string{},
			errCheck: neverHappend,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := map[string][]byte{}
			for filename, source := range test.sources {
				input[filename] = []byte(source)
			}
			fixer := NewFixer(input)

			err := test.fix(fixer)
			if test.errCheck(err) {
				t.Fatalf("failed to check error: %s", err)
			}

			changes := map[string]string{}
			for filename, source := range fixer.changes {
				changes[filename] = string(source)
			}
			if diff := cmp.Diff(test.want, changes); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestInsertText(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	tests := []struct {
		name     string
		sources  map[string]string
		fix      func(*Fixer) error
		want     map[string]string
		errCheck func(error) bool
	}{
		{
			name: "insert before by InsertTextBefore",
			sources: map[string]string{
				"main.tf": `"world!"`,
			},
			fix: func(fixer *Fixer) error {
				return fixer.InsertTextBefore(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 6}}, "Hello, ")
			},
			want: map[string]string{
				"main.tf": `"Hello, world!"`,
			},
			errCheck: neverHappend,
		},
		{
			name: "insert after by InsertTextBefore",
			sources: map[string]string{
				"main.tf": `"Hello"`,
			},
			fix: func(fixer *Fixer) error {
				return fixer.InsertTextBefore(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 6}, End: hcl.Pos{Byte: 7}}, ", world!")
			},
			want: map[string]string{
				"main.tf": `"Hello, world!"`,
			},
			errCheck: neverHappend,
		},
		{
			name: "insert before by InsertTextAfter",
			sources: map[string]string{
				"main.tf": `"world!"`,
			},
			fix: func(fixer *Fixer) error {
				return fixer.InsertTextAfter(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 0}, End: hcl.Pos{Byte: 1}}, "Hello, ")
			},
			want: map[string]string{
				"main.tf": `"Hello, world!"`,
			},
			errCheck: neverHappend,
		},
		{
			name: "insert after by InsertTextAfter",
			sources: map[string]string{
				"main.tf": `"Hello"`,
			},
			fix: func(fixer *Fixer) error {
				return fixer.InsertTextAfter(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 6}}, ", world!")
			},
			want: map[string]string{
				"main.tf": `"Hello, world!"`,
			},
			errCheck: neverHappend,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := map[string][]byte{}
			for filename, source := range test.sources {
				input[filename] = []byte(source)
			}
			fixer := NewFixer(input)

			err := test.fix(fixer)
			if test.errCheck(err) {
				t.Fatalf("failed to check error: %s", err)
			}

			changes := map[string]string{}
			for filename, source := range fixer.changes {
				changes[filename] = string(source)
			}
			if diff := cmp.Diff(test.want, changes); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	tests := []struct {
		name     string
		sources  map[string]string
		fix      func(*Fixer) error
		want     map[string]string
		errCheck func(error) bool
	}{
		{
			name: "remove",
			sources: map[string]string{
				"main.tf": `"Hello, world!"`,
			},
			fix: func(fixer *Fixer) error {
				return fixer.Remove(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 6}, End: hcl.Pos{Byte: 14}})
			},
			want: map[string]string{
				"main.tf": `"Hello"`,
			},
			errCheck: neverHappend,
		},
		{
			name: "remove and shift",
			sources: map[string]string{
				"main.tf": `"Hello, world!"`,
			},
			fix: func(fixer *Fixer) error {
				_ = fixer.Remove(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 1}, End: hcl.Pos{Byte: 8}})
				return fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 8}, End: hcl.Pos{Byte: 14}}, "WORLD!!")
			},
			want: map[string]string{
				"main.tf": `"WORLD!!"`,
			},
			errCheck: neverHappend,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := map[string][]byte{}
			for filename, source := range test.sources {
				input[filename] = []byte(source)
			}
			fixer := NewFixer(input)

			err := test.fix(fixer)
			if test.errCheck(err) {
				t.Fatalf("failed to check error: %s", err)
			}

			changes := map[string]string{}
			for filename, source := range fixer.changes {
				changes[filename] = string(source)
			}
			if diff := cmp.Diff(test.want, changes); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestRemoveAttribute(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }
	// helper to get "foo" attribute in locals
	getFooAttributeInLocals := func(body hcl.Body) (*hcl.Attribute, hcl.Diagnostics) {
		content, _, diags := body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{{Type: "locals"}},
		})
		if diags.HasErrors() {
			return nil, diags
		}
		attributes, diags := content.Blocks[0].Body.JustAttributes()
		if diags.HasErrors() {
			return nil, diags
		}
		return attributes["foo"], nil
	}

	tests := []struct {
		name     string
		source   string
		getAttr  func(hcl.Body) (*hcl.Attribute, hcl.Diagnostics)
		want     string
		errCheck func(error) bool
	}{
		{
			name:   "remove attribute",
			source: `foo = 1`,
			getAttr: func(body hcl.Body) (*hcl.Attribute, hcl.Diagnostics) {
				attributes, diags := body.JustAttributes()
				if diags.HasErrors() {
					return nil, diags
				}
				return attributes["foo"], nil
			},
			want:     ``,
			errCheck: neverHappend,
		},
		{
			name: "remove attribute within block",
			source: `
locals {
  foo = 1
}`,
			getAttr: getFooAttributeInLocals,
			want: `
locals {
}`,
			errCheck: neverHappend,
		},
		{
			name: "remove attribute with trailing comment",
			source: `
locals {
  foo = 1 # comment
}`,
			getAttr: getFooAttributeInLocals,
			want: `
locals {
}`,
			errCheck: neverHappend,
		},
		{
			name: "remove attribute with trailing legacy comment",
			source: `
locals {
  foo = 1 // comment
}`,
			getAttr: getFooAttributeInLocals,
			want: `
locals {
}`,
			errCheck: neverHappend,
		},
		{
			name: "remove attribute with trailing multiline comment",
			source: `
locals {
  foo = 1 /* comment */
}`,
			getAttr: getFooAttributeInLocals,
			want: `
locals {
}`,
			errCheck: neverHappend,
		},
		{
			name: "remove attribute with next line comment",
			source: `
locals {
  foo = 1
  # comment
}`,
			getAttr: getFooAttributeInLocals,
			want: `
locals {
  # comment
}`,
			errCheck: neverHappend,
		},
		{
			name: "remove attribute with prefix comment",
			source: `
locals {
/* comment */ foo = 1
}`,
			getAttr: getFooAttributeInLocals,
			want: `
locals {
}`,
			errCheck: neverHappend,
		},
		{
			name: "remove attribute with previous line comment",
			source: `
locals {
  # comment
  foo = 1
}`,
			getAttr: getFooAttributeInLocals,
			want: `
locals {
}`,
			errCheck: neverHappend,
		},
		{
			name: "remove attribute with previous multiple line comments",
			source: `
locals {
  # comment
  # comment
  foo = 1
}`,
			getAttr: getFooAttributeInLocals,
			want: `
locals {
}`,
			errCheck: neverHappend,
		},
		{
			name: "remove attribute with previous multiline comment",
			source: `
locals {
  /* comment */
  foo = 1
}`,
			getAttr: getFooAttributeInLocals,
			// This is the same behavior as hclwrite.RemoveAttribute.
			want: `
locals {
  /* comment */
}`,
			errCheck: neverHappend,
		},
		{
			name: "remove attribute after attribute with trailing comment",
			source: `
locals {
  bar = 1 # comment
  foo = 1
}`,
			getAttr: getFooAttributeInLocals,
			want: `
locals {
  bar = 1 # comment
}`,
			errCheck: neverHappend,
		},
		{
			name:     "remove attribute within inline block",
			source:   `locals { foo = 1 }`,
			getAttr:  getFooAttributeInLocals,
			want:     `locals {}`,
			errCheck: neverHappend,
		},
		{
			name: "remove attribute in the middle of attributes",
			source: `
locals {
  bar = 1
  foo = 1
  baz = 1
}`,
			getAttr: getFooAttributeInLocals,
			want: `
locals {
  bar = 1
  baz = 1
}`,
			errCheck: neverHappend,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file, diags := hclsyntax.ParseConfig([]byte(test.source), "main.tf", hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatalf("failed to parse HCL: %s", diags)
			}
			attr, diags := test.getAttr(file.Body)
			if diags.HasErrors() {
				t.Fatalf("failed to get attribute: %s", diags)
			}

			fixer := NewFixer(map[string][]byte{"main.tf": []byte(test.source)})

			err := fixer.RemoveAttribute(attr)
			if test.errCheck(err) {
				t.Fatalf("failed to check error: %s", err)
			}

			if diff := cmp.Diff(test.want, string(fixer.changes["main.tf"])); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestRemoveBlock(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }
	// getFirstBlock returns the first block in the given body.
	getFirstBlock := func(body hcl.Body) (*hcl.Block, hcl.Diagnostics) {
		content, _, diags := body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{{Type: "block"}},
		})
		if diags.HasErrors() {
			return nil, diags
		}
		return content.Blocks[0], nil
	}
	// getNestedBlock returns the nested block in the given body.
	getNestedBlock := func(body hcl.Body) (*hcl.Block, hcl.Diagnostics) {
		content, _, diags := body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{{Type: "block"}},
		})
		if diags.HasErrors() {
			return nil, diags
		}
		content, _, diags = content.Blocks[0].Body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{{Type: "nested"}},
		})
		if diags.HasErrors() {
			return nil, diags
		}
		return content.Blocks[0], nil
	}

	tests := []struct {
		name     string
		source   string
		getBlock func(hcl.Body) (*hcl.Block, hcl.Diagnostics)
		want     string
		errCheck func(error) bool
	}{
		{
			name:     "remove inline block",
			source:   `block { foo = 1 }`,
			getBlock: getFirstBlock,
			want:     ``,
			errCheck: neverHappend,
		},
		{
			name: "remove block",
			source: `
block {
  foo = 1
}`,
			getBlock: getFirstBlock,
			want: `
`,
			errCheck: neverHappend,
		},
		{
			name: "remove block with comment",
			source: `
# comment
block {
  foo = 1
}`,
			getBlock: getFirstBlock,
			want: `
`,
			errCheck: neverHappend,
		},
		{
			name: "remove block with multiple comments",
			source: `
# comment
# comment
block {
  foo = 1
}`,
			getBlock: getFirstBlock,
			want: `
`,
			errCheck: neverHappend,
		},
		{
			name: "remove block with multi-line comment",
			source: `
/* comment */
block {
  foo = 1
}`,
			getBlock: getFirstBlock,
			// This is the same behavior as hclwrite.RemoveBlock.
			want: `
/* comment */
`,
			errCheck: neverHappend,
		},
		{
			name: "remove block after attribute",
			source: `
bar = 1
block {
  foo = 1
}`,
			getBlock: getFirstBlock,
			want: `
bar = 1
`,
			errCheck: neverHappend,
		},
		{
			name: "remove block after attribute and newline",
			source: `
bar = 1

block {
  foo = 1
}`,
			getBlock: getFirstBlock,
			want: `
bar = 1

`,
			errCheck: neverHappend,
		},
		{
			name: "remove block after attribute with trailing comment",
			source: `
bar = 1 # comment
block {
  foo = 1
}`,
			getBlock: getFirstBlock,
			want: `
bar = 1 # comment
`,
			errCheck: neverHappend,
		},
		{
			name: "remove inline block after attribute",
			source: `
bar = 1
block { foo = 1 }`,
			getBlock: getFirstBlock,
			want: `
bar = 1
`,
			errCheck: neverHappend,
		},
		{
			name: "remove nested block",
			source: `
block {
  nested {
    foo = 1
  }
}`,
			getBlock: getNestedBlock,
			want: `
block {
}`,
			errCheck: neverHappend,
		},
		{
			name: "remove nested inline block",
			source: `
block {
  nested { foo = 1 }
}`,
			getBlock: getNestedBlock,
			want: `
block {
}`,
			errCheck: neverHappend,
		},
		{
			name: "remove block with traling comment",
			source: `
block {
  foo = 1
} # comment`,
			getBlock: getFirstBlock,
			want: `
`,
			errCheck: neverHappend,
		},
		{
			name: "remove block with next line comment",
			source: `
block {
  foo = 1
}
# comment`,
			getBlock: getFirstBlock,
			want: `
# comment`,
			errCheck: neverHappend,
		},
		{
			name: "remove block in the middle",
			source: `
foo = 1

block {
  foo = 1
}

baz = 1`,
			getBlock: getFirstBlock,
			want: `
foo = 1

baz = 1`,
			errCheck: neverHappend,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file, diags := hclsyntax.ParseConfig([]byte(test.source), "main.tf", hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatalf("failed to parse HCL: %s", diags)
			}
			block, diags := test.getBlock(file.Body)
			if diags.HasErrors() {
				t.Fatalf("failed to get block: %s", diags)
			}

			fixer := NewFixer(map[string][]byte{"main.tf": []byte(test.source)})

			err := fixer.RemoveBlock(block)
			if test.errCheck(err) {
				t.Fatalf("failed to check error: %s", err)
			}

			if diff := cmp.Diff(test.want, string(fixer.changes["main.tf"])); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestRemoveExtBlock(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }
	// getFirstBlock returns the first block in the given body.
	getFirstBlock := func(body hcl.Body) (*hclext.Block, hcl.Diagnostics) {
		content, diags := hclext.PartialContent(body, &hclext.BodySchema{
			Blocks: []hclext.BlockSchema{{Type: "block"}},
		})
		if diags.HasErrors() {
			return nil, diags
		}
		return content.Blocks[0], nil
	}

	tests := []struct {
		name     string
		source   string
		getBlock func(hcl.Body) (*hclext.Block, hcl.Diagnostics)
		want     string
		errCheck func(error) bool
	}{
		{
			name: "remove block",
			source: `
block {
  foo = 1
}`,
			getBlock: getFirstBlock,
			want: `
`,
			errCheck: neverHappend,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file, diags := hclsyntax.ParseConfig([]byte(test.source), "main.tf", hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatalf("failed to parse HCL: %s", diags)
			}
			block, diags := test.getBlock(file.Body)
			if diags.HasErrors() {
				t.Fatalf("failed to get block: %s", diags)
			}

			fixer := NewFixer(map[string][]byte{"main.tf": []byte(test.source)})

			err := fixer.RemoveExtBlock(block)
			if test.errCheck(err) {
				t.Fatalf("failed to check error: %s", err)
			}

			if diff := cmp.Diff(test.want, string(fixer.changes["main.tf"])); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestTextAt(t *testing.T) {
	tests := []struct {
		name string
		src  map[string][]byte
		rng  hcl.Range
		want tflint.TextNode
	}{
		{
			name: "exists",
			src: map[string][]byte{
				"main.tf": []byte(`foo bar baz`),
			},
			rng: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 4}, End: hcl.Pos{Byte: 7}},
			want: tflint.TextNode{
				Bytes: []byte("bar"),
				Range: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 4}, End: hcl.Pos{Byte: 7}},
			},
		},
		{
			name: "does not exists",
			src: map[string][]byte{
				"main.tf": []byte(`foo bar baz`),
			},
			rng: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 14}, End: hcl.Pos{Byte: 17}},
			want: tflint.TextNode{
				Range: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 14}, End: hcl.Pos{Byte: 17}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixer := NewFixer(test.src)
			got := fixer.TextAt(test.rng)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

// @see https://github.com/hashicorp/hcl/blob/v2.17.0/hclwrite/generate_test.go#L18
func TestValueText(t *testing.T) {
	tests := []struct {
		value cty.Value
		want  string
	}{
		{
			value: cty.NullVal(cty.DynamicPseudoType),
			want:  "null",
		},
		{
			value: cty.True,
			want:  "true",
		},
		{
			value: cty.False,
			want:  "false",
		},
		{
			value: cty.NumberIntVal(0),
			want:  "0",
		},
		{
			value: cty.NumberFloatVal(0.5),
			want:  "0.5",
		},
		{
			value: cty.NumberVal(big.NewFloat(0).SetPrec(512).Mul(big.NewFloat(40000000), big.NewFloat(2000000))),
			want:  "80000000000000",
		},
		{
			value: cty.StringVal(""),
			want:  `""`,
		},
		{
			value: cty.StringVal("foo"),
			want:  `"foo"`,
		},
		{
			value: cty.StringVal(`"foo"`),
			want:  `"\"foo\""`,
		},
		{
			value: cty.StringVal("hello\nworld\n"),
			want:  `"hello\nworld\n"`,
		},
		{
			value: cty.StringVal("hello\r\nworld\r\n"),
			want:  `"hello\r\nworld\r\n"`,
		},
		{
			value: cty.StringVal(`what\what`),
			want:  `"what\\what"`,
		},
		{
			value: cty.StringVal("𝄞"),
			want:  `"𝄞"`,
		},
		{
			value: cty.StringVal("👩🏾"),
			want:  `"👩🏾"`,
		},
		{
			value: cty.EmptyTupleVal,
			want:  "[]",
		},
		{
			value: cty.TupleVal([]cty.Value{cty.EmptyTupleVal}),
			want:  "[[]]",
		},
		{
			value: cty.ListValEmpty(cty.String),
			want:  "[]",
		},
		{
			value: cty.SetValEmpty(cty.Bool),
			want:  "[]",
		},
		{
			value: cty.TupleVal([]cty.Value{cty.True}),
			want:  "[true]",
		},
		{
			value: cty.TupleVal([]cty.Value{cty.True, cty.NumberIntVal(0)}),
			want:  "[true, 0]",
		},
		{
			value: cty.EmptyObjectVal,
			want:  "{}",
		},
		{
			value: cty.MapValEmpty(cty.Bool),
			want:  "{}",
		},
		{
			value: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.True,
			}),
			want: "{ foo = true }",
		},
		{
			value: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.True,
				"bar": cty.NumberIntVal(0),
			}),
			want: "{ bar = 0, foo = true }",
		},
		{
			value: cty.ObjectVal(map[string]cty.Value{
				"foo bar": cty.True,
			}),
			want: `{ "foo bar" = true }`,
		},
	}

	for _, test := range tests {
		t.Run(test.value.GoString(), func(t *testing.T) {
			fixer := NewFixer(nil)
			got := fixer.ValueText(test.value)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestRangeTo(t *testing.T) {
	start := hcl.Pos{Byte: 10, Line: 2, Column: 1}

	tests := []struct {
		name string
		to   string
		want hcl.Range
	}{
		{
			name: "empty",
			to:   "",
			want: hcl.Range{Start: start, End: start},
		},
		{
			name: "single line",
			to:   "foo",
			want: hcl.Range{Start: start, End: hcl.Pos{Byte: 13, Line: 2, Column: 4}},
		},
		{
			name: "trailing new line",
			to:   "foo\n",
			want: hcl.Range{Start: start, End: hcl.Pos{Byte: 13, Line: 2, Column: 4}},
		},
		{
			name: "multi new line",
			to:   "foo\nbar",
			want: hcl.Range{Start: start, End: hcl.Pos{Byte: 17, Line: 3, Column: 4}},
		},
		{
			name: "multibytes",
			to:   "こんにちは世界",
			want: hcl.Range{Start: start, End: hcl.Pos{Byte: 31, Line: 2, Column: 8}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixer := NewFixer(nil)

			got := fixer.RangeTo(test.to, "", start)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestChanges(t *testing.T) {
	src := map[string][]byte{
		"main.tf": []byte(`
foo = 1
  bar = 2
`),
		"main.tf.json": []byte(`{"foo": 1, "bar": 2}`),
	}
	fixer := NewFixer(src)

	if len(fixer.Changes()) != 0 {
		t.Errorf("unexpected changes: %#v", fixer.Changes())
	}
	if fixer.HasChanges() {
		t.Errorf("unexpected changes: %#v", fixer.Changes())
	}
	if len(fixer.shifts) != 0 {
		t.Errorf("unexpected shifts: %#v", fixer.shifts)
	}

	// Make changes
	if err := fixer.ReplaceText(
		hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 11}, End: hcl.Pos{Byte: 14}},
		"barbaz",
	); err != nil {
		t.Fatal(err)
	}
	if err := fixer.ReplaceText(
		hcl.Range{Filename: "main.tf.json", Start: hcl.Pos{Byte: 12}, End: hcl.Pos{Byte: 15}},
		"barbaz",
	); err != nil {
		t.Fatal(err)
	}

	changed := map[string][]byte{
		"main.tf": []byte(`
foo = 1
  barbaz = 2
`),
		"main.tf.json": []byte(`{"foo": 1, "barbaz": 2}`),
	}

	if diff := cmp.Diff(src, fixer.sources); diff != "" {
		t.Error(diff)
	}
	if diff := cmp.Diff(fixer.Changes(), changed); diff != "" {
		t.Error(diff)
	}
	if !fixer.HasChanges() {
		t.Errorf("unexpected changes: %#v", fixer.Changes())
	}
	if len(fixer.shifts) != 2 {
		t.Errorf("unexpected shifts: %#v", fixer.shifts)
	}

	// Format changes
	fixer.FormatChanges()

	fixed := map[string][]byte{
		"main.tf": []byte(`
foo    = 1
barbaz = 2
`),
		"main.tf.json": []byte(`{"foo": 1, "barbaz": 2}`),
	}

	if diff := cmp.Diff(fixer.Changes(), fixed); diff != "" {
		t.Error(diff)
	}
	if len(fixer.shifts) != 2 {
		t.Errorf("unexpected shifts: %#v", fixer.shifts)
	}

	// Apply changes
	fixer.ApplyChanges()

	if diff := cmp.Diff(fixed, fixer.sources); diff != "" {
		t.Error(diff)
	}
	if len(fixer.Changes()) != 0 {
		t.Errorf("unexpected changes: %#v", fixer.Changes())
	}
	if fixer.HasChanges() {
		t.Errorf("unexpected changes: %#v", fixer.Changes())
	}
	if len(fixer.shifts) != 0 {
		t.Errorf("unexpected shifts: %#v", fixer.shifts)
	}
}

func TestStashChanges(t *testing.T) {
	tests := []struct {
		name   string
		source string
		fix    func(*Fixer) error
		want   string
		shifts int
	}{
		{
			name:   "no changes",
			source: `foo`,
			fix: func(fixer *Fixer) error {
				fixer.StashChanges()
				fixer.PopChangesFromStash()
				return nil
			},
			want:   "",
			shifts: 0,
		},
		{
			name:   "changes after stash",
			source: `foo`,
			fix: func(fixer *Fixer) error {
				fixer.StashChanges()
				if err := fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 0}, End: hcl.Pos{Byte: 0}}, "bar"); err != nil {
					return err
				}
				fixer.PopChangesFromStash()
				return nil
			},
			want:   "",
			shifts: 0,
		},
		{
			name:   "stash after changes",
			source: `foo`,
			fix: func(fixer *Fixer) error {
				if err := fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 0}, End: hcl.Pos{Byte: 0}}, "bar"); err != nil {
					return err
				}
				fixer.StashChanges()
				if err := fixer.ReplaceText(hcl.Range{Filename: "main.tf", Start: hcl.Pos{Byte: 0}, End: hcl.Pos{Byte: 0}}, "baz"); err != nil {
					return err
				}
				fixer.PopChangesFromStash()
				return nil
			},
			want:   "barfoo",
			shifts: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixer := NewFixer(map[string][]byte{"main.tf": []byte(test.source)})
			if err := test.fix(fixer); err != nil {
				t.Fatalf("failed to fix: %s", err)
			}

			if diff := cmp.Diff(test.want, string(fixer.changes["main.tf"])); diff != "" {
				t.Error(diff)
			}
			if test.shifts != len(fixer.shifts) {
				t.Errorf("shifts: want %d, got %d", test.shifts, len(fixer.shifts))
			}
		})
	}
}
