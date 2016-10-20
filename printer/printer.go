package printer

import (
	"fmt"
	"io"

	"github.com/wata727/tflint/issue"
)

type Printer struct {
	stdout io.Writer
	stderr io.Writer
}

func Print(issues []*issue.Issue, stdout io.Writer, stderr io.Writer) {
	printer := &Printer{
		stdout: stdout,
		stderr: stderr,
	}
	printer.Print(issues)
}

func (p *Printer) Print(issues []*issue.Issue) {
	for _, issue := range issues {
		fmt.Fprintf(p.stdout, "%s: %s Line: %d in %s\n", issue.Type, issue.Message, issue.Line, issue.File)
	}
}
