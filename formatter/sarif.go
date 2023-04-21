package formatter

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/owenrumney/go-sarif/sarif"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/tflint"
)

func (f *Formatter) sarifPrint(issues tflint.Issues, appErr error) {
	report, initErr := sarif.New(sarif.Version210)
	if initErr != nil {
		panic(initErr)
	}

	run := sarif.NewRun("tflint", "https://github.com/terraform-linters/tflint")

	version := tflint.Version.String()
	run.Tool.Driver.Version = &version

	report.AddRun(run)

	for _, issue := range issues {
		rule := run.AddRule(issue.Rule.Name()).WithHelpURI(issue.Rule.Link()).WithDescription("")

		var level string
		switch issue.Rule.Severity() {
		case sdk.ERROR:
			level = "error"
		case sdk.NOTICE:
			level = "note"
		case sdk.WARNING:
			level = "warning"
		default:
			panic(fmt.Errorf("Unexpected lint type: %s", issue.Rule.Severity()))
		}

		var location *sarif.PhysicalLocation
		if issue.Range.Filename != "" {
			location = sarif.NewPhysicalLocation().
				WithArtifactLocation(sarif.NewSimpleArtifactLocation(filepath.ToSlash(issue.Range.Filename)))

			if !issue.Range.Empty() {
				location.WithRegion(
					sarif.NewRegion().
						WithStartLine(issue.Range.Start.Line).
						WithStartColumn(issue.Range.Start.Column).
						WithEndLine(issue.Range.End.Line).
						WithEndColumn(issue.Range.End.Column),
				)
			}
		}

		result := run.AddResult(rule.ID).
			WithLevel(level).
			WithMessage(sarif.NewTextMessage(issue.Message))

		if location != nil {
			result.WithLocation(sarif.NewLocationWithPhysicalLocation(location))
		}
	}

	errRun := sarif.NewRun("tflint-errors", "https://github.com/terraform-linters/tflint")
	errRun.Tool.Driver.Version = &version

	report.AddRun(errRun)

	if appErr != nil {
		var diags hcl.Diagnostics
		if errors.As(appErr, &diags) {
			for _, diag := range diags {
				location := sarif.NewPhysicalLocation().
					WithArtifactLocation(sarif.NewSimpleArtifactLocation(filepath.ToSlash(diag.Subject.Filename))).
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
