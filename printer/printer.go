package printer

import (
	"io"

	"github.com/wata727/tflint/issue"
)

type PrinterIF interface {
	Print(issues []*issue.Issue, format string)
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

func (p *Printer) Print(issues []*issue.Issue, format string) {
	switch format {
	case "default":
		p.DefaultPrint(issues)
	case "json":
		p.JSONPrint(issues)
	case "checkstyle":
		p.CheckstylePrint(issues)
	default:
		p.DefaultPrint(issues)
	}
}
