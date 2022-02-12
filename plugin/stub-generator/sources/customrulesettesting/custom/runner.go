package custom

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type Runner struct {
	tflint.Runner
	CustomConfig *Config
}

func NewRunner(runner tflint.Runner, config *Config) (*Runner, error) {
	provider, err := runner.RootProvider("custom")
	if err != nil {
		return nil, err
	}

	if provider != nil {
		content, _, diags := provider.Config.PartialContent(&hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{
				{Name: "zone"},
			},
			Blocks: []hcl.BlockHeaderSchema{
				{Type: "annotation"},
			},
		})
		if diags.HasErrors() {
			return nil, diags
		}

		if attr, exists := content.Attributes["zone"]; exists {
			var zone string
			err := runner.EvaluateExprOnRootCtx(attr.Expr, &zone, nil)
			err = runner.EnsureNoError(err, func() error {
				config.Zone = zone
				return nil
			})
			if err != nil {
				return nil, err
			}
		}

		for _, block := range content.Blocks {
			content, _, diags := block.Body.PartialContent(&hcl.BodySchema{
				Attributes: []hcl.AttributeSchema{
					{Name: "value"},
				},
			})
			if diags.HasErrors() {
				return nil, diags
			}

			if attr, exists := content.Attributes["value"]; exists {
				var val string
				err := runner.EvaluateExprOnRootCtx(attr.Expr, &val, nil)
				err = runner.EnsureNoError(err, func() error {
					config.Annotation = val
					return nil
				})
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return &Runner{
		Runner:       runner,
		CustomConfig: config,
	}, nil
}
