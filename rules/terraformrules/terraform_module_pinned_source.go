package terraformrules

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform/configs"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/project"
	"github.com/wata727/tflint/tflint"
)

// TerraformModulePinnedSourceRule checks unpinned or default version module source
type TerraformModulePinnedSourceRule struct {
	attributeName string
}

// NewTerraformModulePinnedSourceRule returns new rule with default attributes
func NewTerraformModulePinnedSourceRule() *TerraformModulePinnedSourceRule {
	return &TerraformModulePinnedSourceRule{
		attributeName: "source",
	}
}

// Name returns the rule name
func (r *TerraformModulePinnedSourceRule) Name() string {
	return "terraform_module_pinned_source"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformModulePinnedSourceRule) Enabled() bool {
	return true
}

// Type returns the rule severity
func (r *TerraformModulePinnedSourceRule) Type() string {
	return issue.WARNING
}

// Link returns the rule reference link
func (r *TerraformModulePinnedSourceRule) Link() string {
	return project.ReferenceLink(r.Name())
}

var reGithub = regexp.MustCompile("(^github.com/(.+)/(.+)$)|(^git@github.com:(.+)/(.+)$)")
var reBitbucket = regexp.MustCompile("^bitbucket.org/(.+)/(.+)$")
var reGenericGit = regexp.MustCompile("(git://(.+)/(.+))|(git::https://(.+)/(.+))|(git::ssh://((.+)@)??(.+)/(.+)/(.+))")

// Check checks if module source version is default or unpinned
// Note that this rule is valid only for Git or Mercurial source
func (r *TerraformModulePinnedSourceRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	for _, module := range runner.TFConfig.Module.ModuleCalls {
		log.Printf("[DEBUG] Walk `%s` attribute", module.Name+".source")

		lower := strings.ToLower(module.SourceAddr)

		if reGithub.MatchString(lower) || reBitbucket.MatchString(lower) || reGenericGit.MatchString(lower) {
			r.checkGitSource(runner, module)
		} else if strings.HasPrefix(lower, "hg::") {
			r.checkMercurialSource(runner, module)
		}
	}

	return nil
}

// If the source has `ref=master` or doesn't have reference, it reports an issue for the module
func (r *TerraformModulePinnedSourceRule) checkGitSource(runner *tflint.Runner, module *configs.ModuleCall) {
	lower := strings.ToLower(module.SourceAddr)

	if strings.Contains(lower, "ref=") {
		if strings.Contains(lower, "ref=master") {
			runner.EmitIssue(
				r,
				fmt.Sprintf("Module source \"%s\" uses default ref \"master\"", module.SourceAddr),
				module.SourceAddrRange,
			)
		}
	} else {
		runner.EmitIssue(
			r,
			fmt.Sprintf("Module source \"%s\" is not pinned", module.SourceAddr),
			module.SourceAddrRange,
		)
	}
}

// If the source has `rev=default` or doesn't have reference, it reports an issue for the module
func (r *TerraformModulePinnedSourceRule) checkMercurialSource(runner *tflint.Runner, module *configs.ModuleCall) {
	lower := strings.ToLower(module.SourceAddr)

	if strings.Contains(lower, "rev=") {
		if strings.Contains(lower, "rev=default") {
			runner.EmitIssue(
				r,
				fmt.Sprintf("Module source \"%s\" uses default rev \"default\"", module.SourceAddr),
				module.SourceAddrRange,
			)
		}
	} else {
		runner.EmitIssue(
			r,
			fmt.Sprintf("Module source \"%s\" is not pinned", module.SourceAddr),
			module.SourceAddrRange,
		)
	}
}
