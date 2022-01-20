package formatter

import (
	"encoding/xml"
	"fmt"

	"github.com/terraform-linters/tflint/tflint"
)

type checkstyleIssue struct {
	Rule     string `xml:"rule,attr"`
	Line     int    `xml:"line,attr"`
	Column   int    `xml:"column,attr"`
	Severity string `xml:"severity,attr"`
	Message  string `xml:"message,attr"`
	Link     string `xml:"link,attr,omitempty"`
}

type checkstyleFile struct {
	Name   string             `xml:"name,attr"`
	Issues []*checkstyleIssue `xml:"error"`
}

type checkstyle struct {
	XMLName xml.Name          `xml:"checkstyle"`
	Files   []*checkstyleFile `xml:"file"`
}

func insertOrAppend(files *map[string]*checkstyleFile, filename string, issue *checkstyleIssue) {
	if file, exists := (*files)[filename]; exists {
		file.Issues = append(file.Issues, issue)
	} else {
		(*files)[filename] = &checkstyleFile{
			Name:   filename,
			Issues: []*checkstyleIssue{issue},
		}
	}
}

func (f *Formatter) checkstylePrint(issues tflint.Issues, tferr *tflint.Error) {
	files := map[string]*checkstyleFile{}
	for _, issue := range issues {
		chissue := &checkstyleIssue{
			Rule:     issue.Rule.Name(),
			Line:     issue.Range.Start.Line,
			Column:   issue.Range.Start.Column,
			Severity: toSeverity(issue.Rule.Severity()),
			Message:  issue.Message,
			Link:     issue.Rule.Link(),
		}

		insertOrAppend(&files, issue.Range.Filename, chissue)
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

	if tferr != nil {
		f.checkstylePrintErrors(tferr)
	}
}

func (f *Formatter) checkstylePrintErrors(tferr *tflint.Error) {
	files := map[string]*checkstyleFile{}

	if parseError, ok := tferr.Cause.(tflint.ConfigParseError); ok {
		diags := *parseError.Detail

		for _, diag := range diags {
			chissue := &checkstyleIssue{
				Line:     diag.Subject.Start.Line,
				Column:   diag.Subject.Start.Column,
				Severity: fromHclSeverity(diag.Severity),
				Message:  diag.Detail,
			}

			insertOrAppend(&files, diag.Subject.Filename, chissue)
		}
	} else {
		files[""] = &checkstyleFile{
			Name: "",
			Issues: []*checkstyleIssue{
				&checkstyleIssue{
					Message: tferr.Cause.Error(),
				},
			},
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
	fmt.Fprint(f.Stderr, xml.Header)
	fmt.Fprint(f.Stderr, string(out))
}
