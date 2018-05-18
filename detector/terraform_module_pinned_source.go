package detector

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type TerraformModulePinnedSourceDetector struct {
	*Detector
}

func (d *Detector) CreateTerraformModulePinnedSourceDetector() *TerraformModulePinnedSourceDetector {
	nd := &TerraformModulePinnedSourceDetector{Detector: d}
	nd.Name = "terraform_module_pinned_source"
	nd.IssueType = issue.WARNING
	nd.TargetType = "module"
	nd.DeepCheck = false
	nd.Link = "https://github.com/wata727/tflint/blob/master/docs/terraform_module_pinned_source.md"
	return nd
}

func (d *TerraformModulePinnedSourceDetector) Detect(module *schema.Module, issues *[]*issue.Issue) {
	lower := strings.ToLower(module.ModuleSource)

	reGithub := regexp.MustCompile("(^github.com/(.+)/(.+)$)|(^git@github.com:(.+)/(.+)$)")
	reBitbucket := regexp.MustCompile("^bitbucket.org/(.+)/(.+)$")
	reGenericGit := regexp.MustCompile("(git://(.+)/(.+))|(git::https://(.+)/(.+))|(git::ssh://((.+)@)??(.+)/(.+)/(.+))")

	if reGithub.MatchString(lower) || reBitbucket.MatchString(lower) || reGenericGit.MatchString(lower) {
		if issue := d.detectGitSource(module); issue != nil {
			*issues = append(*issues, issue)
		}
	} else if strings.HasPrefix(lower, "hg::") {
		if issue := d.detectMercurialSource(module); issue != nil {
			*issues = append(*issues, issue)
		}
	}
}

func (d *TerraformModulePinnedSourceDetector) detectGitSource(module *schema.Module) *issue.Issue {
	lower := strings.ToLower(module.ModuleSource)
	sourceToken, _ := module.GetToken("source")

	if strings.Contains(lower, "ref=") {
		if strings.Contains(lower, "ref=master") {
			return &issue.Issue{
				Detector: d.Name,
				Type:     d.IssueType,
				Message:  fmt.Sprintf("Module source \"%s\" uses default ref \"master\"", module.ModuleSource),
				Line:     sourceToken.Pos.Line,
				File:     sourceToken.Pos.Filename,
				Link:     d.Link,
			}
		}
	} else {
		return &issue.Issue{
			Detector: d.Name,
			Type:     d.IssueType,
			Message:  fmt.Sprintf("Module source \"%s\" is not pinned", module.ModuleSource),
			Line:     sourceToken.Pos.Line,
			File:     sourceToken.Pos.Filename,
			Link:     d.Link,
		}
	}

	return nil
}

func (d *TerraformModulePinnedSourceDetector) detectMercurialSource(module *schema.Module) *issue.Issue {
	lower := strings.ToLower(module.ModuleSource)
	sourceToken, _ := module.GetToken("source")

	if strings.Contains(lower, "rev=") {
		if strings.Contains(lower, "rev=default") {
			return &issue.Issue{
				Detector: d.Name,
				Type:     issue.WARNING,
				Message:  fmt.Sprintf("Module source \"%s\" uses default rev \"default\"", module.ModuleSource),
				Line:     sourceToken.Pos.Line,
				File:     sourceToken.Pos.Filename,
				Link:     d.Link,
			}
		}
	} else {
		return &issue.Issue{
			Detector: d.Name,
			Type:     issue.WARNING,
			Message:  fmt.Sprintf("Module source \"%s\" is not pinned", module.ModuleSource),
			Line:     sourceToken.Pos.Line,
			File:     sourceToken.Pos.Filename,
			Link:     d.Link,
		}
	}

	return nil
}
