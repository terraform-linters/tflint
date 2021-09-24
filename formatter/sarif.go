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
		rule := run.AddRule(issue.Rule.Name()).WithHelpURI(issue.Rule.Link())

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

		location := sarif.NewPhysicalLocation().
			WithArtifactLocation(sarif.NewSimpleArtifactLocation(issue.Range.Filename)).
			WithRegion(
				sarif.NewRegion().
					WithStartLine(issue.Range.Start.Line).
					WithStartColumn(issue.Range.Start.Column).
					WithEndLine(issue.Range.End.Line).
					WithEndColumn(issue.Range.End.Column),
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
