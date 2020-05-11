package terraformrules

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"

	"github.com/terraform-linters/tflint/tflint"
)

// TerraformTypedVariablesRule checks whether variables have a type declared
type TerraformTypedVariablesRule struct{}

// NewTerraformTypedVariablesRule returns a new rule
func NewTerraformTypedVariablesRule() *TerraformTypedVariablesRule {
	return &TerraformTypedVariablesRule{}
}

// Name returns the rule name
func (r *TerraformTypedVariablesRule) Name() string {
	return "terraform_typed_variables"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformTypedVariablesRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformTypedVariablesRule) Severity() string {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformTypedVariablesRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks whether variables have type
func (r *TerraformTypedVariablesRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	files := make(map[string]*struct{})
	for _, variable := range runner.TFConfig.Module.Variables {
		files[variable.DeclRange.Filename] = nil
	}

	for filename := range files {
		if err := r.checkFileSchema(runner, filename); err != nil {
			return err
		}
	}

	return nil
}

func (r *TerraformTypedVariablesRule) checkFileSchema(runner *tflint.Runner, filename string) error {
	bytes, err := runner.ReadFile(filename)
	if err != nil {
		return err
	}

	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(bytes, filename)
	if diags.HasErrors() {
		return diags
	}

	content, _, diags := file.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "variable",
				LabelNames: []string{"name"},
			},
		},
	})

	for _, block := range content.Blocks.OfType("variable") {
		_, _, diags := block.Body.PartialContent(&hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{
				{
					Name:     "type",
					Required: true,
				},
			},
		})

		if diags.HasErrors() {
			runner.EmitIssue(
				r,
				fmt.Sprintf("`%v` variable has no type", block.Labels[0]),
				block.DefRange,
			)
		}
	}

	return nil
}
