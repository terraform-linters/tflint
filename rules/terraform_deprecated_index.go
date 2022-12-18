package rules

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
)

// TerraformDeprecatedIndexRule warns about usage of the legacy dot syntax for indexes (foo.0)
type TerraformDeprecatedIndexRule struct {
	tflint.DefaultRule
}

// NewTerraformDeprecatedIndexRule return a new rule
func NewTerraformDeprecatedIndexRule() *TerraformDeprecatedIndexRule {
	return &TerraformDeprecatedIndexRule{}
}

// Name returns the rule name
func (r *TerraformDeprecatedIndexRule) Name() string {
	return "terraform_deprecated_index"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformDeprecatedIndexRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *TerraformDeprecatedIndexRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformDeprecatedIndexRule) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check walks all expressions and emit issues if deprecated index syntax is found
func (r *TerraformDeprecatedIndexRule) Check(runner tflint.Runner) error {
	path, err := runner.GetModulePath()
	if err != nil {
		return err
	}
	if !path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	diags := runner.WalkExpressions(tflint.ExprWalkFunc(func(expr hcl.Expression) hcl.Diagnostics {
		for _, variable := range expr.Variables() {
			filename := expr.Range().Filename
			file := files[filename]

			bytes := expr.Range().SliceBytes(file.Bytes)

			tokens, diags := hclsyntax.LexExpression(bytes, filename, variable.SourceRange().Start)
			if diags.HasErrors() {
				// HACK: If the expression cannot be lexed, try to lex it as a template.
				// If it still cannot be lexed, return the original error.
				tTokens, tDiags := hclsyntax.LexTemplate(bytes, filename, variable.SourceRange().Start)
				if tDiags.HasErrors() {
					return diags
				}

				tokens = tTokens
			}

			tokens = tokens[1:]

			for i, token := range tokens {
				if token.Type == hclsyntax.TokenDot {
					if len(tokens) == i+1 {
						return nil
					}

					next := tokens[i+1].Type
					if next == hclsyntax.TokenNumberLit || next == hclsyntax.TokenStar {
						if tokens[0].Type == hclsyntax.TokenDot {
							if err := runner.EmitIssue(
								r,
								"List items should be accessed using square brackets",
								expr.Range(),
							); err != nil {
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
			}
		}

		return nil
	}))

	if diags.HasErrors() {
		return diags
	}

	return nil
}
