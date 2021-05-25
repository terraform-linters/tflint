package terraformrules

import (
	"fmt"
	"log"
	"net/url"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/go-getter"

	"github.com/hashicorp/terraform/configs"
	"github.com/terraform-linters/tflint/tflint"
)

// TerraformModulePinnedSourceRule checks unpinned or default version module source
type TerraformModulePinnedSourceRule struct {
	attributeName string
}

type terraformModulePinnedSourceRuleConfig struct {
	Style string `hcl:"style,optional"`
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

// Severity returns the rule severity
func (r *TerraformModulePinnedSourceRule) Severity() string {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformModulePinnedSourceRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks if module source version is pinned
// Note that this rule is valid only for Git or Mercurial source
func (r *TerraformModulePinnedSourceRule) Check(runner *tflint.Runner) error {
	if !runner.TFConfig.Path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	config := terraformModulePinnedSourceRuleConfig{Style: "flexible"}
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

func (r *TerraformModulePinnedSourceRule) checkModule(runner *tflint.Runner, module *configs.ModuleCall, config terraformModulePinnedSourceRuleConfig) error {
	log.Printf("[DEBUG] Walk `%s` attribute", module.Name+".source")

	source, err := getter.Detect(module.SourceAddr, "", getter.Detectors)
	if err != nil {
		return err
	}

	u, err := url.Parse(source)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "git", "hg":
	default:
		return nil
	}

	query := u.Query()

	if ref := query.Get("ref"); ref != "" {
		return r.checkRevision(runner, module, config, "ref", ref)
	}

	if rev := query.Get("rev"); rev != "" {
		return r.checkRevision(runner, module, config, "rev", rev)
	}

	runner.EmitIssue(
		r,
		fmt.Sprintf(`Module source "%s" is not pinned`, module.SourceAddr),
		module.SourceAddrRange,
	)

	return nil
}

func (r *TerraformModulePinnedSourceRule) checkRevision(runner *tflint.Runner, module *configs.ModuleCall, config terraformModulePinnedSourceRuleConfig, key string, value string) error {
	switch config.Style {
	// The "flexible" style enforces to pin source, except for the default branch
	case "flexible":
		if key == "ref" && value == "master" {
			runner.EmitIssue(
				r,
				fmt.Sprintf("Module source \"%s\" uses default %s \"master\"", module.SourceAddr, key),
				module.SourceAddrRange,
			)
		}

		if key == "rev" && value == "default" {
			runner.EmitIssue(
				r,
				fmt.Sprintf("Module source \"%s\" uses default %s \"default\"", module.SourceAddr, key),
				module.SourceAddrRange,
			)
		}
	// The "semver" style enforces to pin source like semantic versioning
	case "semver":
		_, err := semver.NewVersion(value)
		if err != nil {
			runner.EmitIssue(
				r,
				fmt.Sprintf("Module source \"%s\" uses a %s which is not a version string", module.SourceAddr, key),
				module.SourceAddrRange,
			)
		}
	default:
		return fmt.Errorf("`%s` is invalid style", config.Style)
	}

	return nil
}
