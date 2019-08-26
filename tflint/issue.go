package tflint

import (
	"encoding/json"
	"sort"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
)

// Issue represents a problem in configurations
type Issue struct {
	Rule    Rule
	Message string
	Range   hcl.Range
	Callers []hcl.Range
}

// Issues is an alias for the map of Issue
type Issues []*Issue

// Sort returns the sorted receiver
func (issues Issues) Sort() Issues {
	sort.Slice(issues, func(i, j int) bool {
		iRange := issues[i].Range
		jRange := issues[j].Range
		if iRange.Filename != jRange.Filename {
			return iRange.Filename < jRange.Filename
		}
		if iRange.Start.Line != jRange.Start.Line {
			return iRange.Start.Line < jRange.Start.Line
		}
		if iRange.Start.Column != jRange.Start.Column {
			return iRange.Start.Column < jRange.Start.Column
		}
		if iRange.End.Line != jRange.End.Line {
			return iRange.End.Line > jRange.End.Line
		}
		if iRange.End.Column != jRange.End.Column {
			return iRange.End.Column > jRange.End.Column
		}
		return issues[i].Rule.Name() < issues[j].Rule.Name()
	})
	return issues
}

// MarshalJSON is a method reserved for conversion to JSON
func (i *Issue) MarshalJSON() ([]byte, error) {
	// Keep JSON structure for the backward compatibility
	issue := &issue.Issue{
		Detector: i.Rule.Name(),
		Type:     i.Rule.Type(),
		Message:  i.Message,
		Line:     i.Range.Start.Line,
		File:     i.Range.Filename,
		Link:     i.Rule.Link(),
	}
	return json.Marshal(issue)
}
