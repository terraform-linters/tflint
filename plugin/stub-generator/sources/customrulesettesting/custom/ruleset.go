package custom

import (
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type RuleSet struct {
	tflint.BuiltinRuleSet
	config *Config
}

func (r *RuleSet) ConfigSchema() *hclext.BodySchema {
	r.config = &Config{}
	return hclext.ImpliedBodySchema(r.config)
}

func (r *RuleSet) ApplyConfig(body *hclext.BodyContent) error {
	diags := hclext.DecodeBody(body, nil, r.config)
	if diags.HasErrors() {
		return diags
	}

	return nil
}

func (r *RuleSet) NewRunner(runner tflint.Runner) (tflint.Runner, error) {
	return NewRunner(runner, r.config)
}
