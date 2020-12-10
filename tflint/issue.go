package tflint

import (
	"sort"

	hcl "github.com/hashicorp/hcl/v2"
)

// Issue represents a problem in configurations
type Issue struct {
	Rule     Rule
	Message  string
	Severity string
	Range    hcl.Range
	Callers  []hcl.Range
}

// Issues is an alias for the map of Issue
type Issues []*Issue

const (
	// ERROR is possible errors
	ERROR = "Error"
	// WARNING doesn't cause problem immediately, but not good
	WARNING = "Warning"
	// NOTICE is not important, it's mentioned
	NOTICE = "Notice"
)

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
