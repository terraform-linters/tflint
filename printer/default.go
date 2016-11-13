package printer

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/fatih/color"
	"github.com/wata727/tflint/issue"
)

var fileColor = color.New(color.Bold).SprintfFunc()
var errorColor = color.New(color.FgRed).SprintFunc()
var warningColor = color.New(color.FgYellow).SprintFunc()
var noticeColor = color.New(color.FgHiWhite).SprintFunc()
var successColor = color.New(color.FgHiGreen).SprintFunc()

func (p *Printer) DefaultPrint(issues []*issue.Issue) {
	if len(issues) == 0 {
		p.printAwesome()
		return
	}

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
		issuePrefix := i.Type + ":" + strconv.Itoa(i.Line)

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

	allResult := fileColor(strconv.Itoa(len(issues)) + " issues")
	eResult := errorColor(strconv.Itoa(len(eIssues)) + " errors")
	wResult := warningColor(strconv.Itoa(len(wIssues)) + " warnings")
	nResult := noticeColor(strconv.Itoa(len(nIssues)) + " notices")

	fmt.Fprintf(p.stdout, "\nResult: %s  (%s , %s , %s)\n", allResult, eResult, wResult, nResult)
}

func (p *Printer) printAwesome() {
	fmt.Fprintf(p.stdout, "%s\n", successColor("Awesome! Your code is following the best practices :)"))
}
