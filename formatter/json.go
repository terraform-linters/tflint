package formatter

import (
	"encoding/json"
	"fmt"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/wata727/tflint/tflint"
)

type jsonIssue struct {
	Rule    jsonRule    `json:"rule"`
	Message string      `json:"message"`
	Range   jsonRange   `json:"range"`
	Callers []jsonRange `json:"callers"`
}

type jsonRule struct {
	Name     string `json:"name"`
	Severity string `json:"severity"`
	Link     string `json:"link"`
}

type jsonRange struct {
	Filename string  `json:"filename"`
	Start    jsonPos `json:"start"`
	End      jsonPos `json:"end"`
}

type jsonPos struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type jsonError struct {
	Message string `json:"message"`
}

// JSONOutput is a temporary structure for converting to JSON
type JSONOutput struct {
	Issues []jsonIssue `json:"issues"`
	Errors []jsonError `json:"errors"`
}

func (f *Formatter) jsonPrint(issues tflint.Issues, tferr *tflint.Error) {
	ret := &JSONOutput{Issues: make([]jsonIssue, len(issues)), Errors: []jsonError{}}

	for idx, issue := range issues.Sort() {
		ret.Issues[idx] = jsonIssue{
			Rule: jsonRule{
				Name:     issue.Rule.Name(),
				Severity: toSeverity(issue.Rule.Severity()),
				Link:     issue.Rule.Link(),
			},
			Message: issue.Message,
			Range: jsonRange{
				Filename: issue.Range.Filename,
				Start:    jsonPos{Line: issue.Range.Start.Line, Column: issue.Range.Start.Column},
				End:      jsonPos{Line: issue.Range.End.Line, Column: issue.Range.End.Column},
			},
			Callers: make([]jsonRange, len(issue.Callers)),
		}
		for i, caller := range issue.Callers {
			ret.Issues[idx].Callers[i] = jsonRange{
				Filename: caller.Filename,
				Start:    jsonPos{Line: caller.Start.Line, Column: caller.Start.Column},
				End:      jsonPos{Line: caller.End.Line, Column: caller.End.Column},
			}
		}
	}

	if tferr != nil {
		var errs []error
		if diags, ok := tferr.Cause.(hcl.Diagnostics); ok {
			errs = diags.Errs()
		} else {
			errs = []error{tferr.Cause}
		}

		ret.Errors = make([]jsonError, len(errs))
		for idx, err := range errs {
			ret.Errors[idx] = jsonError{Message: err.Error()}
		}
	}

	out, err := json.Marshal(ret)
	if err != nil {
		fmt.Fprint(f.Stderr, err)
	}
	fmt.Fprint(f.Stdout, string(out))
}
