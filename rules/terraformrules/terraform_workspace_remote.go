package terraformrules

import (
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/terraform-linters/tflint/terraform/lang"
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
func (r *TerraformWorkspaceRemoteRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformWorkspaceRemoteRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check checks for a "remote" backend and if found emits issues for
// each use of terraform.workspace in an expression.
func (r *TerraformWorkspaceRemoteRule) Check(runner *tflint.Runner) error {
	if !runner.TFConfig.Path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	backend := runner.TFConfig.Root.Module.Backend
	if backend == nil || backend.Type != "remote" {
		return nil
	}

	return runner.WalkExpressions(func(expr hcl.Expression) error {
		return r.checkForTerraformWorkspaceInExpr(runner, expr)
	})
}

func (r *TerraformWorkspaceRemoteRule) checkForTerraformWorkspaceInExpr(runner *tflint.Runner, expr hcl.Expression) error {
	refs, diags := lang.ReferencesInExpr(expr)
	if diags.HasErrors() {
		log.Printf("[DEBUG] Cannot find references in expression, ignoring: %v", diags.Err())
		return nil
	}

	for _, ref := range refs {
		switch sub := ref.Subject.(type) {
		case addrs.TerraformAttr:
			if sub.Name == "workspace" {
				runner.EmitIssue(
					r,
					"terraform.workspace should not be used with a 'remote' backend",
					expr.Range(),
				)
				return nil
			}
		}
	}

	return nil
}
