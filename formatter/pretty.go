package formatter

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/fatih/color"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/tflint"
)

var colorBold = color.New(color.Bold).SprintfFunc()
var colorHighlight = color.New(color.Bold).Add(color.Underline).SprintFunc()
var colorError = color.New(color.FgRed).SprintFunc()
var colorWarning = color.New(color.FgYellow).SprintFunc()
var colorNotice = color.New(color.FgHiWhite).SprintFunc()

func (f *Formatter) prettyPrint(issues tflint.Issues, err error, sources map[string][]byte) {
	if len(issues) > 0 {
		fmt.Fprintf(f.Stdout, "%d issue(s) found:\n\n", len(issues))

		for _, issue := range issues.Sort() {
			f.prettyPrintIssueWithSource(issue, sources)
		}
	}

	if err != nil {
		f.prettyPrintErrors(err, sources, false)
	}
}

func (f *Formatter) prettyPrintIssueWithSource(issue *tflint.Issue, sources map[string][]byte) {
	message := issue.Message
	if issue.Fixable {
		if f.Fix {
			message = "[Fixed] " + message
		} else {
			message = "[Fixable] " + message
		}
	}

	fmt.Fprintf(
		f.Stdout,
		"%s: %s (%s)\n\n",
		colorSeverity(issue.Rule.Severity()), colorBold(message), issue.Rule.Name(),
	)
	fmt.Fprintf(f.Stdout, "  on %s line %d:\n", issue.Range.Filename, issue.Range.Start.Line)

	var src []byte
	if issue.Source != nil {
		src = issue.Source
	} else {
		src = sources[issue.Range.Filename]
	}

	if src == nil {
		fmt.Fprintf(f.Stdout, "   (source code not available)\n")
	} else {
		sc := hcl.NewRangeScanner(src, issue.Range.Filename, bufio.ScanLines)

		for sc.Scan() {
			lineRange := sc.Range()
			if !lineRange.Overlaps(issue.Range) {
				continue
			}

			beforeRange, highlightedRange, afterRange := lineRange.PartitionAround(issue.Range)
			if highlightedRange.Empty() {
				fmt.Fprintf(f.Stdout, "%4d: %s\n", lineRange.Start.Line, sc.Bytes())
			} else {
				before := beforeRange.SliceBytes(src)
				highlighted := highlightedRange.SliceBytes(src)
				after := afterRange.SliceBytes(src)
				fmt.Fprintf(
					f.Stdout,
					"%4d: %s%s%s\n",
					lineRange.Start.Line,
					before,
					colorHighlight(string(highlighted)),
					after,
				)
			}
		}
	}

	if len(issue.Callers) > 0 {
		fmt.Fprint(f.Stdout, "\nCallers:\n")
		for _, caller := range issue.Callers {
			fmt.Fprintf(f.Stdout, "   %s\n", caller)
		}
	}

	if issue.Rule.Link() != "" {
		fmt.Fprintf(f.Stdout, "\nReference: %s\n", issue.Rule.Link())
	}

	fmt.Fprint(f.Stdout, "\n")
}

func (f *Formatter) prettyPrintErrors(err error, sources map[string][]byte, withIndent bool) {
	if err == nil {
		return
	}

	// errors.Join
	if errs, ok := err.(interface{ Unwrap() []error }); ok {
		for _, err := range errs.Unwrap() {
			f.prettyPrintErrors(err, sources, withIndent)
		}
		return
	}

	// hcl.Diagnostics
	var diags hcl.Diagnostics
	if errors.As(err, &diags) {
		fmt.Fprintf(f.Stderr, "%s:\n\n", err)

		writer := hcl.NewDiagnosticTextWriter(f.Stderr, parseSources(sources), 0, !f.NoColor)
		_ = writer.WriteDiagnostics(diags)
		return
	}

	if withIndent {
		fmt.Fprintf(f.Stderr, "%s %s\n", colorError("│"), err)
	} else {
		fmt.Fprintf(f.Stderr, "%s\n", err)
	}
}

// PrettyPrintStderr outputs the given output to stderr with an indent.
func (f *Formatter) PrettyPrintStderr(output string) {
	fmt.Fprintf(f.Stderr, "%s %s\n", colorWarning("│"), output)
}

func parseSources(sources map[string][]byte) map[string]*hcl.File {
	ret := map[string]*hcl.File{}
	parser := hclparse.NewParser()

	var file *hcl.File
	var diags hcl.Diagnostics
	for filename, src := range sources {
		if strings.HasSuffix(filename, ".json") {
			file, diags = parser.ParseJSON(src, filename)
		} else {
			file, diags = parser.ParseHCL(src, filename)
		}

		if diags.HasErrors() {
			log.Printf("[WARN] Failed to parse %s. This file is not available in output. Reason: %s", filename, diags.Error())
		}
		ret[filename] = file
	}

	return ret
}

func colorSeverity(severity tflint.Severity) string {
	switch severity {
	case sdk.ERROR:
		return colorError(severity)
	case sdk.WARNING:
		return colorWarning(severity)
	case sdk.NOTICE:
		return colorNotice(severity)
	default:
		panic("Unreachable")
	}
}
