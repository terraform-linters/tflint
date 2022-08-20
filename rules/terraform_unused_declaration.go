package rules

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
)

// TerraformUnusedDeclarationsRule checks whether variables, data sources, or locals are declared but unused
type TerraformUnusedDeclarationsRule struct {
	tflint.DefaultRule
}

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
	return true
}

// Severity returns the rule severity
func (r *TerraformUnusedDeclarationsRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformUnusedDeclarationsRule) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check emits issues for any variables, locals, and data sources that are declared but not used
func (r *TerraformUnusedDeclarationsRule) Check(runner tflint.Runner) error {
	path, err := runner.GetModulePath()
	if err != nil {
		return err
	}
	if !path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	decl, err := r.declarations(runner)
	if err != nil {
		return err
	}
	err = WalkExpressions(runner, func(expr hcl.Expression) error {
		return r.checkForRefsInExpr(expr, decl)
	})
	if err != nil {
		return err
	}

	for _, variable := range decl.Variables {
		if err := runner.EmitIssue(
			r,
			fmt.Sprintf(`variable "%s" is declared but not used`, variable.Labels[0]),
			variable.DefRange,
		); err != nil {
			return err
		}
	}
	for _, data := range decl.DataResources {
		if err := runner.EmitIssue(
			r,
			fmt.Sprintf(`data "%s" "%s" is declared but not used`, data.Labels[0], data.Labels[1]),
			data.DefRange,
		); err != nil {
			return err
		}
	}
	for _, local := range decl.Locals {
		if err := runner.EmitIssue(
			r,
			fmt.Sprintf(`local.%s is declared but not used`, local.name),
			local.defRange,
		); err != nil {
			return err
		}
	}

	return nil
}

func (r *TerraformUnusedDeclarationsRule) declarations(runner tflint.Runner) (*declarations, error) {
	decl := &declarations{
		Variables:     map[string]*hclext.Block{},
		DataResources: map[string]*hclext.Block{},
		Locals:        map[string]*local{},
	}

	body, err := runner.GetModuleContent(&hclext.BodySchema{
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
	}, &tflint.GetModuleContentOption{IncludeNotCreated: true})
	if err != nil {
		return decl, err
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
	for _, ref := range referencesInExpr(expr) {
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
