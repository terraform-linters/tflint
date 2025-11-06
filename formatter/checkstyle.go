package formatter

import (
	"encoding/xml"
	"fmt"
	"slices"
	"strings"

	hcl "github.com/hashicorp/hcl/v2"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
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

	for _, cherr := range f.checkstyleErrors(appErr) {
		filename := cherr.Source
		if filename == "" {
			filename = applicationErrorSource
		}
		if file, exists := files[filename]; exists {
			file.Errors = append(file.Errors, cherr)
		} else {
			files[filename] = &checkstyleFile{
				Name:   filename,
				Errors: []*checkstyleError{cherr},
			}
		}
	}

	filenames := make([]string, 0, len(files))
	for filename := range files {
		filenames = append(filenames, filename)
	}
	slices.SortFunc(filenames, func(a, b string) int {
		if a == applicationErrorSource {
			return -1
		}
		if b == applicationErrorSource {
			return 1
		}
		return strings.Compare(a, b)
	})

	ret := &checkstyle{}
	for _, filename := range filenames {
		ret.Files = append(ret.Files, files[filename])
	}

	out, err := xml.MarshalIndent(ret, "", "  ")
	if err != nil {
		fmt.Fprint(f.Stderr, err)
	}
	fmt.Fprint(f.Stdout, xml.Header)
	fmt.Fprint(f.Stdout, string(out))
}

func (f *Formatter) checkstyleErrors(err error) []*checkstyleError {
	return mapErrors(err, errorMapper[*checkstyleError]{
		diagnostic: func(diag *hcl.Diagnostic) *checkstyleError {
			return &checkstyleError{
				Source:   diag.Summary,
				Line:     diag.Subject.Start.Line,
				Column:   diag.Subject.Start.Column,
				Severity: fromHclSeverity(diag.Severity),
				Message:  diag.Detail,
			}
		},
		error: func(err error) *checkstyleError {
			return &checkstyleError{
				Source:   applicationErrorSource,
				Severity: toSeverity(sdk.ERROR),
				Message:  err.Error(),
				Rule:     applicationErrorSource,
			}
		},
	})
}
