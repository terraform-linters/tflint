package rules

import (
	"github.com/hashicorp/go-version"
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

// @see https://releases.hashicorp.com/terraform/
var tf10Versions = []*version.Version{
	version.Must(version.NewVersion("1.0.0")),
	version.Must(version.NewVersion("1.0.1")),
	version.Must(version.NewVersion("1.0.2")),
	version.Must(version.NewVersion("1.0.3")),
	version.Must(version.NewVersion("1.0.4")),
	version.Must(version.NewVersion("1.0.5")),
	version.Must(version.NewVersion("1.0.6")),
	version.Must(version.NewVersion("1.0.7")),
	version.Must(version.NewVersion("1.0.8")),
	version.Must(version.NewVersion("1.0.9")),
	version.Must(version.NewVersion("1.0.10")),
	version.Must(version.NewVersion("1.0.11")),
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
					Attributes: []hclext.AttributeSchema{
						{Name: "required_version"},
					},
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
	}, &tflint.GetModuleContentOption{ExpandMode: tflint.ExpandModeNone})
	if err != nil {
		return err
	}

	var remoteBackend bool
	var tf10Support bool
	for _, terraform := range body.Blocks {
		for _, requiredVersion := range terraform.Body.Attributes {
			err := runner.EvaluateExpr(requiredVersion.Expr, func(v string) error {
				constraints, err := version.NewConstraint(v)
				if err != nil {
					return err
				}

				for _, tf10Version := range tf10Versions {
					if constraints.Check(tf10Version) {
						tf10Support = true
					}
				}
				return nil
			}, nil)
			if err != nil {
				return err
			}
		}

		for _, backend := range terraform.Body.Blocks {
			if backend.Labels[0] == "remote" {
				remoteBackend = true
			}
		}
	}
	if !remoteBackend || !tf10Support {
		return nil
	}

	diags := runner.WalkExpressions(tflint.ExprWalkFunc(func(expr hcl.Expression) hcl.Diagnostics {
		return r.checkForTerraformWorkspaceInExpr(runner, expr)
	}))
	if diags.HasErrors() {
		return diags
	}
	return nil
}

func (r *TerraformWorkspaceRemoteRule) checkForTerraformWorkspaceInExpr(runner tflint.Runner, expr hcl.Expression) hcl.Diagnostics {
	_, isScopeTraversalExpr := expr.(*hclsyntax.ScopeTraversalExpr)
	if !isScopeTraversalExpr && !json.IsJSONExpression(expr) {
		return nil
	}

	for _, ref := range lang.ReferencesInExpr(expr) {
		switch sub := ref.Subject.(type) {
		case addrs.TerraformAttr:
			if sub.Name == "workspace" {
				err := runner.EmitIssue(
					r,
					"terraform.workspace should not be used with a 'remote' backend",
					expr.Range(),
				)
				if err != nil {
					return hcl.Diagnostics{
						{
							Severity: hcl.DiagError,
							Summary:  "failed to call EmitIssue()",
							Detail:   err.Error(),
						},
					}
				}
			}
		}
	}

	return nil
}
