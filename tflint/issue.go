package tflint

import (
	"encoding/json"
	"fmt"
	"sort"

	hcl "github.com/hashicorp/hcl/v2"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// Issue represents a problem in configurations
type Issue struct {
	Rule    Rule
	Message string
	Range   hcl.Range
	Fixable bool
	Callers []hcl.Range

	// Source is the source code of the file where the issue was found.
	// Usually this is the same as the originally loaded source,
	// but it may be a different if rewritten by autofixes.
	Source []byte
}

// Issues is an alias for the map of Issue
type Issues []*Issue

// Severity indicates the severity of the issue
type Severity = sdk.Severity

// Creates a new severity from a string
func NewSeverity(s string) (Severity, error) {
	switch s {
	case "error":
		return sdk.ERROR, nil
	case "warning":
		return sdk.WARNING, nil
	case "notice":
		return sdk.NOTICE, nil
	default:
		return sdk.NOTICE, fmt.Errorf("%s is not a recognized severity", s)
	}
}

// Converts a severity into an ascending int32
func SeverityToInt32(s Severity) (int32, error) {
	switch s {
	case sdk.ERROR:
		return 2, nil
	case sdk.WARNING:
		return 1, nil
	case sdk.NOTICE:
		return 0, nil
	default:
		return 0, fmt.Errorf("%s is not a recognized severity", s)
	}
}

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
		return issues[i].Message < issues[j].Message
	})
	return issues
}

type issue struct {
	Rule    *rule       `json:"rule"`
	Message string      `json:"message"`
	Range   hcl.Range   `json:"range"`
	Fixable bool        `json:"fixable"`
	Callers []hcl.Range `json:"callers"`
	Source  []byte      `json:"source"`
}

type rule struct {
	RawName     string   `json:"name"`
	RawSeverity Severity `json:"severity"`
	RawLink     string   `json:"link"`
}

var _ Rule = (*rule)(nil)

func (r *rule) Name() string       { return r.RawName }
func (r *rule) Severity() Severity { return r.RawSeverity }
func (r *rule) Link() string       { return r.RawLink }

func (i *Issue) MarshalJSON() ([]byte, error) {
	return json.Marshal(issue{
		Rule: &rule{
			RawName:     i.Rule.Name(),
			RawSeverity: i.Rule.Severity(),
			RawLink:     i.Rule.Link(),
		},
		Message: i.Message,
		Range:   i.Range,
		Fixable: i.Fixable,
		Callers: i.Callers,
		Source:  i.Source,
	})
}

func (i *Issue) UnmarshalJSON(data []byte) error {
	var out issue
	err := json.Unmarshal(data, &out)
	if err != nil {
		return err
	}

	i.Rule = out.Rule
	i.Message = out.Message
	i.Range = out.Range
	i.Fixable = out.Fixable
	i.Callers = out.Callers
	i.Source = out.Source

	return nil
}
