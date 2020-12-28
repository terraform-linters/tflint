package rules

import "github.com/terraform-linters/tflint/tflint"

// RuleSet is a pseudo RuleSet to handle core rules like plugin
type RuleSet struct{}

// RuleSetName is the name of the rule set.
func (r *RuleSet) RuleSetName() (string, error) {
	return "Core", nil
}

// RuleSetVersion is the version of the plugin.
func (r *RuleSet) RuleSetVersion() (string, error) {
	return tflint.Version, nil
}

// RuleNames is a list of rule names provided by the plugin.
func (r *RuleSet) RuleNames() ([]string, error) {
	names := []string{}
	for _, rule := range DefaultRules {
		names = append(names, rule.Name())
	}
	return names, nil
}
