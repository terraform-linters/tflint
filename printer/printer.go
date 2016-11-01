package printer

import (
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/fatih/color"
	"github.com/wata727/tflint/issue"
)

type Printer struct {
	stdout io.Writer
	stderr io.Writer
}

var fileColor = color.New(color.Bold).SprintfFunc()
var errorColor = color.New(color.FgRed).SprintFunc()
var warningColor = color.New(color.FgYellow).SprintFunc()
var noticeColor = color.New(color.FgHiWhite).SprintFunc()

func Print(issues []*issue.Issue, stdout io.Writer, stderr io.Writer) {
	printer := &Printer{
		stdout: stdout,
		stderr: stderr,
	}
	printer.Print(issues)
}

func (p *Printer) Print(issues []*issue.Issue) {
	sort.Sort(issue.ByFile{issue.Issues(issues)})
	bIssue := issues[0]
	sIssues := []*issue.Issue{
		issues[0],
	}

	for _, is := range issues[1:] {
		if bIssue.File != is.File {
			p.printByLine(sIssues)
			sIssues = nil
		}

		sIssues = append(sIssues, is)
		bIssue = is
	}
	if sIssues != nil {
		p.printByLine(sIssues)
	}

	p.printSummary(issues)
}

func (p *Printer) printByLine(issues []*issue.Issue) {
	sort.Sort(issue.ByLine{issue.Issues(issues)})

	fmt.Fprintf(p.stdout, "%s\n", fileColor(issues[0].File))
	for _, i := range issues {
		var issuePrefix string = i.Type + ":" + strconv.Itoa(i.Line)

		switch i.Type {
		case "ERROR":
			issuePrefix = errorColor(issuePrefix)
		case "WARNING":
			issuePrefix = warningColor(issuePrefix)
		case "NOTICE":
			issuePrefix = noticeColor(issuePrefix)
		}
		fmt.Fprintf(p.stdout, "\t%s %s\n", issuePrefix, i.Message)
	}
}

func (p *Printer) printSummary(issues []*issue.Issue) {
	eIssues := []*issue.Issue{}
	wIssues := []*issue.Issue{}
	nIssues := []*issue.Issue{}

	for _, i := range issues {
		switch i.Type {
		case "ERROR":
			eIssues = append(eIssues, i)
		case "WARNING":
			wIssues = append(wIssues, i)
		case "NOTICE":
			nIssues = append(nIssues, i)
		}
	}

	allResult := fileColor("All Issues: " + strconv.Itoa(len(issues)))
	eResult := errorColor("error " + strconv.Itoa(len(eIssues)))
	wResult := warningColor("warning " + strconv.Itoa(len(wIssues)))
	nResult := noticeColor("notice " + strconv.Itoa(len(nIssues)))

	fmt.Fprintf(p.stdout, "\n%s  %s , %s , %s\n", allResult, eResult, wResult, nResult)
}
