package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

// Fixer is a tool to rewrite HCL source code.
type Fixer struct {
	sources map[string][]byte
	changes map[string][]byte
	shifts  []shift

	stashedChanges map[string][]byte
	stashedShifts  []shift
}

type shift struct {
	target hcl.Range // rewrite target range caused by the shift
	start  int       // start byte index of the shift
	offset int       // shift offset
}

// NewFixer creates a new Fixer instance.
func NewFixer(sources map[string][]byte) *Fixer {
	return &Fixer{
		sources: sources,
		changes: map[string][]byte{},
		shifts:  []shift{},

		stashedChanges: map[string][]byte{},
		stashedShifts:  []shift{},
	}
}

// ReplaceText rewrites the given range of source code to a new text.
// If the range is overlapped with a previous rewrite range, it returns an error.
//
// Either string or tflint.TextNode is valid as an argument.
// TextNode can be obtained with fixer.TextAt(range).
// If the argument is a TextNode, and the range is contained in the replacement range,
// this function automatically minimizes the replacement range as much as possible.
//
// For example, if the source code is "(foo)", ReplaceText(range, "[foo]")
// rewrites the whole "(foo)". But ReplaceText(range, "[", TextAt(fooRange), "]")
// rewrites only "(" and ")". This is useful to avoid unintended conflicts.
func (f *Fixer) ReplaceText(rng hcl.Range, texts ...any) error {
	if len(texts) == 0 {
		return fmt.Errorf("no text to replace")
	}

	var start hcl.Pos = rng.Start
	var new string

	for _, text := range texts {
		switch text := text.(type) {
		case string:
			new += text
		case tflint.TextNode:
			if rng.Filename == text.Range.Filename && start.Byte <= text.Range.Start.Byte {
				if err := f.replaceText(hcl.Range{Filename: rng.Filename, Start: start, End: text.Range.Start}, new); err != nil {
					return err
				}
				start = text.Range.End
				new = ""
			} else {
				// If the text node is not contained in the replacement range, just append the text.
				new += string(text.Bytes)
			}
		default:
			return fmt.Errorf("ReplaceText only accepts string or textNode, but got %T", text)
		}
	}
	return f.replaceText(hcl.Range{Filename: rng.Filename, Start: start, End: rng.End}, new)
}

func (f *Fixer) replaceText(rng hcl.Range, new string) error {
	// If there are already changes, overwrite the changed content.
	var file []byte
	if change, exists := f.changes[rng.Filename]; exists {
		file = change
	} else if source, exists := f.sources[rng.Filename]; exists {
		file = source
	} else {
		return fmt.Errorf("file not found: %s", rng.Filename)
	}

	// Apply rewrite gaps so that you can chain rewrites using pre-change ranges.
	for _, shift := range f.shifts {
		if shift.target.Filename != rng.Filename {
			continue
		}
		if !shift.target.Overlap(rng).Empty() {
			// If the range is the same as before, just update the content.
			// Note that only the end byte index should reflect the shift.
			if shift.target.Start.Byte == rng.Start.Byte && shift.target.End.Byte == rng.End.Byte {
				rng.End.Byte += shift.offset
				continue
			}
			return fmt.Errorf("range overlaps with a previous rewrite range: %s", shift.target.String())
		}
		// Apply shift to the range if the shift is before the range.
		if shift.start <= rng.Start.Byte {
			rng.Start.Byte += shift.offset
			rng.End.Byte += shift.offset
		}
	}

	buf := bytes.NewBuffer(bytes.Clone(file[:rng.Start.Byte]))
	buf.WriteString(new)
	buf.Write(file[rng.End.Byte:])

	// If the new content is the same as the before, do nothing.
	if bytes.Equal(file, buf.Bytes()) {
		return nil
	}

	// Tracks rewrite gaps
	oldBytes := rng.End.Byte - rng.Start.Byte
	newBytes := len(new)
	if oldBytes == newBytes {
		// no shift: foo -> bar
	} else if oldBytes < newBytes {
		// shift right: foo -> foooo
		//                        |-| shift
		f.shifts = append(f.shifts, shift{
			target: rng,
			start:  rng.End.Byte,
			offset: newBytes - oldBytes,
		})
	} else {
		// shift left: foooo -> foo
		//                         |-| shift
		f.shifts = append(f.shifts, shift{
			target: rng,
			start:  rng.End.Byte - (oldBytes - newBytes),
			offset: -(oldBytes - newBytes),
		})
	}

	f.changes[rng.Filename] = buf.Bytes()
	return nil
}

// InsertTextBefore inserts the given text before the given range.
func (f *Fixer) InsertTextBefore(rng hcl.Range, text string) error {
	return f.ReplaceText(hcl.Range{Filename: rng.Filename, Start: rng.Start, End: rng.Start}, text)
}

// InsertTextAfter inserts the given text after the given range.
func (f *Fixer) InsertTextAfter(rng hcl.Range, text string) error {
	return f.ReplaceText(hcl.Range{Filename: rng.Filename, Start: rng.End, End: rng.End}, text)
}

// Remove removes the given range of source code.
func (f *Fixer) Remove(rng hcl.Range) error {
	return f.ReplaceText(rng, "")
}

// RemoveAttribute removes the given attribute from the source code.
// The difference from Remove is that it removes the attribute
// and the associated newlines, indentations, and comments.
// This only works for HCL native syntax. JSON syntax is not supported
// and returns tflint.ErrFixNotSupported.
func (f *Fixer) RemoveAttribute(attr *hcl.Attribute) error {
	if terraform.IsJSONFilename(attr.Range.Filename) {
		return tflint.ErrFixNotSupported
	}

	rng, err := f.expandRangeToTrivialTokens(attr.Range)
	if err != nil {
		return err
	}
	return f.Remove(rng)
}

// RemoveBlock removes the given block from the source code.
// The difference from Remove is that it removes the block
// and the associated newlines, indentations, and comments.
// This only works for HCL native syntax. JSON syntax is not supported
// and returns tflint.ErrFixNotSupported.
func (f *Fixer) RemoveBlock(block *hcl.Block) error {
	if terraform.IsJSONFilename(block.DefRange.Filename) {
		return tflint.ErrFixNotSupported
	}

	source, exists := f.sources[block.DefRange.Filename]
	if !exists {
		return fmt.Errorf("file not found: %s", block.DefRange.Filename)
	}
	// Parse the source code to get the whole block range.
	// Notice that hcl.Block does not have the whole range, but hclsyntax.Block does.
	file, diags := hclsyntax.ParseConfig(source, block.DefRange.Filename, hcl.InitialPos)
	if diags.HasErrors() {
		return diags
	}

	var blockRange hcl.Range
	diags = hclsyntax.VisitAll(file.Body.(*hclsyntax.Body), func(node hclsyntax.Node) hcl.Diagnostics {
		if nativeBlock, ok := node.(*hclsyntax.Block); ok {
			if nativeBlock.TypeRange.Start.Byte == block.TypeRange.Start.Byte {
				blockRange = hcl.RangeBetween(block.DefRange, nativeBlock.CloseBraceRange)
				return nil
			}
		}
		return nil
	})
	if diags.HasErrors() {
		return diags
	}
	if blockRange.Empty() {
		return fmt.Errorf("block not found at %s:%d,%d", block.DefRange.Filename, block.DefRange.Start.Line, block.DefRange.Start.Column)
	}

	rng, err := f.expandRangeToTrivialTokens(blockRange)
	if err != nil {
		return err
	}

	return f.Remove(rng)
}

// RemoveExtBlock removes the given block from the source code.
// This is similar to RemoveBlock, but it works for *hclext.Block.
func (f *Fixer) RemoveExtBlock(block *hclext.Block) error {
	// In RemoveBlock, body is not important, so convert the given block
	// to a native block without the body.
	return f.RemoveBlock(&hcl.Block{
		Type:   block.Type,
		Labels: block.Labels,

		DefRange:    block.DefRange,
		TypeRange:   block.TypeRange,
		LabelRanges: block.LabelRanges,
	})
}

// expandRangeToTrivialTokens expands the given range to include comments, newlines, and indentations.
func (f *Fixer) expandRangeToTrivialTokens(rng hcl.Range) (hcl.Range, error) {
	source, exists := f.sources[rng.Filename]
	if !exists {
		return rng, fmt.Errorf("file not found: %s", rng.Filename)
	}
	// Use tokenScanner to find tokens before and after the attribute/block range,
	// in order to remove comments, newlines, and indentations.
	scanner, diags := newTokenScanner(source, rng.Filename)
	if diags.HasErrors() {
		return rng, diags
	}

	var expanded = rng

	// Scan backward until a newline is found, and expand the start position.
	//
	//   <-- start
	//         |
	//         foo = 1
	if err := scanner.seek(rng.Start, tokenStart); err != nil {
		return rng, err
	}
endScanBackward:
	for scanner.scanBackward() {
		switch scanner.token().Type {
		case hclsyntax.TokenNewline:
			// Seek to the end of the token to keep the newline.
			scanner.seekTokenEnd()
			break endScanBackward

		case hclsyntax.TokenComment:
			// For a trailing single-line comment, determines whether the comment is associated with itself.
			// For example, the following comment is associated with the "foo" attribute and should be removed.
			//
			// # comment
			// foo = 1
			//
			// On the other hand, the following comment is associated with the "bar" attribute and should not be removed.
			//
			// bar = 2 # comment
			// foo = 1
			//
			// To determine these, we need to look at the tokens before the comment token.
			if strings.HasPrefix(string(scanner.token().Bytes), "#") || strings.HasPrefix(string(scanner.token().Bytes), "//") {
				trailingCommentIndex := scanner.index

				for scanner.scanBackward() {
					switch scanner.token().Type {
					case hclsyntax.TokenComment:
						// Ignore comment tokens in case there are multiple comments.
						//
						// # comment1
						// # comment2
						// foo = 1
						continue

					case hclsyntax.TokenNewline:
						// If there is only a comment after the newline, the line can be deleted.
						scanner.seekTokenEnd()
						break endScanBackward

					default:
						// If there is a token other than comment or newline, seek to the ending position of the trailing comment.
						if err := scanner.seekByIndex(trailingCommentIndex, tokenEnd); err != nil {
							return rng, err
						}
						break endScanBackward
					}
				}
			}

		// For an inline block, use an opening brace instead.
		//
		// block { foo = 1 }   => TokenOBrace + Attribute + TokenCBrace
		case hclsyntax.TokenOBrace:
			// Seek to the end of the token to keep the brace.
			scanner.seekTokenEnd()
			break endScanBackward
		}
	}
	expanded.Start = scanner.pos

	// Count the number of newlines before the range.
	// This is because it doesn't leave a nonsense newline after deletion
	newlineCountInBackward := 0
	for scanner.scanBackwardIf(hclsyntax.TokenNewline) {
		newlineCountInBackward++
	}

	// Scan forward until a newline is found, and expand the end position.
	//
	//              end -->
	//               |
	//         foo = 1
	if err := scanner.seek(rng.End, tokenEnd); err != nil {
		return rng, err
	}
endScan:
	for scanner.scan() {
		switch scanner.token().Type {
		case hclsyntax.TokenNewline:
			// Remove newline
			break endScan

		case hclsyntax.TokenComment:
			// For a trailing single-line comment, use a comment token instead because it does not produce a newline token.
			//
			// foo = 1                 => Attribute + TokenNewline
			// foo = 1 # comment       => Attribute + TokenComment
			// foo = 1 /* comment */   => Attribute + TokenComment + TokenNewline
			if strings.HasPrefix(string(scanner.token().Bytes), "#") || strings.HasPrefix(string(scanner.token().Bytes), "//") {
				break endScan
			}

		// For an inline block, use an closing brace instead.
		//
		// block { foo = 1 }   => TokenOBrace + Attribute + TokenCBrace
		case hclsyntax.TokenCBrace:
			// Seek to the start of the token to keep the brace.
			scanner.seekTokenStart()
			break endScan
		}
	}
	expanded.End = scanner.pos

	// Count the number of newlines after the range.
	newlineCountInForward := 0
	for scanner.scanIf(hclsyntax.TokenNewline) {
		newlineCountInForward++
	}
	// If the number of newlines before and after the range is the same,
	// expand the end position to delete nonsense newlines.
	//
	// foo = 1
	//
	// bar = 2   <-- delete this attribute
	//
	// baz = 3
	//
	// Newlines are removed like this:
	//
	// foo = 1
	//
	// baz = 3
	//
	if newlineCountInForward > 0 && newlineCountInBackward == newlineCountInForward {
		expanded.End = scanner.pos
	}

	return expanded, nil
}

// TextAt returns a text node at the given range.
// This is expected to be passed as an argument to ReplaceText.
// Note this doesn't take into account the changes made by the fixer in a rule.
func (f *Fixer) TextAt(rng hcl.Range) tflint.TextNode {
	source := f.sources[rng.Filename]
	if !rng.CanSliceBytes(source) {
		return tflint.TextNode{Range: rng}
	}
	return tflint.TextNode{Bytes: rng.SliceBytes(source), Range: rng}
}

// ValueText returns a text representation of the given cty.Value.
// Values are always converted to a single line. For more pretty-printing,
// implement your own conversion function.
//
// This function is inspired by hclwrite.TokensForValue.
// https://github.com/hashicorp/hcl/blob/v2.16.2/hclwrite/generate.go#L26
func (f *Fixer) ValueText(val cty.Value) string {
	switch {
	case !val.IsKnown():
		panic("cannot produce text for unknown value")

	case val.IsNull():
		return "null"

	case val.Type() == cty.Bool:
		if val.True() {
			return "true"
		}
		return "false"

	case val.Type() == cty.Number:
		return val.AsBigFloat().Text('f', -1)

	case val.Type() == cty.String:
		return fmt.Sprintf(`"%s"`, escapeQuotedStringLit(val.AsString()))

	case val.Type().IsListType() || val.Type().IsSetType() || val.Type().IsTupleType():
		items := make([]string, 0, val.LengthInt())
		for it := val.ElementIterator(); it.Next(); {
			_, v := it.Element()
			items = append(items, f.ValueText(v))
		}
		return fmt.Sprintf("[%s]", strings.Join(items, ", "))

	case val.Type().IsMapType() || val.Type().IsObjectType():
		if val.LengthInt() == 0 {
			return "{}"
		}
		items := make([]string, 0, val.LengthInt())
		for it := val.ElementIterator(); it.Next(); {
			k, v := it.Element()
			if hclsyntax.ValidIdentifier(k.AsString()) {
				items = append(items, fmt.Sprintf("%s = %s", k.AsString(), f.ValueText(v)))
			} else {
				items = append(items, fmt.Sprintf("%s = %s", f.ValueText(k), f.ValueText(v)))
			}
		}
		return fmt.Sprintf("{ %s }", strings.Join(items, ", "))

	default:
		panic(fmt.Sprintf("cannot produce text for %s", val.Type().FriendlyName()))
	}
}

func escapeQuotedStringLit(s string) []byte {
	if len(s) == 0 {
		return nil
	}
	buf := make([]byte, 0, len(s))
	for i, r := range s {
		switch r {
		case '\n':
			buf = append(buf, '\\', 'n')
		case '\r':
			buf = append(buf, '\\', 'r')
		case '\t':
			buf = append(buf, '\\', 't')
		case '"':
			buf = append(buf, '\\', '"')
		case '\\':
			buf = append(buf, '\\', '\\')
		case '$', '%':
			buf = appendRune(buf, r)
			remain := s[i+1:]
			if len(remain) > 0 && remain[0] == '{' {
				// Double up our template introducer symbol to escape it.
				buf = appendRune(buf, r)
			}
		default:
			if !unicode.IsPrint(r) {
				var fmted string
				if r < 65536 {
					fmted = fmt.Sprintf("\\u%04x", r)
				} else {
					fmted = fmt.Sprintf("\\U%08x", r)
				}
				buf = append(buf, fmted...)
			} else {
				buf = appendRune(buf, r)
			}
		}
	}
	return buf
}

func appendRune(b []byte, r rune) []byte {
	l := utf8.RuneLen(r)
	for i := 0; i < l; i++ {
		b = append(b, 0) // make room at the end of our buffer
	}
	ch := b[len(b)-l:]
	utf8.EncodeRune(ch, r)
	return b
}

// RangeTo returns a range from the given start position to the given text.
// Note that it doesn't check if the text is actually in the range.
func (f *Fixer) RangeTo(to string, filename string, start hcl.Pos) hcl.Range {
	end := start
	if to == "" {
		return hcl.Range{Filename: filename, Start: start, End: end}
	}

	scanner := hcl.NewRangeScanner([]byte(to), filename, bufio.ScanLines)
	for scanner.Scan() {
		end = scanner.Range().End
	}
	if scanner.Err() != nil {
		// never happen
		panic(scanner.Err())
	}

	var line, column, bytes int
	line = start.Line + end.Line - 1
	if end.Line == 1 {
		column = start.Column + end.Column - 1
	} else {
		column = end.Column
	}
	bytes = start.Byte + end.Byte

	return hcl.Range{
		Filename: filename,
		Start:    start,
		End:      hcl.Pos{Line: line, Column: column, Byte: bytes},
	}
}

// Changes returns the changes made by the fixer.
// Note this API is not intended to be used by plugins.
func (f *Fixer) Changes() map[string][]byte {
	return f.changes
}

// HasChanges returns true if the fixer has changes.
// Note this API is not intended to be used by plugins.
func (f *Fixer) HasChanges() bool {
	return len(f.changes) > 0
}

// FormatChanges formats the changes made by the fixer.
// Note this API is not intended to be used by plugins.
func (f *Fixer) FormatChanges() {
	for filename, content := range f.changes {
		if terraform.IsJSONFilename(filename) {
			continue
		}
		f.changes[filename] = hclwrite.Format(content)
	}
}

// ApplyChanges applies the changes made by the fixer.
// Note this API is not intended to be used by plugins.
func (f *Fixer) ApplyChanges() {
	for filename, content := range f.changes {
		f.sources[filename] = content
	}
	f.changes = map[string][]byte{}
	f.shifts = []shift{}
}

// StashChanges stashes the current changes.
// Note this API is not intended to be used by plugins.
func (f *Fixer) StashChanges() {
	f.stashedChanges = map[string][]byte{}
	for k, v := range f.changes {
		f.stashedChanges[k] = v
	}
	f.stashedShifts = make([]shift, len(f.shifts))
	copy(f.stashedShifts, f.shifts)
}

// PopChangesFromStash pops changes from the stash.
// Note this API is not intended to be used by plugins.
func (f *Fixer) PopChangesFromStash() {
	f.changes = map[string][]byte{}
	for k, v := range f.stashedChanges {
		f.changes[k] = v
	}
	f.shifts = make([]shift, len(f.stashedShifts))
	copy(f.shifts, f.stashedShifts)
}
