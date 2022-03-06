package formatter

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/owenrumney/go-sarif/sarif"
	"github.com/terraform-linters/tflint/tflint"
)

func (f *Formatter) sarifPrint(issues tflint.Issues, appErr error) {
	report, initErr := sarif.New(sarif.Version210)
	if initErr != nil {
		panic(initErr)
	}

	run := sarif.NewRun("tflint", "https://github.com/terraform-linters/tflint")
	report.AddRun(run)

	for _, issue := range issues {
		rule := run.AddRule(issue.Rule.Name()).WithHelpURI(issue.Rule.Link()).WithDescription("")

		var level string
		switch issue.Rule.Severity() {
		case tflint.ERROR:
			level = "error"
		case tflint.NOTICE:
			level = "note"
		case tflint.WARNING:
			level = "warning"
		default:
			panic(fmt.Errorf("Unexpected lint type: %s", issue.Rule.Severity()))
		}

		endLine := issue.Range.End.Line
		if endLine == 0 {
			endLine = 1
		}
		endColumn := issue.Range.End.Column
		if endColumn == 0 {
			endColumn = 1
		}

		location := sarif.NewPhysicalLocation().
			WithArtifactLocation(sarif.NewSimpleArtifactLocation(issue.Range.Filename)).
			WithRegion(
				sarif.NewRegion().
					WithStartLine(issue.Range.Start.Line).
					WithStartColumn(issue.Range.Start.Column).
					WithEndLine(endLine).
					WithEndColumn(endColumn),
			)

		run.AddResult(rule.ID).
			WithLevel(level).
			WithLocation(sarif.NewLocationWithPhysicalLocation(location)).
			WithMessage(sarif.NewTextMessage(issue.Message))
	}

	errRun := sarif.NewRun("tflint-errors", "https://github.com/terraform-linters/tflint")
	report.AddRun(errRun)
	if appErr != nil {
		var diags hcl.Diagnostics
		if errors.As(appErr, &diags) {
			for _, diag := range diags {
				location := sarif.NewPhysicalLocation().
					WithArtifactLocation(sarif.NewSimpleArtifactLocation(diag.Subject.Filename)).
					WithRegion(
						sarif.NewRegion().
							WithByteOffset(diag.Subject.Start.Byte).
							WithByteLength(diag.Subject.End.Byte - diag.Subject.Start.Byte).
							WithStartLine(diag.Subject.Start.Line).
							WithStartColumn(diag.Subject.Start.Column).
							WithEndLine(diag.Subject.End.Line).
							WithEndColumn(diag.Subject.End.Column),
					)

				errRun.AddResult(diag.Summary).
					WithLevel(fromHclSeverity(diag.Severity)).
					WithLocation(sarif.NewLocationWithPhysicalLocation(location)).
					WithMessage(sarif.NewTextMessage(diag.Detail))
			}
		} else {
			errRun.AddResult("application_error").
				WithLevel("error").
				WithMessage(sarif.NewTextMessage(appErr.Error()))
		}
	}

	stdoutErr := report.PrettyWrite(f.Stdout)
	if stdoutErr != nil {
		panic(stdoutErr)
	}
}
