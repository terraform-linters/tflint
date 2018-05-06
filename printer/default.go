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

func (p *Printer) DefaultPrint(issues []*issue.Issue, quiet bool) {
	if len(issues) == 0 {
		if !quiet {
			p.printAwesome()
		}
		return
	}

	sort.Sort(issue.ByFile{Issues: issue.Issues(issues)})
	issuesByFile := []*issue.Issue{}

	for _, i := range issues {
		if len(issuesByFile) != 0 && issuesByFile[len(issuesByFile)-1].File != i.File {
			p.printByLine(issuesByFile)
			issuesByFile = nil
		}
		issuesByFile = append(issuesByFile, i)
	}
	p.printByLine(issuesByFile)
	p.printSummary(issues)
}

func (p *Printer) printByLine(issues []*issue.Issue) {
	sort.Sort(issue.ByLine{Issues: issue.Issues(issues)})

	fmt.Fprintf(p.stdout, "%s\n", fileColor(issues[0].File))
	for _, i := range issues {
		issuePrefix := i.Type + ":" + strconv.Itoa(i.Line)
		message := fmt.Sprintf("%s (%s)", i.Message, i.Detector)

		switch i.Type {
		case issue.ERROR:
			issuePrefix = errorColor(issuePrefix)
		case issue.WARNING:
			issuePrefix = warningColor(issuePrefix)
		case issue.NOTICE:
			issuePrefix = noticeColor(issuePrefix)
		}
		fmt.Fprintf(p.stdout, "\t%s %s\n", issuePrefix, message)
	}
}

func (p *Printer) printSummary(issues []*issue.Issue) {
	eIssues := []*issue.Issue{}
	wIssues := []*issue.Issue{}
	nIssues := []*issue.Issue{}

	for _, i := range issues {
		switch i.Type {
		case issue.ERROR:
			eIssues = append(eIssues, i)
		case issue.WARNING:
			wIssues = append(wIssues, i)
		case issue.NOTICE:
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
