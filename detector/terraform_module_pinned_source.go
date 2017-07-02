package detector

import (
	"fmt"
	"strings"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type TerraformModulePinnedSourceDetector struct {
	*Detector
	IssueType  string
	TargetType string
	DeepCheck  bool
}

func (d *Detector) CreateTerraformModulePinnedSourceDetector() *TerraformModulePinnedSourceDetector {
	return &TerraformModulePinnedSourceDetector{
		Detector:   d,
		IssueType:  issue.WARNING,
		TargetType: "module",
		DeepCheck:  false,
	}
}

func (d *TerraformModulePinnedSourceDetector) Detect(module *schema.Module, issues *[]*issue.Issue) {
	lower := strings.ToLower(module.ModuleSource)

	if strings.Contains(lower, "git") || strings.Contains(lower, "bitbucket") {
		if issue := d.detectGitSource(module); issue != nil {
			*issues = append(*issues, issue)
		}
	} else if strings.HasPrefix(lower, "hg:") {
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
				Type:    d.IssueType,
				Message: fmt.Sprintf("Module source \"%s\" uses default ref \"master\"", module.ModuleSource),
				Line:    sourceToken.Pos.Line,
				File:    sourceToken.Pos.Filename,
			}
		}
	} else {
		return &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("Module source \"%s\" is not pinned", module.ModuleSource),
			Line:    sourceToken.Pos.Line,
			File:    sourceToken.Pos.Filename,
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
				Type:    issue.WARNING,
				Message: fmt.Sprintf("Module source \"%s\" uses default rev \"default\"", module.ModuleSource),
				Line:    sourceToken.Pos.Line,
				File:    sourceToken.Pos.Filename,
			}
		}
	} else {
		return &issue.Issue{
			Type:    issue.WARNING,
			Message: fmt.Sprintf("Module source \"%s\" is not pinned", module.ModuleSource),
			Line:    sourceToken.Pos.Line,
			File:    sourceToken.Pos.Filename,
		}
	}

	return nil
}
