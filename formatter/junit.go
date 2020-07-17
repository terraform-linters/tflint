package formatter

import (
	"encoding/xml"
	"fmt"

	"github.com/terraform-linters/tflint/tflint"

	"github.com/jstemmer/go-junit-report/formatter"
)

func (f *Formatter) junitPrint(issues tflint.Issues, tferr *tflint.Error, sources map[string][]byte) {
	cases := make([]formatter.JUnitTestCase, len(issues))

	for _, issue := range issues.Sort() {
		cases = append(cases, formatter.JUnitTestCase{
			Name: issue.Rule.Name(),
			Classname: issue.Range.Filename,
			Time: "0",
			Failure: &formatter.JUnitFailure{
				Message: issue.Message,
				Contents: fmt.Sprintf(
					"line %d, col %d, %s - %s (%s)",
					issue.Range.Start.Line,
					issue.Range.Start.Column,
					issue.Rule.Severity(),
					issue.Message,
					issue.Rule.Name(),
				),
			},
		})
	}

	suites := formatter.JUnitTestSuites{
		Suites: []formatter.JUnitTestSuite{
			formatter.JUnitTestSuite{
				Time: "0",
				Tests: len(issues),
				Failures: len(issues),
				TestCases: cases,
			},
		},
	}

	out, err := xml.MarshalIndent(suites, "", "  ")
	if err != nil {
		fmt.Fprint(f.Stderr, err)
	}
	fmt.Fprint(f.Stdout, xml.Header)
	fmt.Fprint(f.Stdout, string(out))

	if tferr != nil {
		f.printErrors(tferr, sources)
	}
}
