package formatter

import (
	"fmt"
	"io"

	"github.com/wata727/tflint/tflint"
)

// Formatter outputs appropriate results to stdout and stderr depending on the format
type Formatter struct {
	Stdout io.Writer
	Stderr io.Writer
	Format string
}

// Print outputs the given issues and errors according to configured format
func (f *Formatter) Print(issues tflint.Issues, err *tflint.Error, sources map[string][]byte) {
	switch f.Format {
	case "default":
		f.prettyPrint(issues, err, sources)
	case "json":
		f.jsonPrint(issues, err)
	case "checkstyle":
		f.checkstylePrint(issues, err, sources)
	default:
		f.prettyPrint(issues, err, sources)
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
