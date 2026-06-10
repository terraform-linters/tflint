package formatter

import (
	"fmt"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

type compactFormat struct{ bufferedFormat }

func (compactFormat) print(f *Formatter, issues tflint.Issues, appErr error, sources map[string][]byte) {
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

	f.compactPrintErrors(appErr, sources)
}

func (f *Formatter) compactPrintErrors(err error, sources map[string][]byte) {
	mapErrors(err, errorMapper[struct{}]{
		diagnostics: func(_ error, diags hcl.Diagnostics) []struct{} {
			for _, diag := range diags {
				rng := diagRange(diag)
				fmt.Fprintf(
					f.Stdout,
					"%s:%d:%d: %s - %s. %s\n",
					rng.Filename,
					rng.Start.Line,
					rng.Start.Column,
					fromHclSeverity(diag.Severity),
					diag.Summary,
					diag.Detail,
				)
			}
			return nil
		},
		error: func(err error) struct{} {
			f.prettyPrintErrors(err, sources, false)
			return struct{}{}
		},
	})
}
