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
		f.prettyPrintErrors(err, sources)
	}
}

func (f *Formatter) prettyPrintIssueWithSource(issue *tflint.Issue, sources map[string][]byte) {
	fmt.Fprintf(
		f.Stdout,
		"%s: %s (%s)\n\n",
		colorSeverity(issue.Rule.Severity()), colorBold(issue.Message), issue.Rule.Name(),
	)
	fmt.Fprintf(f.Stdout, "  on %s line %d:\n", issue.Range.Filename, issue.Range.Start.Line)

	src := sources[issue.Range.Filename]

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

func (f *Formatter) prettyPrintErrors(err error, sources map[string][]byte) {
	var diags hcl.Diagnostics
	if errors.As(err, &diags) {
		fmt.Fprintf(f.Stderr, "%s:\n\n", err)

		writer := hcl.NewDiagnosticTextWriter(f.Stderr, parseSources(sources), 0, !f.NoColor)
		_ = writer.WriteDiagnostics(diags)
	} else {
		fmt.Fprintf(f.Stderr, "%s\n", err)
	}
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
	case tflint.ERROR:
		return colorError(severity)
	case tflint.WARNING:
		return colorWarning(severity)
	case tflint.NOTICE:
		return colorNotice(severity)
	default:
		panic("Unreachable")
	}
}
