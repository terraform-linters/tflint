package formatter

import (
	"fmt"

	"github.com/terraform-linters/tflint/tflint"
)

func (f *Formatter) compactPrint(issues tflint.Issues, appErr error, sources map[string][]byte) {
	if len(issues) > 0 {
		fmt.Fprintf(f.Stdout, "%d issue(s) found:\n\n", len(issues))
	}

	for _, issue := range issues {
		fmt.Fprintf(
			f.Stdout,
			"%s:%d:%d: %s - %s (%s)\n",
			issue.Range.Filename,
			issue.Range.Start.Line,
			issue.Range.Start.Column,
			issue.Rule.Severity(),
			issue.Message,
			issue.Rule.Name(),
		)
	}

	if appErr != nil {
		f.prettyPrintErrors(appErr, sources)
	}
}
