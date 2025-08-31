package formatter

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/owenrumney/go-sarif/v2/sarif"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/tflint"
)

func (f *Formatter) sarifPrint(issues tflint.Issues, appErr error) {
	report, initErr := sarif.New(sarif.Version210)
	if initErr != nil {
		panic(initErr)
	}

	run := sarif.NewRunWithInformationURI("tflint", "https://github.com/terraform-linters/tflint")

	version := tflint.Version.String()
	run.Tool.Driver.Version = &version

	report.AddRun(run)

	for _, issue := range issues {
		rule := run.AddRule(issue.Rule.Name()).
			WithHelpURI(issue.Rule.Link()).
			WithDescription("")

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

		result := run.CreateResultForRule(rule.ID).
			WithLevel(level).
			WithMessage(sarif.NewTextMessage(issue.Message))

		if location != nil {
			result.AddLocation(sarif.NewLocationWithPhysicalLocation(location))
		}
	}

	errRun := sarif.NewRunWithInformationURI("tflint-errors", "https://github.com/terraform-linters/tflint")
	errRun.Tool.Driver.Version = &version

	report.AddRun(errRun)
	f.sarifAddErrors(errRun, appErr)

	stdoutErr := report.PrettyWrite(f.Stdout)
	if stdoutErr != nil {
		panic(stdoutErr)
	}
}

func (f *Formatter) sarifAddErrors(errRun *sarif.Run, err error) {
	if err == nil {
		return
	}

	// errors.Join
	if errs, ok := err.(interface{ Unwrap() []error }); ok {
		for _, err := range errs.Unwrap() {
			f.sarifAddErrors(errRun, err)
		}
		return
	}

	// hcl.Diagnostics
	var diags hcl.Diagnostics
	if errors.As(err, &diags) {
		for _, diag := range diags {
			rule := errRun.AddRule(diag.Summary).WithDescription("")

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

			errRun.CreateResultForRule(rule.ID).
				WithLevel(fromHclSeverity(diag.Severity)).
				WithMessage(sarif.NewTextMessage(diag.Detail)).
				AddLocation(sarif.NewLocationWithPhysicalLocation(location))
		}
		return
	}

	rule := errRun.AddRule("application_error").WithDescription("")

	errRun.CreateResultForRule(rule.ID).
		WithLevel("error").
		WithMessage(sarif.NewTextMessage(err.Error()))
}
