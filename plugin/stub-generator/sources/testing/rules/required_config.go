package rules

import (
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// RequredConfigRule checks whether ...
type RequredConfigRule struct{}

type ruleConfig struct {
	Options []string `hcl:"options"`
}

// NewRequredConfigRule returns a new rule
func NewRequredConfigRule() *RequredConfigRule {
	return &RequredConfigRule{}
}

// Name returns the rule name
func (r *RequredConfigRule) Name() string {
	return "required_config"
}

// Enabled returns whether the rule is enabled by default
func (r *RequredConfigRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *RequredConfigRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *RequredConfigRule) Link() string {
	return ""
}

// Check checks whether ...
func (r *RequredConfigRule) Check(runner tflint.Runner) error {
	config := ruleConfig{}
	return runner.DecodeRuleConfig(r.Name(), &config)
}
