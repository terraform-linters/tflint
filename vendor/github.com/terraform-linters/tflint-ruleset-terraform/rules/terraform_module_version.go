package rules

import (
	"fmt"
	"regexp"

	tfaddr "github.com/hashicorp/terraform-registry-address"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
	"github.com/terraform-linters/tflint-ruleset-terraform/terraform"
)

// SemVer regexp with optional leading =
// https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
var exactVersionRegexp = regexp.MustCompile(`^=?\s*` + `(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

// TerraformModuleVersionRule checks that Terraform modules sourced from a registry specify a version
type TerraformModuleVersionRule struct {
	tflint.DefaultRule
}

// TerraformModuleVersionRuleConfig is the config structure for the TerraformModuleVersionRule rule
type TerraformModuleVersionRuleConfig struct {
	Exact bool `hclext:"exact,optional"`
}

// NewTerraformModuleVersionRule returns a new rule
func NewTerraformModuleVersionRule() *TerraformModuleVersionRule {
	return &TerraformModuleVersionRule{}
}

// Name returns the rule name
func (r *TerraformModuleVersionRule) Name() string {
	return "terraform_module_version"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformModuleVersionRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *TerraformModuleVersionRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformModuleVersionRule) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check checks whether module source attributes resolve to a Terraform registry
// If they do, it checks a version (or range) is set
func (r *TerraformModuleVersionRule) Check(rr tflint.Runner) error {
	runner := rr.(*terraform.Runner)

	path, err := runner.GetModulePath()
	if err != nil {
		return err
	}
	if !path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	config := TerraformModuleVersionRuleConfig{}
	if err := runner.DecodeRuleConfig(r.Name(), &config); err != nil {
		return err
	}

	calls, diags := runner.GetModuleCalls()
	if diags.HasErrors() {
		return diags
	}

	for _, call := range calls {
		if err := r.checkModule(runner, call, config); err != nil {
			return err
		}
	}

	return nil
}

func (r *TerraformModuleVersionRule) checkModule(runner tflint.Runner, module *terraform.ModuleCall, config TerraformModuleVersionRuleConfig) error {
	_, err := tfaddr.ParseModuleSource(module.Source)
	if err != nil {
		// If parsing fails, the source does not expect to specify a version,
		// such as local or remote. So instead of returning an error,
		// it returns nil to stop the check.
		return nil
	}

	return r.checkVersion(runner, module, config)
}

func (r *TerraformModuleVersionRule) checkVersion(runner tflint.Runner, module *terraform.ModuleCall, config TerraformModuleVersionRuleConfig) error {
	if module.Version == nil {
		return runner.EmitIssue(
			r,
			fmt.Sprintf("module %q should specify a version", module.Name),
			module.DefRange,
		)
	}

	if !config.Exact {
		return nil
	}

	if len(module.Version) > 1 {
		return runner.EmitIssue(
			r,
			fmt.Sprintf("module %q should specify an exact version, but multiple constraints were found", module.Name),
			module.VersionAttr.Range,
		)
	}

	if !exactVersionRegexp.MatchString(module.Version[0].String()) {
		return runner.EmitIssue(
			r,
			fmt.Sprintf("module %q should specify an exact version, but a range was found", module.Name),
			module.VersionAttr.Range,
		)
	}

	return nil
}
