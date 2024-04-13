package formatter

import (
	"encoding/xml"
	"fmt"

	"github.com/jstemmer/go-junit-report/formatter"
	"github.com/terraform-linters/tflint/tflint"
)

// https://www.ibm.com/docs/en/developer-for-zos/14.1.0?topic=formats-junit-xml-format

func (f *Formatter) junitPrint(issues tflint.Issues, appErr error, sources map[string][]byte) {
	cases := make([]formatter.JUnitTestCase, len(issues))

	for i, issue := range issues.Sort() {
		cases[i] = formatter.JUnitTestCase{
			Name:      issue.Rule.Name(),
			Classname: issue.Range.Filename,
			Time:      "0",
			Failure: &formatter.JUnitFailure{
				Message: fmt.Sprintf("%s: %s", issue.Range, issue.Message),
				Type:    issue.Rule.Severity().String(),
				Contents: fmt.Sprintf(
					"%s: %s\nRule: %s\nRange: %s",
					issue.Rule.Severity(),
					issue.Message,
					issue.Rule.Name(),
					issue.Range,
				),
			},
		}
	}

	suites := formatter.JUnitTestSuites{
		Suites: []formatter.JUnitTestSuite{
			{
				Time:      "0",
				Tests:     len(issues),
				Failures:  len(issues),
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

	if appErr != nil {
		f.prettyPrintErrors(appErr, sources, false)
	}
}
