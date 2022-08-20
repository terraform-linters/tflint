package rules

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
)

// TerraformWorkspaceRemoteRule warns of the use of terraform.workspace with a remote backend
type TerraformWorkspaceRemoteRule struct {
	tflint.DefaultRule
}

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
	return project.ReferenceLink(r.Name())
}

// Check checks for a "remote" backend and if found emits issues for
// each use of terraform.workspace in an expression.
func (r *TerraformWorkspaceRemoteRule) Check(runner tflint.Runner) error {
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
				Type: "terraform",
				Body: &hclext.BodySchema{
					Blocks: []hclext.BlockSchema{
						{
							Type:       "backend",
							LabelNames: []string{"type"},
							Body:       &hclext.BodySchema{},
						},
					},
				},
			},
		},
	}, &tflint.GetModuleContentOption{IncludeNotCreated: true})
	if err != nil {
		return err
	}

	var remoteBackend bool
	for _, terraform := range body.Blocks {
		for _, backend := range terraform.Body.Blocks {
			if backend.Labels[0] == "remote" {
				remoteBackend = true
			}
		}
	}
	if !remoteBackend {
		return nil
	}

	return WalkExpressions(runner, func(expr hcl.Expression) error {
		return r.checkForTerraformWorkspaceInExpr(runner, expr)
	})
}

func (r *TerraformWorkspaceRemoteRule) checkForTerraformWorkspaceInExpr(runner tflint.Runner, expr hcl.Expression) error {
	_, isScopeTraversalExpr := expr.(*hclsyntax.ScopeTraversalExpr)
	if !isScopeTraversalExpr && !json.IsJSONExpression(expr) {
		return nil
	}

	for _, ref := range lang.ReferencesInExpr(expr) {
		switch sub := ref.Subject.(type) {
		case addrs.TerraformAttr:
			if sub.Name == "workspace" {
				return runner.EmitIssue(
					r,
					"terraform.workspace should not be used with a 'remote' backend",
					expr.Range(),
				)
			}
		}
	}

	return nil
}
