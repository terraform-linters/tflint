package terraformrules

import (
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint/tflint"
)

// TerraformWorkspaceRemoteRule warns of the use of terraform.workspace with a remote backend
type TerraformWorkspaceRemoteRule struct{}

// NewTerraformWorkspaceRemoteRule return a new rule
func NewTerraformWorkspaceRemoteRule() *TerraformWorkspaceRemoteRule {
	return &TerraformWorkspaceRemoteRule{}
}

// Name returns the rule name
func (r *TerraformWorkspaceRemoteRule) Name() string {
	return "terraform_workspace_remote"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformWorkspaceRemoteRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *TerraformWorkspaceRemoteRule) Severity() string {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformWorkspaceRemoteRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks for a "remote" backend and if found emits issues for
// each use of terraform.workspace in an expression.
func (r *TerraformWorkspaceRemoteRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	backend := runner.TFConfig.Root.Module.Backend
	if backend == nil || backend.Type != "remote" {
		return nil
	}

	for _, resource := range runner.TFConfig.Module.ManagedResources {
		r.checkForTerraformWorkspaceInBody(runner, resource.Config)
	}
	for _, resource := range runner.TFConfig.Module.DataResources {
		r.checkForTerraformWorkspaceInBody(runner, resource.Config)
	}
	for _, provider := range runner.TFConfig.Module.ProviderConfigs {
		r.checkForTerraformWorkspaceInBody(runner, provider.Config)
	}
	for _, provider := range runner.TFConfig.Module.ModuleCalls {
		r.checkForTerraformWorkspaceInBody(runner, provider.Config)
	}
	for _, local := range runner.TFConfig.Module.Locals {
		r.checkForTerraformWorkspaceInExpr(runner, local.Expr)
	}
	for _, output := range runner.TFConfig.Module.Outputs {
		r.checkForTerraformWorkspaceInExpr(runner, output.Expr)
	}

	return nil
}

func (r *TerraformWorkspaceRemoteRule) checkForTerraformWorkspaceInBody(runner *tflint.Runner, body hcl.Body) {
	nativeBody, ok := body.(*hclsyntax.Body)
	if !ok {
		return
	}

	for _, attr := range nativeBody.Attributes {
		r.checkForTerraformWorkspaceInExpr(runner, attr.Expr)
	}

	for _, block := range nativeBody.Blocks {
		r.checkForTerraformWorkspaceInBody(runner, block.Body)
	}

	return
}

func (r *TerraformWorkspaceRemoteRule) checkForTerraformWorkspaceInExpr(runner *tflint.Runner, expr hcl.Expression) {
	var used bool
	for _, t := range expr.Variables() {
		if len(t) != 2 || t.IsRelative() || t.RootName() != "terraform" {
			continue
		}

		if t[1].(hcl.TraverseAttr).Name == "workspace" {
			used = true
			break
		}
	}

	if used {
		runner.EmitIssue(
			r,
			"terraform.workspace should not be used with a 'remote' backend",
			expr.Range(),
		)
	}
}
