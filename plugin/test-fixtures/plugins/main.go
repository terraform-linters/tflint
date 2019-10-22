package main

import (
	"github.com/wata727/tflint/plugin"
	"github.com/wata727/tflint/tflint"
)

// PluginRule is a example rule for testing
type PluginRule struct{}

// Name is the rule name
func (r *PluginRule) Name() string {
	return "plugin_rule"
}

// Enabled indicates whether the rule is enabled by default
func (r *PluginRule) Enabled() bool {
	return true
}

// Severity is a severity of the rule
func (r *PluginRule) Severity() string {
	return tflint.NOTICE
}

// Link returns the documentation URL of the rule
func (r *PluginRule) Link() string {
	return ""
}

// Check is a core process of the rule
func (r *PluginRule) Check(runner *tflint.Runner) error {
	return nil
}

// Name is the plugin name
func Name() string {
	return "plugin_rule"
}

// Version is the plugin version
func Version() string {
	return "0.0.1"
}

// NewRules returns plugin rules
func NewRules() []plugin.Rule {
	return []plugin.Rule{&PluginRule{}}
}
