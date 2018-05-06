package printer

import (
	"io"

	"github.com/wata727/tflint/issue"
)

type PrinterIF interface {
	Print(issues []*issue.Issue, format string, quiet bool)
}

type Printer struct {
	stdout io.Writer
	stderr io.Writer
}

func NewPrinter(stdout io.Writer, stderr io.Writer) *Printer {
	return &Printer{
		stdout: stdout,
		stderr: stderr,
	}
}

func (p *Printer) Print(issues []*issue.Issue, format string, quiet bool) {
	switch format {
	case "default":
		p.DefaultPrint(issues, quiet)
	case "json":
		p.JSONPrint(issues)
	case "checkstyle":
		p.CheckstylePrint(issues)
	default:
		p.DefaultPrint(issues, quiet)
	}
}
