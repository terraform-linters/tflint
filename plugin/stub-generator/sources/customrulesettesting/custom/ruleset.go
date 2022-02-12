package custom

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type RuleSet struct {
	tflint.BuiltinRuleSet
	config *Config
}

func (r *RuleSet) ApplyConfig(config *tflint.Config) error {
	r.ApplyCommonConfig(config)

	cfg := Config{}
	diags := gohcl.DecodeBody(config.Body, nil, &cfg)
	if diags.HasErrors() {
		return diags
	}
	r.config = &cfg

	return nil
}

func (r *RuleSet) Check(rr tflint.Runner) error {
	runner, err := NewRunner(rr, r.config)
	if err != nil {
		return err
	}

	for _, rule := range r.EnabledRules {
		if err := rule.Check(runner); err != nil {
			return fmt.Errorf("Failed to check `%s` rule: %s", rule.Name(), err)
		}
	}
	return nil
}
