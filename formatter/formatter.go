package formatter

import (
	"io"

	"github.com/wata727/tflint/tflint"
)

type Formatter struct {
	Stdout io.Writer
	Stderr io.Writer
	Format string
	Quiet  bool
}

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
