package tflint

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

var annotationPattern = regexp.MustCompile(`tflint-ignore: ([^\n*/#]+)`)

// Annotation represents comments with special meaning in TFLint
type Annotation struct {
	Content string
	Token   hclsyntax.Token
}

// Annotations is slice of Annotation
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

		match := annotationPattern.FindStringSubmatch(string(token.Bytes))
		if len(match) != 2 {
			continue
		}
		ret = append(ret, Annotation{
			Content: strings.TrimSpace(match[1]),
			Token:   token,
		})
	}

	return ret, diags
}

// IsAffected checks if the passed issue is affected with the annotation
func (a *Annotation) IsAffected(issue *Issue) bool {
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
func (a *Annotation) String() string {
	return fmt.Sprintf("annotation:%s (%s)", a.Content, a.Token.Range.String())
}
