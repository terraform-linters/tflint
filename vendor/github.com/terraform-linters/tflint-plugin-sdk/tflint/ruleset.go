package tflint

import (
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/logger"
)

var _ RuleSet = &BuiltinRuleSet{}

// BuiltinRuleSet is the basis of the ruleset. Plugins can serve this ruleset directly.
// You can serve a custom ruleset by embedding this ruleset if you need special extensions.
type BuiltinRuleSet struct {
	Name       string
	Version    string
	Constraint string
	Rules      []Rule

	EnabledRules []Rule
}

// RuleSetName is the name of the ruleset.
// Generally, this is synonymous with the name of the plugin.
func (r *BuiltinRuleSet) RuleSetName() string {
	return r.Name
}

// RuleSetVersion is the version of the plugin.
func (r *BuiltinRuleSet) RuleSetVersion() string {
	return r.Version
}

// RuleNames is a list of rule names provided by the plugin.
func (r *BuiltinRuleSet) RuleNames() []string {
	names := make([]string, len(r.Rules))
	for idx, rule := range r.Rules {
		names[idx] = rule.Name()
	}
	return names
}

// VersionConstraint declares the version of TFLint the plugin will work with.
// Default is no constraint.
func (r *BuiltinRuleSet) VersionConstraint() string {
	return r.Constraint
}

// ApplyGlobalConfig applies the common config to the ruleset.
// This is not supposed to be overridden from custom rulesets.
// Override the ApplyConfig if you want to apply the plugin's own configuration.
//
// The priority of rule configs is as follows:
//
// 1. --only option
// 2. Rule config declared in each "rule" block
// 3. The `disabled_by_default` declared in global "config" block
func (r *BuiltinRuleSet) ApplyGlobalConfig(config *Config) error {
	r.EnabledRules = []Rule{}
	only := map[string]bool{}

	if len(config.Only) > 0 {
		logger.Debug("Only mode is enabled. Ignoring default plugin rules")
		for _, rule := range config.Only {
			only[rule] = true
		}
	} else if config.DisabledByDefault {
		logger.Debug("Default plugin rules are disabled by default")
	}

	for _, rule := range r.Rules {
		enabled := rule.Enabled()
		if len(only) > 0 {
			enabled = only[rule.Name()]
		} else if cfg := config.Rules[rule.Name()]; cfg != nil {
			enabled = cfg.Enabled
		} else if config.DisabledByDefault {
			enabled = false
		}

		if enabled {
			r.EnabledRules = append(r.EnabledRules, rule)
		}
	}
	return nil
}

// ConfigSchema returns the ruleset plugin config schema.
// This schema should be a schema inside of "plugin" block.
// Custom rulesets can override this method to return the plugin's own config schema.
func (r *BuiltinRuleSet) ConfigSchema() *hclext.BodySchema {
	return nil
}

// ApplyConfig applies the configuration to the ruleset.
// Custom rulesets can override this method to reflect the plugin's own configuration.
func (r *BuiltinRuleSet) ApplyConfig(content *hclext.BodyContent) error {
	return nil
}

// NewRunner returns a new runner based on the original runner.
// Custom rulesets can override this method to inject a custom runner.
func (r *BuiltinRuleSet) NewRunner(runner Runner) (Runner, error) {
	return runner, nil
}

// BuiltinImpl returns the receiver itself as BuiltinRuleSet.
// This is not supposed to be overridden from custom rulesets.
func (r *BuiltinRuleSet) BuiltinImpl() *BuiltinRuleSet {
	return r
}

func (r *BuiltinRuleSet) mustEmbedBuiltinRuleSet() {}
