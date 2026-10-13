package plugin

import (
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	sdkplugin "github.com/terraform-linters/tflint-plugin-sdk/plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/plugin2host"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
	"github.com/terraform-linters/tflint-ruleset-terraform/rules"
	"github.com/terraform-linters/tflint-ruleset-terraform/terraform"
	"github.com/terraform-linters/tflint/internal/sdkfixer"
)

type localRuleSet struct {
	impl *terraform.RuleSet
	fix  bool
}

func newLocalRuleSet() *localRuleSet {
	return &localRuleSet{impl: &terraform.RuleSet{
		BuiltinRuleSet: tflint.BuiltinRuleSet{Name: "terraform", Version: fmt.Sprintf("%s-bundled", project.Version)},
		PresetRules:    freshPresetRules(),
	}}
}

func (r *localRuleSet) RuleSetName() (string, error)    { return r.impl.RuleSetName(), nil }
func (r *localRuleSet) RuleSetVersion() (string, error) { return r.impl.RuleSetVersion(), nil }
func (r *localRuleSet) RuleNames() ([]string, error)    { return r.impl.RuleNames(), nil }
func (r *localRuleSet) VersionConstraints() (version.Constraints, error) {
	if r.impl.VersionConstraint() == "" {
		return version.NewConstraint(">= 0.46")
	}
	return version.NewConstraint(r.impl.VersionConstraint() + ", >= 0.46")
}
func (r *localRuleSet) SDKVersion() (*version.Version, error) {
	return version.NewVersion(sdkplugin.SDKVersion)
}
func (r *localRuleSet) ConfigSchema() (*hclext.BodySchema, error) { return r.impl.ConfigSchema(), nil }
func (r *localRuleSet) ApplyGlobalConfig(config *tflint.Config) error {
	r.fix = config.Fix
	return r.impl.ApplyGlobalConfig(config)
}
func (r *localRuleSet) ApplyConfig(content *hclext.BodyContent, _ map[string][]byte) error {
	return r.impl.ApplyConfig(content)
}
func (r *localRuleSet) Check(server plugin2host.Server) error {
	local := &localRunner{server: server, Fixer: sdkfixer.NewFixer(server.GetFiles(tflint.SelfModuleCtxType)), FixEnabled: r.fix}
	runner, err := r.impl.NewRunner(local)
	if err != nil {
		return err
	}
	for _, rule := range r.impl.EnabledRules {
		if err := rule.Check(runner); err != nil {
			return fmt.Errorf("failed to check %q rule: %s", rule.Name(), err)
		}
		if local.Fixer.HasChanges() {
			local.Fixer.FormatChanges()
			if err := server.ApplyChanges(local.Fixer.Changes()); err != nil {
				return fmt.Errorf("failed to apply fixes by %q rule: %s", rule.Name(), err)
			}
			local.Fixer.ApplyChanges()
		}
	}
	return nil
}

// Keep rule instances private to an inspection, matching plugin process isolation.
func freshPresetRules() map[string][]tflint.Rule {
	return map[string][]tflint.Rule{
		"all": {
			rules.NewTerraformCommentSyntaxRule(),
			rules.NewTerraformDeprecatedIndexRule(),
			rules.NewTerraformDeprecatedInterpolationRule(),
			rules.NewTerraformDeprecatedLookupRule(),
			rules.NewTerraformDocumentedOutputsRule(),
			rules.NewTerraformDocumentedVariablesRule(),
			rules.NewTerraformEmptyListEqualityRule(),
			rules.NewTerraformJSONSyntaxRule(),
			rules.NewTerraformMapDuplicateKeysRule(),
			rules.NewTerraformModulePinnedSourceRule(),
			rules.NewTerraformModuleShallowCloneRule(),
			rules.NewTerraformModuleVersionRule(),
			rules.NewTerraformNamingConventionRule(),
			rules.NewTerraformRequiredProvidersRule(),
			rules.NewTerraformRequiredVersionRule(),
			rules.NewTerraformStandardModuleStructureRule(),
			rules.NewTerraformTypedVariablesRule(),
			rules.NewTerraformUnusedDeclarationsRule(),
			rules.NewTerraformUnusedRequiredProvidersRule(),
			rules.NewTerraformWorkspaceRemoteRule(),
		},
		"recommended": {
			rules.NewTerraformDeprecatedIndexRule(),
			rules.NewTerraformDeprecatedInterpolationRule(),
			rules.NewTerraformDeprecatedLookupRule(),
			rules.NewTerraformEmptyListEqualityRule(),
			rules.NewTerraformJSONSyntaxRule(),
			rules.NewTerraformMapDuplicateKeysRule(),
			rules.NewTerraformModulePinnedSourceRule(),
			rules.NewTerraformModuleVersionRule(),
			rules.NewTerraformRequiredProvidersRule(),
			rules.NewTerraformRequiredVersionRule(),
			rules.NewTerraformTypedVariablesRule(),
			rules.NewTerraformUnusedDeclarationsRule(),
			rules.NewTerraformWorkspaceRemoteRule(),
		},
	}
}
