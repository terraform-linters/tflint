package printer

import (
	"encoding/xml"
	"fmt"

	"sort"

	"github.com/wata727/tflint/issue"
)

type Error struct {
	Detector string `xml:"detector,attr"`
	Line     int    `xml:"line,attr"`
	Severity string `xml:"severity,attr"`
	Message  string `xml:"message,attr"`
	Link     string `xml:"link,attr"`
}

type File struct {
	Name   string  `xml:"name,attr"`
	Errors []Error `xml:"error"`
}

type Checkstyle struct {
	XMLName xml.Name `xml:"checkstyle"`
	Files   []File   `xml:"file"`
}

func (p *Printer) CheckstylePrint(issues []*issue.Issue) {
	sort.Sort(issue.ByFile{Issues: issue.Issues(issues)})

	v := &Checkstyle{}

	for _, i := range issues {
		if len(v.Files) == 0 || v.Files[len(v.Files)-1].Name != i.File {
			v.Files = append(v.Files, File{Name: i.File})
		}
		v.Files[len(v.Files)-1].Errors = append(
			v.Files[len(v.Files)-1].Errors,
			Error{
				Detector: i.Detector,
				Line:     i.Line,
				Severity: toSeverity(i.Type),
				Message:  i.Message,
				Link:     i.Link,
			},
		)
	}

	result, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprint(p.stderr, err)
	}
	fmt.Fprint(p.stdout, xml.Header)
	fmt.Fprint(p.stdout, string(result))
}

func toSeverity(lintType string) string {
	switch lintType {
	case issue.ERROR:
		return "error"
	case issue.WARNING:
		return "warning"
	case issue.NOTICE:
		return "info"
	default:
		return "unknown"
	}
}
