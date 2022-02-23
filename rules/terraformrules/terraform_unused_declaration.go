package terraformrules

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/terraform-linters/tflint/terraform/configs"
	"github.com/terraform-linters/tflint/terraform/lang"
	"github.com/terraform-linters/tflint/tflint"
)

// TerraformUnusedDeclarationsRule checks whether variables, data sources, or locals are declared but unused
type TerraformUnusedDeclarationsRule struct{}

type declarations struct {
	Variables     map[string]*configs.Variable
	DataResources map[string]*configs.Resource
	Locals        map[string]*configs.Local
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

	decl := r.declarations(runner.TFConfig.Module)
	err := runner.WalkExpressions(func(expr hcl.Expression) error {
		return r.checkForRefsInExpr(expr, decl)
	})
	if err != nil {
		return err
	}

	for _, variable := range decl.Variables {
		runner.EmitIssue(
			r,
			fmt.Sprintf(`variable "%s" is declared but not used`, variable.Name),
			variable.DeclRange,
		)
	}
	for _, data := range decl.DataResources {
		runner.EmitIssue(
			r,
			fmt.Sprintf(`data "%s" "%s" is declared but not used`, data.Type, data.Name),
			data.DeclRange,
		)
	}
	for _, local := range decl.Locals {
		runner.EmitIssue(
			r,
			fmt.Sprintf(`local.%s is declared but not used`, local.Name),
			local.DeclRange,
		)
	}

	return nil
}

func (r *TerraformUnusedDeclarationsRule) declarations(module *configs.Module) *declarations {
	decl := &declarations{
		Variables:     make(map[string]*configs.Variable, len(module.Variables)),
		DataResources: make(map[string]*configs.Resource, len(module.DataResources)),
		Locals:        make(map[string]*configs.Local, len(module.Locals)),
	}

	for k, v := range module.Variables {
		decl.Variables[k] = v
	}
	for k, v := range module.DataResources {
		decl.DataResources[k] = v
	}
	for k, v := range module.Locals {
		decl.Locals[k] = v
	}

	return decl
}

func (r *TerraformUnusedDeclarationsRule) checkForRefsInExpr(expr hcl.Expression, decl *declarations) error {
	refs, diags := lang.ReferencesInExpr(expr)
	if diags.HasErrors() {
		log.Printf("[DEBUG] Cannot find references in expression, ignoring: %v", diags.Err())
		return nil
	}

	for _, ref := range refs {
		switch sub := ref.Subject.(type) {
		case addrs.InputVariable:
			delete(decl.Variables, sub.Name)
		case addrs.LocalValue:
			delete(decl.Locals, sub.Name)
		case addrs.Resource:
			delete(decl.DataResources, sub.String())
		case addrs.ResourceInstance:
			delete(decl.DataResources, sub.Resource.String())
		}
	}

	return nil
}
