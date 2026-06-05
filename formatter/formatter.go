package formatter

import (
	"errors"
	"fmt"
	"io"

	hcl "github.com/hashicorp/hcl/v2"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/tflint"
)

const applicationErrorSource = "(application)"

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

// format is a per-format adapter owning how a format prints output and
// whether it buffers parallel errors to the end instead of printing them in real time.
type format interface {
	print(f *Formatter, issues tflint.Issues, err error, sources map[string][]byte)
	buffersErrors() bool
}

// bufferedFormat is embedded by formats that accumulate parallel errors and
// print them at the end rather than in real time. Only pretty streams errors.
type bufferedFormat struct{}

func (bufferedFormat) buffersErrors() bool { return true }

var formats = map[string]format{
	"default":    prettyFormat{},
	"json":       jsonFormat{},
	"checkstyle": checkstyleFormat{},
	"junit":      junitFormat{},
	"compact":    compactFormat{},
	"sarif":      sarifFormat{},
}

func (f *Formatter) resolveFormat() format {
	if format, ok := formats[f.Format]; ok {
		return format
	}
	return prettyFormat{} // unknown format falls back to pretty, matching today's default
}

// Print outputs the given issues and errors according to configured format
func (f *Formatter) Print(issues tflint.Issues, err error, sources map[string][]byte) {
	f.resolveFormat().print(f, issues, err, sources)
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

	if f.resolveFormat().buffersErrors() {
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
	if f.resolveFormat().buffersErrors() {
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
