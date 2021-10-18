package formatter

import (
	"fmt"

	"github.com/owenrumney/go-sarif/sarif"
	"github.com/terraform-linters/tflint/tflint"
)

func (f *Formatter) sarifPrint(issues tflint.Issues, tferr *tflint.Error, sources map[string][]byte) {
	report, err := sarif.New(sarif.Version210)
	if err != nil {
		panic(err)
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

	err = report.PrettyWrite(f.Stdout)
	if err != nil {
		panic(err)
	}
}
