package terraform

import (
	"fmt"
	"strings"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/logger"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// RuleSet is the custom ruleset for the Terraform Language.
type RuleSet struct {
	tflint.BuiltinRuleSet

	PresetRules map[string][]tflint.Rule

	globalConfig  *tflint.Config
	rulesetConfig *Config
}

func (r *RuleSet) ConfigSchema() *hclext.BodySchema {
	r.rulesetConfig = &Config{}
	return hclext.ImpliedBodySchema(r.rulesetConfig)
}

// ApplyGlobalConfig is normally not expected to be overridden,
// but here the preset setting takes precedence over DisabledByDefault,
// so it just overrides and saves the global config to the ruleset.
func (r *RuleSet) ApplyGlobalConfig(config *tflint.Config) error {
	r.globalConfig = config
	return nil
}

// ApplyConfig controls rule activation based on global and preset configs.
// The priority of rules is in the following order:
//
//  1. Rule config declared in each "rule" block
//  2. Preset config declared in "plugin" block
//  3. The `disabled_by_default` declared in global "config" block
//
// Individual rule configs always take precedence over anything else.
// Preset rules are then prioritized. For example, if `disabled_by_default = true`
// and `preset = "recommended"` is declared, all recommended rules will be enabled.
func (r *RuleSet) ApplyConfig(body *hclext.BodyContent) error {
	diags := hclext.DecodeBody(body, nil, r.rulesetConfig)
	if diags.HasErrors() {
		return diags
	}

	if r.globalConfig.DisabledByDefault {
		logger.Debug("Only mode is enabled. Ignoring default plugin rules")
	}

	// Default preset is "all"
	rules := r.PresetRules["all"]
	_, presetExists := body.Attributes["preset"]
	if presetExists {
		presetRules, exists := r.PresetRules[r.rulesetConfig.Preset]
		if !exists {
			validPresets := []string{}
			for name := range r.PresetRules {
				validPresets = append(validPresets, name)
			}
			return fmt.Errorf(`preset "%s" is not found. Valid presets are %s`, r.rulesetConfig.Preset, strings.Join(validPresets, ", "))
		}
		rules = presetRules
	}

	r.EnabledRules = []tflint.Rule{}
	for _, rule := range rules {
		enabled := rule.Enabled()
		if cfg := r.globalConfig.Rules[rule.Name()]; cfg != nil {
			enabled = cfg.Enabled
		} else if r.globalConfig.DisabledByDefault && !presetExists {
			// Preset takes precedence over DisabledByDefault
			enabled = false
		}

		if enabled {
			r.EnabledRules = append(r.EnabledRules, rule)
		}
	}

	return nil
}

// Check runs inspection for each rule by applying Runner.
func (r *RuleSet) Check(rr tflint.Runner) error {
	runner := NewRunner(rr)

	for _, rule := range r.EnabledRules {
		if err := rule.Check(runner); err != nil {
			return fmt.Errorf("Failed to check `%s` rule: %s", rule.Name(), err)
		}
	}
	return nil
}
