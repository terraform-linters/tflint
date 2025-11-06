package formatter

import (
	"encoding/xml"
	"fmt"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/jstemmer/go-junit-report/formatter"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
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

	// Add application errors as test case failures
	errorCases := f.junitErrors(appErr)
	cases = append(cases, errorCases...)

	suites := formatter.JUnitTestSuites{
		Suites: []formatter.JUnitTestSuite{
			{
				Time:      "0",
				Tests:     len(cases),
				Failures:  len(cases),
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
}

func (f *Formatter) junitErrors(err error) []formatter.JUnitTestCase {
	return mapErrors(err, errorMapper[formatter.JUnitTestCase]{
		diagnostic: func(diag *hcl.Diagnostic) formatter.JUnitTestCase {
			return formatter.JUnitTestCase{
				Name:      diag.Summary,
				Classname: diag.Subject.Filename,
				Time:      "0",
				Failure: &formatter.JUnitFailure{
					Message: fmt.Sprintf("%s:%d,%d-%d,%d: %s",
						diag.Subject.Filename,
						diag.Subject.Start.Line,
						diag.Subject.Start.Column,
						diag.Subject.End.Line,
						diag.Subject.End.Column,
						diag.Detail,
					),
					Type: fromHclSeverity(diag.Severity),
					Contents: fmt.Sprintf(
						"%s: %s\nSummary: %s\nRange: %s:%d,%d-%d,%d",
						fromHclSeverity(diag.Severity),
						diag.Detail,
						diag.Summary,
						diag.Subject.Filename,
						diag.Subject.Start.Line,
						diag.Subject.Start.Column,
						diag.Subject.End.Line,
						diag.Subject.End.Column,
					),
				},
			}
		},
		error: func(err error) formatter.JUnitTestCase {
			return formatter.JUnitTestCase{
				Name:      "application_error",
				Classname: "(application)",
				Time:      "0",
				Failure: &formatter.JUnitFailure{
					Message:  err.Error(),
					Type:     toSeverity(sdk.ERROR),
					Contents: fmt.Sprintf("Error: %s", err.Error()),
				},
			}
		},
	})
}
