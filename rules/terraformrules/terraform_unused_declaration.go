package terraformrules

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/tflint"
)

// TerraformUnusedDeclarationsRule checks whether variables, data sources, or locals are declared but unused
type TerraformUnusedDeclarationsRule struct{}

type declarations struct {
	Variables     map[string]*hclext.Block
	DataResources map[string]*hclext.Block
	Locals        map[string]*local
}

// NewTerraformUnusedDeclarationsRule returns a new rule
func NewTerraformUnusedDeclarationsRule() *TerraformUnusedDeclarationsRule {
	return &TerraformUnusedDeclarationsRule{}
}

// Name returns the rule name
func (r *TerraformUnusedDeclarationsRule) Name() string {
	return "terraform_unused_declarations"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformUnusedDeclarationsRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformUnusedDeclarationsRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformUnusedDeclarationsRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check emits issues for any variables, locals, and data sources that are declared but not used
func (r *TerraformUnusedDeclarationsRule) Check(runner *tflint.Runner) error {
	if !runner.TFConfig.Path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	decl, err := r.declarations(runner)
	if err != nil {
		return err
	}
	err = runner.WalkExpressions(func(expr hcl.Expression) error {
		return r.checkForRefsInExpr(expr, decl)
	})
	if err != nil {
		return err
	}

	for _, variable := range decl.Variables {
		runner.EmitIssue(
			r,
			fmt.Sprintf(`variable "%s" is declared but not used`, variable.Labels[0]),
			variable.DefRange,
		)
	}
	for _, data := range decl.DataResources {
		runner.EmitIssue(
			r,
			fmt.Sprintf(`data "%s" "%s" is declared but not used`, data.Labels[0], data.Labels[1]),
			data.DefRange,
		)
	}
	for _, local := range decl.Locals {
		runner.EmitIssue(
			r,
			fmt.Sprintf(`local.%s is declared but not used`, local.name),
			local.defRange,
		)
	}

	return nil
}

func (r *TerraformUnusedDeclarationsRule) declarations(runner *tflint.Runner) (*declarations, error) {
	decl := &declarations{
		Variables:     map[string]*hclext.Block{},
		DataResources: map[string]*hclext.Block{},
		Locals:        map[string]*local{},
	}

	body, diags := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "variable",
				LabelNames: []string{"name"},
				Body:       &hclext.BodySchema{},
			},
			{
				Type:       "data",
				LabelNames: []string{"type", "name"},
				Body:       &hclext.BodySchema{},
			},
		},
	}, sdk.GetModuleContentOption{})
	if diags.HasErrors() {
		return decl, diags
	}

	for _, block := range body.Blocks {
		if block.Type == "variable" {
			decl.Variables[block.Labels[0]] = block
		} else {
			decl.DataResources[fmt.Sprintf("data.%s.%s", block.Labels[0], block.Labels[1])] = block
		}
	}

	locals, diags := getLocals(runner)
	if diags.HasErrors() {
		return decl, diags
	}
	decl.Locals = locals

	return decl, nil
}

func (r *TerraformUnusedDeclarationsRule) checkForRefsInExpr(expr hcl.Expression, decl *declarations) error {
	refs, diags := referencesInExpr(expr)
	if diags.HasErrors() {
		log.Printf("[DEBUG] Cannot find references in expression, ignoring: %v", diags)
		return nil
	}

	for _, ref := range refs {
		switch sub := ref.subject.(type) {
		case inputVariableReference:
			delete(decl.Variables, sub.name)
		case localValueReference:
			delete(decl.Locals, sub.name)
		case dataResourceReference:
			delete(decl.DataResources, sub.String())
		}
	}

	return nil
}
