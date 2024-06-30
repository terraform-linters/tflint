package formatter

import (
	"encoding/xml"
	"fmt"

	"github.com/terraform-linters/tflint/tflint"
)

type checkstyleError struct {
	Source   string `xml:"source,attr"`
	Line     int    `xml:"line,attr"`
	Column   int    `xml:"column,attr"`
	Severity string `xml:"severity,attr"`
	Message  string `xml:"message,attr"`
	Link     string `xml:"link,attr"`

	// Deprecated: Use `source` instead
	Rule string `xml:"rule,attr"`
}

type checkstyleFile struct {
	Name   string             `xml:"name,attr"`
	Errors []*checkstyleError `xml:"error"`
}

type checkstyle struct {
	XMLName xml.Name          `xml:"checkstyle"`
	Files   []*checkstyleFile `xml:"file"`
}

func (f *Formatter) checkstylePrint(issues tflint.Issues, appErr error, sources map[string][]byte) {
	files := map[string]*checkstyleFile{}
	for _, issue := range issues {
		cherr := &checkstyleError{
			Source:   issue.Rule.Name(),
			Line:     issue.Range.Start.Line,
			Column:   issue.Range.Start.Column,
			Severity: toSeverity(issue.Rule.Severity()),
			Message:  issue.Message,
			Link:     issue.Rule.Link(),

			Rule: issue.Rule.Name(),
		}

		if file, exists := files[issue.Range.Filename]; exists {
			file.Errors = append(file.Errors, cherr)
		} else {
			files[issue.Range.Filename] = &checkstyleFile{
				Name:   issue.Range.Filename,
				Errors: []*checkstyleError{cherr},
			}
		}
	}

	ret := &checkstyle{}
	for _, file := range files {
		ret.Files = append(ret.Files, file)
	}

	out, err := xml.MarshalIndent(ret, "", "  ")
	if err != nil {
		fmt.Fprint(f.Stderr, err)
	}
	fmt.Fprint(f.Stdout, xml.Header)
	fmt.Fprint(f.Stdout, string(out))

	if appErr != nil {
		f.prettyPrintErrors(appErr, sources, false)
	}
}
