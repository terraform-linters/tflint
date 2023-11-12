package tflint

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// Annotation represents comments with special meaning in TFLint
type Annotation interface {
	IsAffected(*Issue) bool
	String() string
}

// Annotations is a slice of Annotation
type Annotations []Annotation

// NewAnnotations find annotations from the passed tokens and return that list.
func NewAnnotations(path string, file *hcl.File) (Annotations, hcl.Diagnostics) {
	ret := Annotations{}

	tokens, diags := hclsyntax.LexConfig(file.Bytes, path, hcl.Pos{Byte: 0, Line: 1, Column: 1})
	if diags.HasErrors() {
		return ret, diags
	}

	for _, token := range tokens {
		if token.Type != hclsyntax.TokenComment {
			continue
		}

		// tflint-ignore annotation
		match := lineAnnotationPattern.FindStringSubmatch(string(token.Bytes))
		if len(match) == 2 {
			ret = append(ret, &LineAnnotation{
				Content: strings.TrimSpace(match[1]),
				Token:   token,
			})
			continue
		}

		// tflint-ignore-file annotation
		match = fileAnnotationPattern.FindStringSubmatch(string(token.Bytes))
		if len(match) == 2 {
			if !(token.Range.Start.Line == 1 && token.Range.Start.Column == 1) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "tflint-ignore-file annotation must be written at the top of file",
					Detail:   fmt.Sprintf("tflint-ignore-file annotation is written at line %d, column %d", token.Range.Start.Line, token.Range.Start.Column),
					Subject:  token.Range.Ptr(),
				})
				continue
			}
			ret = append(ret, &FileAnnotation{
				Content: strings.TrimSpace(match[1]),
				Token:   token,
			})
			continue
		}
	}

	return ret, diags
}

var lineAnnotationPattern = regexp.MustCompile(`tflint-ignore: ([^\n*/#]+)`)

// LineAnnotation is an annotation for ignoring issues in a line
type LineAnnotation struct {
	Content string
	Token   hclsyntax.Token
}

// IsAffected checks if the passed issue is affected with the annotation
func (a *LineAnnotation) IsAffected(issue *Issue) bool {
	if a.Token.Range.Filename != issue.Range.Filename {
		return false
	}

	rules := strings.Split(a.Content, ",")
	for i, rule := range rules {
		rules[i] = strings.TrimSpace(rule)
	}

	if slices.Contains(rules, issue.Rule.Name()) || slices.Contains(rules, "all") {
		if a.Token.Range.Start.Line == issue.Range.Start.Line {
			return true
		}
		if a.Token.Range.Start.Line == issue.Range.Start.Line-1 {
			return true
		}
	}
	return false
}

// String returns the string representation of the annotation
func (a *LineAnnotation) String() string {
	return fmt.Sprintf("tflint-ignore: %s (%s)", a.Content, a.Token.Range.String())
}

var fileAnnotationPattern = regexp.MustCompile(`tflint-ignore-file: ([^\n*/#]+)`)

// FileAnnotation is an annotation for ignoring issues in a file
type FileAnnotation struct {
	Content string
	Token   hclsyntax.Token
}

// IsAffected checks if the passed issue is affected with the annotation
func (a *FileAnnotation) IsAffected(issue *Issue) bool {
	if a.Token.Range.Filename != issue.Range.Filename {
		return false
	}

	rules := strings.Split(a.Content, ",")
	for i, rule := range rules {
		rules[i] = strings.TrimSpace(rule)
	}

	if slices.Contains(rules, issue.Rule.Name()) || slices.Contains(rules, "all") {
		return true
	}
	return false
}

// String returns the string representation of the annotation
func (a *FileAnnotation) String() string {
	return fmt.Sprintf("tflint-ignore-file: %s (%s)", a.Content, a.Token.Range.String())
}
