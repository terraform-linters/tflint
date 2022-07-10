package terraformrules

import (
	"fmt"
	"log"

	tfaddr "github.com/hashicorp/terraform-registry-address"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/tflint"
	"github.com/zclconf/go-cty/cty"
)

// TerraformRequiredProvidersRule checks whether Terraform sets version constraints for all configured providers
type TerraformRequiredProvidersRule struct{}

// NewTerraformRequiredProvidersRule returns new rule with default attributes
func NewTerraformRequiredProvidersRule() *TerraformRequiredProvidersRule {
	return &TerraformRequiredProvidersRule{}
}

// Name returns the rule name
func (r *TerraformRequiredProvidersRule) Name() string {
	return "terraform_required_providers"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformRequiredProvidersRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformRequiredProvidersRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformRequiredProvidersRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check Checks whether provider required version is set
func (r *TerraformRequiredProvidersRule) Check(runner *tflint.Runner) error {
	if !runner.TFConfig.Path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	body, diags := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "provider",
				LabelNames: []string{"name"},
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{
						{Name: "version"},
					},
				},
			},
		},
	}, sdk.GetModuleContentOption{})
	if diags.HasErrors() {
		return diags
	}

	for _, provider := range body.Blocks {
		if _, exists := provider.Body.Attributes["version"]; exists {
			runner.EmitIssue(
				r,
				`provider version constraint should be specified via "required_providers"`,
				provider.DefRange,
			)
		}
	}

	providerRefs, diags := getProviderRefs(runner)
	if diags.HasErrors() {
		return diags
	}

	requiredProvidersSchema := []hclext.AttributeSchema{}
	for name := range providerRefs {
		requiredProvidersSchema = append(requiredProvidersSchema, hclext.AttributeSchema{Name: name})
	}

	body, diags = runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type: "terraform",
				Body: &hclext.BodySchema{
					Blocks: []hclext.BlockSchema{
						{
							Type: "required_providers",
							Body: &hclext.BodySchema{
								Attributes: requiredProvidersSchema,
							},
						},
					},
				},
			},
		},
	}, sdk.GetModuleContentOption{})
	if diags.HasErrors() {
		return diags
	}

	requiredProviders := hclext.Attributes{}
	for _, terraform := range body.Blocks {
		for _, requiredProvidersBlock := range terraform.Body.Blocks {
			requiredProviders = requiredProvidersBlock.Body.Attributes
		}
	}

	for name, ref := range providerRefs {
		if name == "terraform" {
			// "terraform" provider is a builtin provider
			// @see https://github.com/hashicorp/terraform/blob/v1.2.5/internal/addrs/provider.go#L106-L112
			continue
		}

		provider, exists := requiredProviders[name]
		if !exists {
			runner.EmitIssue(r, fmt.Sprintf(`Missing version constraint for provider "%s" in "required_providers"`, name), ref.defRange)
			continue
		}

		val, diags := provider.Expr.Value(nil)
		if diags.HasErrors() {
			return diags
		}
		// Look for a single static string, in case we have the legacy version-only
		// format in the configuration.
		if val.Type() == cty.String {
			continue
		}

		vm := val.AsValueMap()
		if _, exists := vm["version"]; !exists {
			if source, exists := vm["source"]; exists {
				p, err := tfaddr.ParseProviderSource(source.AsString())
				if err != nil {
					return err
				}

				if p.IsBuiltIn() {
					continue
				}
			}
			runner.EmitIssue(r, fmt.Sprintf(`Missing version constraint for provider "%s" in "required_providers"`, name), ref.defRange)
		}
	}

	return nil
}
