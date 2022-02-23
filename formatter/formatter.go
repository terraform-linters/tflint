package formatter

import (
	"fmt"
	"io"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// Formatter outputs appropriate results to stdout and stderr depending on the format
type Formatter struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Format  string
	NoColor bool
}

// Print outputs the given issues and errors according to configured format
func (f *Formatter) Print(issues tflint.Issues, err error, sources map[string][]byte) {
	switch f.Format {
	case "default":
		f.prettyPrint(issues, err, sources)
	case "json":
		f.jsonPrint(issues, err)
	case "checkstyle":
		f.checkstylePrint(issues, err, sources)
	case "junit":
		f.junitPrint(issues, err, sources)
	case "compact":
		f.compactPrint(issues, err, sources)
	case "sarif":
		f.sarifPrint(issues, err)
	default:
		f.prettyPrint(issues, err, sources)
	}
}

func toSeverity(lintType tflint.Severity) string {
	switch lintType {
	case tflint.ERROR:
		return "error"
	case tflint.WARNING:
		return "warning"
	case tflint.NOTICE:
		return "info"
	default:
		panic(fmt.Errorf("Unexpected lint type: %s", lintType))
	}
}

func fromHclSeverity(severity hcl.DiagnosticSeverity) string {
	switch severity {
	case hcl.DiagError:
		return "error"
	case hcl.DiagWarning:
		return "warning"
	default:
		panic(fmt.Errorf("Unexpected HCL severity: %v", severity))
	}
}
