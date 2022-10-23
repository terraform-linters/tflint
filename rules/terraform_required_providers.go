package rules

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	tfaddr "github.com/hashicorp/terraform-registry-address"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
	"github.com/terraform-linters/tflint-ruleset-terraform/terraform"
	"github.com/zclconf/go-cty/cty"
)

// TerraformRequiredProvidersRule checks whether Terraform sets version constraints for all configured providers
type TerraformRequiredProvidersRule struct {
	tflint.DefaultRule
}

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
	return true
}

// Severity returns the rule severity
func (r *TerraformRequiredProvidersRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformRequiredProvidersRule) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check Checks whether provider required version is set
func (r *TerraformRequiredProvidersRule) Check(rr tflint.Runner) error {
	runner := rr.(*terraform.Runner)

	path, err := runner.GetModulePath()
	if err != nil {
		return err
	}
	if !path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	body, err := runner.GetModuleContent(&hclext.BodySchema{
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
	}, &tflint.GetModuleContentOption{ExpandMode: tflint.ExpandModeNone})
	if err != nil {
		return err
	}

	for _, provider := range body.Blocks {
		if _, exists := provider.Body.Attributes["version"]; exists {
			if err := runner.EmitIssue(
				r,
				`provider version constraint should be specified via "required_providers"`,
				provider.DefRange,
			); err != nil {
				return err
			}
		}
	}

	providerRefs, diags := runner.GetProviderRefs()
	if diags.HasErrors() {
		return diags
	}

	requiredProvidersSchema := []hclext.AttributeSchema{}
	for name := range providerRefs {
		requiredProvidersSchema = append(requiredProvidersSchema, hclext.AttributeSchema{Name: name})
	}

	body, err = runner.GetModuleContent(&hclext.BodySchema{
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
	}, &tflint.GetModuleContentOption{ExpandMode: tflint.ExpandModeNone})
	if err != nil {
		return err
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
			if err := runner.EmitIssue(r, fmt.Sprintf(`Missing version constraint for provider "%s" in "required_providers"`, name), ref.DefRange); err != nil {
				return err
			}
			continue
		}

		val, diags := provider.Expr.Value(&hcl.EvalContext{
			Variables: map[string]cty.Value{
				// configuration_aliases can declare additional provider instances
				// required provider "foo" could have: configuration_aliases = [foo.a, foo.b]
				// @see https://www.terraform.io/language/modules/develop/providers#provider-aliases-within-modules
				name: cty.DynamicVal,
			},
		})
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
			if err := runner.EmitIssue(r, fmt.Sprintf(`Missing version constraint for provider "%s" in "required_providers"`, name), ref.DefRange); err != nil {
				return err
			}
		}
	}

	return nil
}
