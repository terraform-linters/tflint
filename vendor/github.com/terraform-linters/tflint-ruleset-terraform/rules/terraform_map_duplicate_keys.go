package rules

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/logger"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

// This rule checks for map literals with duplicate keys
type TerraformMapDuplicateKeysRule struct {
	tflint.DefaultRule
}

func NewTerraformMapDuplicateKeysRule() *TerraformMapDuplicateKeysRule {
	return &TerraformMapDuplicateKeysRule{}
}

func (r *TerraformMapDuplicateKeysRule) Name() string {
	return "terraform_map_duplicate_keys"
}

func (r *TerraformMapDuplicateKeysRule) Enabled() bool {
	return true
}

func (r *TerraformMapDuplicateKeysRule) Severity() tflint.Severity {
	return tflint.WARNING
}

func (r *TerraformMapDuplicateKeysRule) Link() string {
	return project.ReferenceLink(r.Name())
}

func (r *TerraformMapDuplicateKeysRule) Check(runner tflint.Runner) error {
	path, err := runner.GetModulePath()
	if err != nil {
		return err
	}
	if !path.IsRoot() {
		// This rule does not evaluate child modules
		return nil
	}

	diags := runner.WalkExpressions(tflint.ExprWalkFunc(func(e hcl.Expression) hcl.Diagnostics {
		return r.checkObjectConsExpr(e, runner)
	}))
	if diags.HasErrors() {
		return diags
	}

	return nil
}

func (r *TerraformMapDuplicateKeysRule) checkObjectConsExpr(e hcl.Expression, runner tflint.Runner) hcl.Diagnostics {
	objExpr, ok := e.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return nil
	}

	var diags hcl.Diagnostics
	keys := make(map[string]hcl.Range)

	for _, item := range objExpr.Items {
		expr := item.KeyExpr.(*hclsyntax.ObjectConsKeyExpr)
		var val cty.Value

		// There is an issue with the SDK's EvaluateExpr not being able to evaluate naked identifiers of map keys, so we will handle this here.
		// @see https://github.com/terraform-linters/tflint-plugin-sdk/issues/338
		//
		// Checks whether the key expression can be extracted as a keyword and retrieves its value in the same way as ObjectConsKeyExpr.Value.
		// @see https://github.com/hashicorp/hcl/blob/v2.21.0/hclsyntax/expression.go#L1311
		if keyword := hcl.ExprAsKeyword(expr.Wrapped); !expr.ForceNonLiteral && keyword != "" {
			val = cty.StringVal(keyword)
		} else {
			err := runner.EvaluateExpr(expr, &val, nil)
			if err != nil {
				// When a key fails to evaluate, ignore the key and continue processing rather than terminating with an error.
				// This is due to a limitation that expressions with different scopes, such as for expressions, cannot be evaluated.
				// @see https://github.com/terraform-linters/tflint-ruleset-terraform/issues/199
				logger.Debug("Failed to evaluate key. The key will be ignored", "range", expr.Range(), "error", err.Error())
				continue
			}
		}

		if !val.IsKnown() || val.IsNull() || val.IsMarked() {
			logger.Debug("Unprocessable key, continuing", "range", expr.Range())
			continue
		}
		// Map keys must be strings, but some values ​​can be converted to strings and become valid keys,
		// so try to convert them here.
		if converted, err := convert.Convert(val, cty.String); err == nil {
			val = converted
		}
		if val.Type() != cty.String {
			logger.Debug("Unprocessable key, continuing", "range", expr.Range())
			continue
		}

		if declRange, exists := keys[val.AsString()]; exists {
			if err := runner.EmitIssue(
				r,
				fmt.Sprintf("Duplicate key: %q, first defined at %s", val.AsString(), declRange),
				expr.Range(),
			); err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "failed to call EmitIssue()",
					Detail:   err.Error(),
				})

				return diags
			}

			continue
		}

		keys[val.AsString()] = expr.Range()
	}

	return diags
}
