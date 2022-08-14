package rules

import (
	"fmt"
	"regexp"

	tfaddr "github.com/hashicorp/terraform-registry-address"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
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
func (r *TerraformModuleVersionRule) Check(runner tflint.Runner) error {
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

	body, err := runner.GetModuleContent(moduleCallSchema, &tflint.GetModuleContentOption{IncludeNotCreated: true})
	if err != nil {
		return err
	}

	for _, block := range body.Blocks {
		module, diags := decodeModuleCall(block)
		if diags.HasErrors() {
			return diags
		}

		if err := r.checkModule(runner, module, config); err != nil {
			return err
		}
	}

	return nil
}

func (r *TerraformModuleVersionRule) checkModule(runner tflint.Runner, module *moduleCall, config TerraformModuleVersionRuleConfig) error {
	_, err := tfaddr.ParseModuleSource(module.source)
	if err != nil {
		// If parsing fails, the source does not expect to specify a version,
		// such as local or remote. So instead of returning an error,
		// it returns nil to stop the check.
		return nil
	}

	return r.checkVersion(runner, module, config)
}

func (r *TerraformModuleVersionRule) checkVersion(runner tflint.Runner, module *moduleCall, config TerraformModuleVersionRuleConfig) error {
	if module.version == nil {
		return runner.EmitIssue(
			r,
			fmt.Sprintf("module %q should specify a version", module.name),
			module.defRange,
		)
	}

	if !config.Exact {
		return nil
	}

	if len(module.version) > 1 {
		return runner.EmitIssue(
			r,
			fmt.Sprintf("module %q should specify an exact version, but multiple constraints were found", module.name),
			module.versionAttr.Range,
		)
	}

	if !exactVersionRegexp.MatchString(module.version[0].String()) {
		return runner.EmitIssue(
			r,
			fmt.Sprintf("module %q should specify an exact version, but a range was found", module.name),
			module.versionAttr.Range,
		)
	}

	return nil
}
