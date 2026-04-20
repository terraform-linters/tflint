// SPDX-License-Identifier: MPL-2.0

package terraform

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint/terraform/tfdiags"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

type Variable struct {
	Name    string
	Default cty.Value

	Type           cty.Type
	ConstraintType cty.Type
	TypeDefaults   *typeexpr.Defaults

	DeclRange hcl.Range

	ParsingMode VariableParsingMode
	Sensitive   bool
	Ephemeral   bool
	Nullable    bool
}

func decodeVariableBlock(block *hclext.Block) (*Variable, hcl.Diagnostics) {
	variable := newVariable(block.Labels[0], block.DefRange)
	var diags hcl.Diagnostics

	if attr, ok := block.Body.Attributes["type"]; ok {
		applyVariableType(variable, attr.Expr, &diags)
	}
	if attr, ok := block.Body.Attributes["sensitive"]; ok {
		diags = diags.Extend(decodeBoolAttribute(attr.Expr, &variable.Sensitive))
	}
	if attr, ok := block.Body.Attributes["ephemeral"]; ok {
		diags = diags.Extend(decodeBoolAttribute(attr.Expr, &variable.Ephemeral))
	}
	if attr, ok := block.Body.Attributes["nullable"]; ok {
		diags = diags.Extend(decodeBoolAttribute(attr.Expr, &variable.Nullable))
	} else {
		variable.Nullable = true
	}
	if attr, ok := block.Body.Attributes["default"]; ok {
		applyVariableDefault(variable, attr, &diags)
	}

	return variable, diags
}

func decodeVariableType(expr hcl.Expression) (cty.Type, *typeexpr.Defaults, VariableParsingMode, hcl.Diagnostics) {
	if ty, defaults, mode, diags, handled := decodeLegacyQuotedVariableType(expr); handled {
		return ty, defaults, mode, diags
	}
	if ty, mode, handled := decodeShorthandVariableType(expr); handled {
		return ty, nil, mode, nil
	}

	ty, typeDefaults, diags := typeexpr.TypeConstraintWithDefaults(expr)
	if diags.HasErrors() {
		return cty.DynamicPseudoType, nil, VariableParseHCL, diags
	}
	if ty.IsPrimitiveType() {
		return ty, typeDefaults, VariableParseLiteral, diags
	}
	return ty, typeDefaults, VariableParseHCL, diags
}

// VariableParsingMode defines how values of a particular variable given by
// text-only mechanisms (command line arguments and environment variables)
// should be parsed to produce the final value.
type VariableParsingMode rune

// VariableParseLiteral is a variable parsing mode that just takes the given
// string directly as a cty.String value.
const VariableParseLiteral VariableParsingMode = 'L'

// VariableParseHCL is a variable parsing mode that attempts to parse the given
// string as an HCL expression and returns the result.
const VariableParseHCL VariableParsingMode = 'H'

func (m VariableParsingMode) Parse(name, value string) (cty.Value, hcl.Diagnostics) {
	switch m {
	case VariableParseLiteral:
		return cty.StringVal(value), nil
	case VariableParseHCL:
		filename := fmt.Sprintf("<value for var.%s>", name)
		expr, diags := hclsyntax.ParseExpression([]byte(value), filename, hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			return cty.DynamicVal, diags
		}
		val, valueDiags := expr.Value(nil)
		diags = diags.Extend(valueDiags)
		return val, diags
	default:
		panic(fmt.Errorf("Parse called on invalid VariableParsingMode %#v", m))
	}
}

func newVariable(name string, declRange hcl.Range) *Variable {
	return &Variable{
		Name:           name,
		Type:           cty.DynamicPseudoType,
		ConstraintType: cty.DynamicPseudoType,
		ParsingMode:    VariableParseLiteral,
		DeclRange:      declRange,
	}
}

func applyVariableType(variable *Variable, expr hcl.Expression, diags *hcl.Diagnostics) {
	ty, defaults, parsingMode, typeDiags := decodeVariableType(expr)
	*diags = diags.Extend(typeDiags)
	variable.ConstraintType = ty
	variable.TypeDefaults = defaults
	variable.Type = ty.WithoutOptionalAttributesDeep()
	variable.ParsingMode = parsingMode
}

func decodeBoolAttribute(expr hcl.Expression, target *bool) hcl.Diagnostics {
	return gohcl.DecodeExpression(expr, nil, target)
}

func applyVariableDefault(variable *Variable, attr *hclext.Attribute, diags *hcl.Diagnostics) {
	value, valueDiags := attr.Expr.Value(nil)
	*diags = diags.Extend(valueDiags)

	if variable.ConstraintType != cty.NilType {
		value = applyVariableDefaults(variable, value)

		converted, err := convert.Convert(value, variable.ConstraintType)
		if err != nil {
			*diags = append(*diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid default value for variable",
				Detail: fmt.Sprintf(
					"This default value is not compatible with the variable's type constraint: %s.",
					tfdiags.FormatError(err),
				),
				Subject: attr.Expr.Range().Ptr(),
			})
			value = cty.DynamicVal
		} else {
			value = converted
		}
	}

	variable.Default = value
}

func applyVariableDefaults(variable *Variable, value cty.Value) cty.Value {
	if variable.TypeDefaults != nil && !value.IsNull() {
		return variable.TypeDefaults.Apply(value)
	}
	return value
}

func decodeLegacyQuotedVariableType(expr hcl.Expression) (cty.Type, *typeexpr.Defaults, VariableParsingMode, hcl.Diagnostics, bool) {
	if !exprIsNativeQuotedString(expr) {
		return cty.NilType, nil, VariableParseHCL, nil, false
	}

	value, diags := expr.Value(nil)
	if diags.HasErrors() {
		return cty.DynamicPseudoType, nil, VariableParseHCL, diags, true
	}

	switch value.AsString() {
	case "string":
		return cty.DynamicPseudoType, nil, VariableParseLiteral, invalidQuotedTypeDiag(expr, `Terraform 0.11 and earlier required type constraints to be given in quotes, but that form is now deprecated and will be removed in a future version of Terraform. Remove the quotes around "string".`), true
	case "list":
		return cty.DynamicPseudoType, nil, VariableParseHCL, invalidQuotedTypeDiag(expr, `Terraform 0.11 and earlier required type constraints to be given in quotes, but that form is now deprecated and will be removed in a future version of Terraform. Remove the quotes around "list" and write list(string) instead to explicitly indicate that the list elements are strings.`), true
	case "map":
		return cty.DynamicPseudoType, nil, VariableParseHCL, invalidQuotedTypeDiag(expr, `Terraform 0.11 and earlier required type constraints to be given in quotes, but that form is now deprecated and will be removed in a future version of Terraform. Remove the quotes around "map" and write map(string) instead to explicitly indicate that the map elements are strings.`), true
	default:
		return cty.DynamicPseudoType, nil, VariableParseHCL, hcl.Diagnostics{{
			Severity: hcl.DiagError,
			Summary:  "Invalid legacy variable type hint",
			Detail:   `To provide a full type expression, remove the surrounding quotes and give the type expression directly.`,
			Subject:  expr.Range().Ptr(),
		}}, true
	}
}

func decodeShorthandVariableType(expr hcl.Expression) (cty.Type, VariableParsingMode, bool) {
	switch hcl.ExprAsKeyword(expr) {
	case "list":
		return cty.List(cty.DynamicPseudoType), VariableParseHCL, true
	case "map":
		return cty.Map(cty.DynamicPseudoType), VariableParseHCL, true
	default:
		return cty.NilType, VariableParseHCL, false
	}
}

func invalidQuotedTypeDiag(expr hcl.Expression, detail string) hcl.Diagnostics {
	return hcl.Diagnostics{{
		Severity: hcl.DiagError,
		Summary:  "Invalid quoted type constraints",
		Detail:   detail,
		Subject:  expr.Range().Ptr(),
	}}
}

func exprIsNativeQuotedString(expr hcl.Expression) bool {
	_, ok := expr.(*hclsyntax.TemplateExpr)
	return ok
}

var variableBlockSchema = &hclext.BodySchema{
	Attributes: []hclext.AttributeSchema{
		{Name: "default"},
		{Name: "type"},
		{Name: "sensitive"},
		{Name: "ephemeral"},
		{Name: "nullable"},
	},
}
