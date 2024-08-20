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

func (r *RuleSet) RuleNames() []string {
	names := make([]string, len(r.PresetRules["all"]))
	for idx, rule := range r.PresetRules["all"] {
		names[idx] = rule.Name()
	}
	return names
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
//  1. --only option
//  2. Rule config declared in each "rule" block
//  3. Preset config declared in "plugin" block
//  4. The `disabled_by_default` declared in global "config" block
//
// Individual rule configs always take precedence over anything else.
// Preset rules are then prioritized. For example, if `disabled_by_default = true`
// and `preset = "recommended"` is declared, all recommended rules will be enabled.
func (r *RuleSet) ApplyConfig(body *hclext.BodyContent) error {
	diags := hclext.DecodeBody(body, nil, r.rulesetConfig)
	if diags.HasErrors() {
		return diags
	}

	only := map[string]bool{}
	if len(r.globalConfig.Only) > 0 {
		logger.Debug("Only mode is enabled. Ignoring default plugin rules")
		for _, rule := range r.globalConfig.Only {
			only[rule] = true
		}
	} else if r.globalConfig.DisabledByDefault {
		logger.Debug("Default plugin rules are disabled by default")
	}

	preset := map[string]bool{}
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
		for _, rule := range presetRules {
			preset[rule.Name()] = true
		}
	}

	r.EnabledRules = []tflint.Rule{}
	for _, rule := range r.PresetRules["all"] {
		enabled := rule.Enabled()
		if len(only) > 0 {
			enabled = only[rule.Name()]
		} else if cfg := r.globalConfig.Rules[rule.Name()]; cfg != nil {
			enabled = cfg.Enabled
		} else if presetExists {
			// Ignore rules not in preset
			if !preset[rule.Name()] {
				enabled = false
			}
		} else if r.globalConfig.DisabledByDefault {
			enabled = false
		}

		if enabled {
			r.EnabledRules = append(r.EnabledRules, rule)
		}
	}

	return nil
}

// NewRunner injects a custom runner
func (r *RuleSet) NewRunner(runner tflint.Runner) (tflint.Runner, error) {
	return NewRunner(runner), nil
}
