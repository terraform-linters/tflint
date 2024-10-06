package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformRequiredProviders checks whether ...
type TerraformRequiredProviders struct {
	tflint.DefaultRule
}

// NewTerraformRequiredProvidersRule returns a new rule
func NewTerraformRequiredProvidersRule() *TerraformRequiredProviders {
	return &TerraformRequiredProviders{}
}

// Name returns the rule name
func (r *TerraformRequiredProviders) Name() string {
	return "terraform_required_providers"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformRequiredProviders) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformRequiredProviders) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *TerraformRequiredProviders) Link() string {
	return ""
}

// Check checks whether ...
func (r *TerraformRequiredProviders) Check(runner tflint.Runner) error {
	module, err := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type: "terraform",
				Body: &hclext.BodySchema{
					Blocks: []hclext.BlockSchema{
						{
							Type: "required_providers",
							Body: &hclext.BodySchema{Mode: hclext.SchemaJustAttributesMode},
						},
					},
				},
			},
		},
	}, nil)
	if err != nil {
		return err
	}

	for _, terraform := range module.Blocks {
		for _, requiredProvider := range terraform.Body.Blocks {
			ret := []string{}
			for name, attr := range requiredProvider.Body.Attributes {
				v, diags := attr.Expr.Value(nil)
				if diags.HasErrors() {
					return diags
				}
				ret = append(ret, fmt.Sprintf("%s=%s", name, v.AsString()))
			}
			sort.Strings(ret)

			err := runner.EmitIssue(
				r,
				fmt.Sprintf("required_providers: %s", strings.Join(ret, ",")),
				requiredProvider.DefRange,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
