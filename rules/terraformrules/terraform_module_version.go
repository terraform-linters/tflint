package terraformrules

import (
	"fmt"
	"log"
	"regexp"

	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/terraform-linters/tflint/terraform/configs"
	"github.com/terraform-linters/tflint/tflint"
)

// SemVer regexp with optional leading =
// https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
var exactVersionRegexp = regexp.MustCompile(`^=?\s*` + `(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

// TerraformModuleVersionRule checks that Terraform modules sourced from a registry specify a version
type TerraformModuleVersionRule struct{}

// TerraformModuleVersionRuleConfig is the config structure for the TerraformModuleVersionRule rule
type TerraformModuleVersionRuleConfig struct {
	Exact bool `hcl:"exact,optional"`
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
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether module source attributes resolve to a Terraform registry
// If they do, it checks a version (or range) is set
func (r *TerraformModuleVersionRule) Check(runner *tflint.Runner) error {
	if !runner.TFConfig.Path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	config := TerraformModuleVersionRuleConfig{}
	if err := runner.DecodeRuleConfig(r.Name(), &config); err != nil {
		return err
	}

	for _, module := range runner.TFConfig.Module.ModuleCalls {
		if err := r.checkModule(runner, module, config); err != nil {
			return err
		}
	}

	return nil
}

func (r *TerraformModuleVersionRule) checkModule(runner *tflint.Runner, module *configs.ModuleCall, config TerraformModuleVersionRuleConfig) error {
	log.Printf("[DEBUG] Walk `%s` attribute", module.Name+".source")

	source, err := addrs.ParseModuleSource(module.SourceAddrRaw)
	if err != nil {
		return err
	}

	switch source.(type) {
	case addrs.ModuleSourceRegistry:
		return r.checkVersion(runner, module, config)
	}

	return nil
}

func (r *TerraformModuleVersionRule) checkVersion(runner *tflint.Runner, module *configs.ModuleCall, config TerraformModuleVersionRuleConfig) error {
	if module.Version.Required == nil {
		runner.EmitIssue(
			r,
			fmt.Sprintf("module %q should specify a version", module.Name),
			module.DeclRange,
		)

		return nil
	}

	if !config.Exact {
		return nil
	}

	if len(module.Version.Required) > 1 {
		runner.EmitIssue(
			r,
			fmt.Sprintf("module %q should specify an exact version, but multiple constraints were found", module.Name),
			module.Version.DeclRange,
		)

		return nil
	}

	if !exactVersionRegexp.MatchString(module.Version.Required[0].String()) {
		runner.EmitIssue(
			r,
			fmt.Sprintf("module %q should specify an exact version, but a range was found", module.Name),
			module.Version.DeclRange,
		)
	}

	return nil
}
