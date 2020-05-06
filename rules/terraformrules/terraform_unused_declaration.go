package terraformrules

import (
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/lang"
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
func (r *TerraformUnusedDeclarationsRule) Severity() string {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformUnusedDeclarationsRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check emits issues for any variables, locals, and data sources that are declared but not used
func (r *TerraformUnusedDeclarationsRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	decl := r.declarations(runner.TFConfig.Module)
	for _, resource := range runner.TFConfig.Module.ManagedResources {
		r.checkForRefsInBody(runner, resource.Config, decl)
	}
	for _, data := range runner.TFConfig.Module.DataResources {
		r.checkForRefsInBody(runner, data.Config, decl)
	}
	for _, provider := range runner.TFConfig.Module.ProviderConfigs {
		r.checkForRefsInBody(runner, provider.Config, decl)
	}
	for _, module := range runner.TFConfig.Module.ModuleCalls {
		r.checkForRefsInBody(runner, module.Config, decl)
	}
	for _, output := range runner.TFConfig.Module.Outputs {
		r.checkForRefsInExpr(runner, output.Expr, decl)
	}
	for _, local := range runner.TFConfig.Module.Locals {
		r.checkForRefsInExpr(runner, local.Expr, decl)
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

func (r *TerraformUnusedDeclarationsRule) checkForRefsInBody(runner *tflint.Runner, body hcl.Body, decl *declarations) {
	nativeBody, ok := body.(*hclsyntax.Body)
	if !ok {
		return
	}

	for _, attr := range nativeBody.Attributes {
		r.checkForRefsInExpr(runner, attr.Expr, decl)
	}

	for _, block := range nativeBody.Blocks {
		r.checkForRefsInBody(runner, block.Body, decl)
	}

	return
}

func (r *TerraformUnusedDeclarationsRule) checkForRefsInExpr(runner *tflint.Runner, expr hcl.Expression, decl *declarations) {
	refs, diags := lang.ReferencesInExpr(expr)
	if diags.HasErrors() {
		return
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
}
