package terraformrules

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform/configs"
	"github.com/terraform-linters/tflint/tflint"
)

// TerraformModuleSemverSourceRule checks the module source is semvere
type TerraformModuleSemverSourceRule struct {
	attributeName string
}

// NewTerraformModuleSemverSourceRule returns new rule with default attributes
func NewTerraformModuleSemverSourceRule() *TerraformModuleSemverSourceRule {
	return &TerraformModuleSemverSourceRule{
		attributeName: "source",
	}
}

// Name returns the rule name
func (r *TerraformModuleSemverSourceRule) Name() string {
	return "terraform_module_semver_source"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformModuleSemverSourceRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformModuleSemverSourceRule) Severity() string {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformModuleSemverSourceRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks if module source version is not semver or unpinned
func (r *TerraformModuleSemverSourceRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	for _, module := range runner.TFConfig.Module.ModuleCalls {
		log.Printf("[DEBUG] Walk `%s` attribute", module.Name+".source")

		lower := strings.ToLower(module.SourceAddr)

		if ReGitHub.MatchString(lower) || ReBitbucket.MatchString(lower) || ReGenericGit.MatchString(lower) {
			r.checkGitSemverSource(runner, module)
		} else if strings.HasPrefix(lower, "hg::") {
			r.checkMercurialSemverSource(runner, module)
		}
	}

	return nil
}

var reSemverReference = regexp.MustCompile("\\?ref=v?\\d+\\.\\d+\\.\\d+$")
var reSemverRevision = regexp.MustCompile("\\?rev=v?\\d+\\.\\d+\\.\\d+$")

func (r *TerraformModuleSemverSourceRule) checkGitSemverSource(runner *tflint.Runner, module *configs.ModuleCall) {
	lower := strings.ToLower(module.SourceAddr)

	if strings.Contains(lower, "ref=") {
		if !reSemverReference.MatchString(lower) {
			runner.EmitIssue(
				r,
				fmt.Sprintf("Module source \"%s\" uses a ref which is not a version string", module.SourceAddr),
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

func (r *TerraformModuleSemverSourceRule) checkMercurialSemverSource(runner *tflint.Runner, module *configs.ModuleCall) {
	lower := strings.ToLower(module.SourceAddr)

	if strings.Contains(lower, "rev=") {
		if !reSemverRevision.MatchString(lower) {
			runner.EmitIssue(
				r,
				fmt.Sprintf("Module source \"%s\" uses a rev which is not a version string", module.SourceAddr),
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
