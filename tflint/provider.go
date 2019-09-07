package tflint

import (
	"log"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configschema"
)

// ProviderConfig represents a provider block with an eval context (runner)
type ProviderConfig struct {
	tfProvider *configs.Provider
	runner     *Runner
	attributes hcl.Attributes
	blocks     hcl.Blocks
}

// NewProviderConfig returns a provider config from the given `configs.Provider` and runner
func NewProviderConfig(tfProvider *configs.Provider, runner *Runner, schema *hcl.BodySchema) (*ProviderConfig, error) {
	providerConfig := &ProviderConfig{
		tfProvider: tfProvider,
		runner:     runner,
		attributes: map[string]*hcl.Attribute{},
		blocks:     []*hcl.Block{},
	}

	if tfProvider != nil {
		content, _, diags := tfProvider.Config.PartialContent(schema)
		if diags.HasErrors() {
			return nil, diags
		}

		providerConfig.attributes = content.Attributes
		providerConfig.blocks = content.Blocks
	}

	return providerConfig, nil
}

// Get returns a value corresponding to the given key
// It should be noted that the value is evaluated if it is evaluable
// The second return value is a flag that determines whether a value exists
// We assume the provider has only simple attributes, so it just returns string
func (p *ProviderConfig) Get(key string) (string, bool, error) {
	attribute, exists := p.attributes[key]
	if !exists {
		log.Printf("[INFO] `%s` is not found in the provider block.", key)
		return "", false, nil
	}

	var val string
	err := p.runner.EvaluateExpr(attribute.Expr, &val)

	err = p.runner.EnsureNoError(err, func() error { return nil })
	if err != nil {
		return "", true, err
	}
	return val, true, nil
}

// GetBlock is Get for blocks. Obviously from the return type,
// nested blocks and blocks with complex types are not supported.
// Also note that all attributes are lost if the given block
// contains unevalable expressions because the entire block is evaluated.
func (p *ProviderConfig) GetBlock(key string, schema *configschema.Block) (map[string]string, bool, error) {
	val := map[string]string{}

	var ret *hcl.Block
	for _, block := range p.blocks {
		if block.Type == key {
			ret = block
		}
	}
	if ret == nil {
		log.Printf("[INFO] `%s` is not found in the provider block.", key)
		return val, false, nil
	}

	err := p.runner.EvaluateBlock(ret, schema, &val)

	err = p.runner.EnsureNoError(err, func() error { return nil })
	if err != nil {
		return val, true, err
	}
	return val, true, nil
}
