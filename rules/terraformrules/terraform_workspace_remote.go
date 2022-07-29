package terraformrules

import (
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
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

	body, diags := runner.GetModuleContent(&hclext.BodySchema{
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
	}, sdk.GetModuleContentOption{IncludeNotCreated: true})
	if diags.HasErrors() {
		return diags
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

	return runner.WalkExpressions(func(expr hcl.Expression) error {
		return r.checkForTerraformWorkspaceInExpr(runner, expr)
	})
}

func (r *TerraformWorkspaceRemoteRule) checkForTerraformWorkspaceInExpr(runner *tflint.Runner, expr hcl.Expression) error {
	_, isScopeTraversalExpr := expr.(*hclsyntax.ScopeTraversalExpr)
	if !isScopeTraversalExpr && !json.IsJSONExpression(expr) {
		return nil
	}

	refs, diags := referencesInExpr(expr)
	if diags.HasErrors() {
		return diags
	}

	for _, ref := range refs {
		switch sub := ref.subject.(type) {
		case terraformReference:
			if sub.name == "workspace" {
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
