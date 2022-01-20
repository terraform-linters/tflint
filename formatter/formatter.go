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
func (f *Formatter) Print(issues tflint.Issues, tferr *tflint.Error, sources map[string][]byte) {

	if tferr != nil {
		if diags, ok := tferr.Cause.(hcl.Diagnostics); ok {
			tferr.Cause = tflint.ConfigParseError{&diags}
		}
	}

	switch f.Format {
	case "default":
		f.prettyPrint(issues, tferr, sources)
	case "json":
		f.jsonPrint(issues, tferr)
	case "checkstyle":
		f.checkstylePrint(issues, tferr, sources)
	case "junit":
		f.junitPrint(issues, tferr, sources)
	case "compact":
		f.compactPrint(issues, tferr, sources)
	case "sarif":
		f.sarifPrint(issues, tferr)
	default:
		f.prettyPrint(issues, tferr, sources)
	}
}

func toSeverity(lintType string) string {
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
