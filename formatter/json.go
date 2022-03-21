package formatter

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// JSONIssue is a temporary structure for converting TFLint issues to JSON.
type JSONIssue struct {
	Rule    JSONRule    `json:"rule"`
	Message string      `json:"message"`
	Range   JSONRange   `json:"range"`
	Callers []JSONRange `json:"callers"`
}

// JSONRule is a temporary structure for converting TFLint rules to JSON.
type JSONRule struct {
	Name     string `json:"name"`
	Severity string `json:"severity"`
	Link     string `json:"link"`
}

// JSONRange is a temporary structure for converting ranges to JSON.
type JSONRange struct {
	Filename string  `json:"filename"`
	Start    JSONPos `json:"start"`
	End      JSONPos `json:"end"`
}

// JSONPos is a temporary structure for converting positions to JSON.
type JSONPos struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// JSONError is a temporary structure for converting errors to JSON.
type JSONError struct {
	Summary  string     `json:"summary,omitempty"`
	Message  string     `json:"message"`
	Severity string     `json:"severity"`
	Range    *JSONRange `json:"range,omitempty"` // pointer so omitempty works
}

// JSONOutput is a temporary structure for converting to JSON.
type JSONOutput struct {
	Issues []JSONIssue `json:"issues"`
	Errors []JSONError `json:"errors"`
}

func (f *Formatter) jsonPrint(issues tflint.Issues, appErr error) {
	ret := &JSONOutput{Issues: make([]JSONIssue, len(issues)), Errors: []JSONError{}}

	for idx, issue := range issues.Sort() {
		ret.Issues[idx] = JSONIssue{
			Rule: JSONRule{
				Name:     issue.Rule.Name(),
				Severity: toSeverity(issue.Rule.Severity()),
				Link:     issue.Rule.Link(),
			},
			Message: issue.Message,
			Range: JSONRange{
				Filename: issue.Range.Filename,
				Start:    JSONPos{Line: issue.Range.Start.Line, Column: issue.Range.Start.Column},
				End:      JSONPos{Line: issue.Range.End.Line, Column: issue.Range.End.Column},
			},
			Callers: make([]JSONRange, len(issue.Callers)),
		}
		for i, caller := range issue.Callers {
			ret.Issues[idx].Callers[i] = JSONRange{
				Filename: caller.Filename,
				Start:    JSONPos{Line: caller.Start.Line, Column: caller.Start.Column},
				End:      JSONPos{Line: caller.End.Line, Column: caller.End.Column},
			}
		}
	}

	if appErr != nil {
		var diags hcl.Diagnostics
		if errors.As(appErr, &diags) {
			ret.Errors = make([]JSONError, len(diags))
			for idx, diag := range diags {
				ret.Errors[idx] = JSONError{
					Severity: fromHclSeverity(diag.Severity),
					Summary:  diag.Summary,
					Message:  diag.Detail,
					Range: &JSONRange{
						Filename: diag.Subject.Filename,
						Start:    JSONPos{Line: diag.Subject.Start.Line, Column: diag.Subject.Start.Column},
						End:      JSONPos{Line: diag.Subject.End.Line, Column: diag.Subject.End.Column},
					},
				}
			}
		} else {
			ret.Errors = []JSONError{{
				Severity: toSeverity(tflint.ERROR),
				Message:  appErr.Error(),
			}}
		}
	}

	out, err := json.Marshal(ret)
	if err != nil {
		fmt.Fprint(f.Stderr, err)
	}
	fmt.Fprint(f.Stdout, string(out))
}
