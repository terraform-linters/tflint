package formatter

import (
	"errors"
	"fmt"
	"io"
	"slices"

	hcl "github.com/hashicorp/hcl/v2"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/tflint"
)

// Formatter outputs appropriate results to stdout and stderr depending on the format
type Formatter struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Format  string
	Fix     bool
	NoColor bool

	// Errors occurred in parallel workers.
	// Some formats do not output immediately, so they are saved here.
	errInParallel error
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

// PrintErrorParallel outputs an error occurred in parallel workers.
// Depending on the configured format, errors may not be output immediately.
// This function itself is called serially, so changes to f.errInParallel are safe.
func (f *Formatter) PrintErrorParallel(err error, sources map[string][]byte) {
	if f.errInParallel == nil {
		f.errInParallel = err
	} else {
		f.errInParallel = errors.Join(f.errInParallel, err)
	}

	if slices.Contains([]string{"json", "checkstyle", "junit", "compact", "sarif"}, f.Format) {
		// These formats require errors to be printed at the end, so do nothing here
		return
	}

	// Print errors immediately for other formats
	f.prettyPrintErrors(err, sources, true)
}

// PrintParallel outputs issues and errors in parallel workers.
// Errors stored with PrintErrorParallel are output,
// but in the default format they are output in real time, so they are ignored.
func (f *Formatter) PrintParallel(issues tflint.Issues, sources map[string][]byte) error {
	if slices.Contains([]string{"json", "checkstyle", "junit", "compact", "sarif"}, f.Format) {
		f.Print(issues, f.errInParallel, sources)
		return f.errInParallel
	}

	if f.errInParallel != nil {
		// Do not print the errors since they are already printed in real time
		return f.errInParallel
	}

	f.Print(issues, nil, sources)
	return nil
}

func toSeverity(lintType tflint.Severity) string {
	switch lintType {
	case sdk.ERROR:
		return "error"
	case sdk.WARNING:
		return "warning"
	case sdk.NOTICE:
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
