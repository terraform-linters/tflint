package tflint

import (
	"fmt"

	version "github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl2/ext/typeexpr"
	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/configs"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

func sniffCoreVersionRequirements(body hcl.Body) ([]configs.VersionConstraint, hcl.Diagnostics) {
	rootContent, _, diags := body.PartialContent(configFileVersionSniffRootSchema)

	var constraints []configs.VersionConstraint

	for _, block := range rootContent.Blocks {
		content, _, blockDiags := block.Body.PartialContent(configFileVersionSniffBlockSchema)
		diags = append(diags, blockDiags...)

		attr, exists := content.Attributes["required_version"]
		if !exists {
			continue
		}

		constraint, constraintDiags := decodeVersionConstraint(attr)
		diags = append(diags, constraintDiags...)
		if !constraintDiags.HasErrors() {
			constraints = append(constraints, constraint)
		}
	}

	return constraints, diags
}

var configFileSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "terraform",
		},
		{
			Type:       "provider",
			LabelNames: []string{"name"},
		},
		{
			Type:       "variable",
			LabelNames: []string{"name"},
		},
		{
			Type: "locals",
		},
		{
			Type:       "output",
			LabelNames: []string{"name"},
		},
		{
			Type:       "module",
			LabelNames: []string{"name"},
		},
		{
			Type:       "resource",
			LabelNames: []string{"type", "name"},
		},
		{
			Type:       "data",
			LabelNames: []string{"type", "name"},
		},
	},
}

var terraformBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: "required_version",
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "backend",
			LabelNames: []string{"type"},
		},
		{
			Type: "required_providers",
		},
	},
}

var configFileVersionSniffRootSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "terraform",
		},
	},
}

var configFileVersionSniffBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: "required_version",
		},
	},
}

// https://github.com/hashicorp/terraform/blob/v0.12.6/configs/version_constraint.go
func decodeVersionConstraint(attr *hcl.Attribute) (configs.VersionConstraint, hcl.Diagnostics) {
	ret := configs.VersionConstraint{
		DeclRange: attr.Range,
	}

	val, diags := attr.Expr.Value(nil)
	if diags.HasErrors() {
		return ret, diags
	}
	var err error
	val, err = convert.Convert(val, cty.String)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid version constraint",
			Detail:   fmt.Sprintf("A string value is required for %s.", attr.Name),
			Subject:  attr.Expr.Range().Ptr(),
		})
		return ret, diags
	}

	if val.IsNull() {
		// A null version constraint is strange, but we'll just treat it
		// like an empty constraint set.
		return ret, diags
	}

	if !val.IsWhollyKnown() {
		// If there is a syntax error, HCL sets the value of the given attribute
		// to cty.DynamicVal. A diagnostic for the syntax error will already
		// bubble up, so we will move forward gracefully here.
		return ret, diags
	}

	constraintStr := val.AsString()
	constraints, err := version.NewConstraint(constraintStr)
	if err != nil {
		// NewConstraint doesn't return user-friendly errors, so we'll just
		// ignore the provided error and produce our own generic one.
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid version constraint",
			Detail:   "This string does not use correct version constraint syntax.", // Not very actionable :(
			Subject:  attr.Expr.Range().Ptr(),
		})
		return ret, diags
	}

	ret.Required = constraints
	return ret, diags
}

// https://github.com/hashicorp/terraform/blob/v0.12.6/configs/backend.go
func decodeBackendBlock(block *hcl.Block) (*configs.Backend, hcl.Diagnostics) {
	return &configs.Backend{
		Type:      block.Labels[0],
		TypeRange: block.LabelRanges[0],
		Config:    block.Body,
		DeclRange: block.DefRange,
	}, nil
}

// https://github.com/hashicorp/terraform/blob/v0.12.6/configs/provider.go
func decodeProviderBlock(block *hcl.Block) (*configs.Provider, hcl.Diagnostics) {
	content, config, diags := block.Body.PartialContent(providerBlockSchema)

	provider := &configs.Provider{
		Name:      block.Labels[0],
		NameRange: block.LabelRanges[0],
		Config:    config,
		DeclRange: block.DefRange,
	}

	if attr, exists := content.Attributes["alias"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &provider.Alias)
		diags = append(diags, valDiags...)
		provider.AliasRange = attr.Expr.Range().Ptr()

		if !hclsyntax.ValidIdentifier(provider.Alias) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid provider configuration alias",
				Detail:   fmt.Sprintf("An alias must be a valid name. %s", badIdentifierDetail),
			})
		}
	}

	if attr, exists := content.Attributes["version"]; exists {
		var versionDiags hcl.Diagnostics
		provider.Version, versionDiags = decodeVersionConstraint(attr)
		diags = append(diags, versionDiags...)
	}

	// Reserved attribute names
	for _, name := range []string{"count", "depends_on", "for_each", "source"} {
		if attr, exists := content.Attributes[name]; exists {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Reserved argument name in provider block",
				Detail:   fmt.Sprintf("The provider argument name %q is reserved for use by Terraform in a future version.", name),
				Subject:  &attr.NameRange,
			})
		}
	}

	// Reserved block types (all of them)
	for _, block := range content.Blocks {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Reserved block type name in provider block",
			Detail:   fmt.Sprintf("The block type name %q is reserved for use by Terraform in a future version.", block.Type),
			Subject:  &block.TypeRange,
		})
	}

	return provider, diags
}

func decodeRequiredProvidersBlock(block *hcl.Block) ([]*configs.ProviderRequirement, hcl.Diagnostics) {
	attrs, diags := block.Body.JustAttributes()
	var reqs []*configs.ProviderRequirement
	for name, attr := range attrs {
		req, reqDiags := decodeVersionConstraint(attr)
		diags = append(diags, reqDiags...)
		if !diags.HasErrors() {
			reqs = append(reqs, &configs.ProviderRequirement{
				Name:        name,
				Requirement: req,
			})
		}
	}
	return reqs, diags
}

var providerBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: "alias",
		},
		{
			Name: "version",
		},

		// Attribute names reserved for future expansion.
		{Name: "count"},
		{Name: "depends_on"},
		{Name: "for_each"},
		{Name: "source"},
	},
	Blocks: []hcl.BlockHeaderSchema{
		// _All_ of these are reserved for future expansion.
		{Type: "lifecycle"},
		{Type: "locals"},
	},
}

const badIdentifierDetail = "A name must start with a letter and may contain only letters, digits, underscores, and dashes."

// https://github.com/hashicorp/terraform/blob/v0.12.6/configs/named_values.go
func decodeVariableBlock(block *hcl.Block, override bool) (*configs.Variable, hcl.Diagnostics) {
	v := &configs.Variable{
		Name:      block.Labels[0],
		DeclRange: block.DefRange,
	}

	// Unless we're building an override, we'll set some defaults
	// which we might override with attributes below. We leave these
	// as zero-value in the override case so we can recognize whether
	// or not they are set when we merge.
	if !override {
		v.Type = cty.DynamicPseudoType
		v.ParsingMode = configs.VariableParseLiteral
	}

	content, diags := block.Body.Content(variableBlockSchema)

	if !hclsyntax.ValidIdentifier(v.Name) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid variable name",
			Detail:   badIdentifierDetail,
			Subject:  &block.LabelRanges[0],
		})
	}

	// Don't allow declaration of variables that would conflict with the
	// reserved attribute and block type names in a "module" block, since
	// these won't be usable for child modules.
	for _, attr := range moduleBlockSchema.Attributes {
		if attr.Name == v.Name {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid variable name",
				Detail:   fmt.Sprintf("The variable name %q is reserved due to its special meaning inside module blocks.", attr.Name),
				Subject:  &block.LabelRanges[0],
			})
		}
	}
	for _, blockS := range moduleBlockSchema.Blocks {
		if blockS.Type == v.Name {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid variable name",
				Detail:   fmt.Sprintf("The variable name %q is reserved due to its special meaning inside module blocks.", blockS.Type),
				Subject:  &block.LabelRanges[0],
			})
		}
	}

	if attr, exists := content.Attributes["description"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &v.Description)
		diags = append(diags, valDiags...)
		v.DescriptionSet = true
	}

	if attr, exists := content.Attributes["type"]; exists {
		ty, parseMode, tyDiags := decodeVariableType(attr.Expr)
		diags = append(diags, tyDiags...)
		v.Type = ty
		v.ParsingMode = parseMode
	}

	if attr, exists := content.Attributes["default"]; exists {
		val, valDiags := attr.Expr.Value(nil)
		diags = append(diags, valDiags...)

		// Convert the default to the expected type so we can catch invalid
		// defaults early and allow later code to assume validity.
		// Note that this depends on us having already processed any "type"
		// attribute above.
		// However, we can't do this if we're in an override file where
		// the type might not be set; we'll catch that during merge.
		if v.Type != cty.NilType {
			var err error
			val, err = convert.Convert(val, v.Type)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid default value for variable",
					Detail:   fmt.Sprintf("This default value is not compatible with the variable's type constraint: %s.", err),
					Subject:  attr.Expr.Range().Ptr(),
				})
				val = cty.DynamicVal
			}
		}

		v.Default = val
	}

	return v, diags
}

func decodeVariableType(expr hcl.Expression) (cty.Type, configs.VariableParsingMode, hcl.Diagnostics) {
	if exprIsNativeQuotedString(expr) {
		// Here we're accepting the pre-0.12 form of variable type argument where
		// the string values "string", "list" and "map" are accepted has a hint
		// about the type used primarily for deciding how to parse values
		// given on the command line and in environment variables.
		// Only the native syntax ends up in this codepath; we handle the
		// JSON syntax (which is, of course, quoted even in the new format)
		// in the normal codepath below.
		val, diags := expr.Value(nil)
		if diags.HasErrors() {
			return cty.DynamicPseudoType, configs.VariableParseHCL, diags
		}
		str := val.AsString()
		switch str {
		case "string":
			return cty.String, configs.VariableParseLiteral, diags
		case "list":
			return cty.List(cty.DynamicPseudoType), configs.VariableParseHCL, diags
		case "map":
			return cty.Map(cty.DynamicPseudoType), configs.VariableParseHCL, diags
		default:
			return cty.DynamicPseudoType, configs.VariableParseHCL, hcl.Diagnostics{{
				Severity: hcl.DiagError,
				Summary:  "Invalid legacy variable type hint",
				Detail:   `The legacy variable type hint form, using a quoted string, allows only the values "string", "list", and "map". To provide a full type expression, remove the surrounding quotes and give the type expression directly.`,
				Subject:  expr.Range().Ptr(),
			}}
		}
	}

	// First we'll deal with some shorthand forms that the HCL-level type
	// expression parser doesn't include. These both emulate pre-0.12 behavior
	// of allowing a list or map of any element type as long as all of the
	// elements are consistent. This is the same as list(any) or map(any).
	switch hcl.ExprAsKeyword(expr) {
	case "list":
		return cty.List(cty.DynamicPseudoType), configs.VariableParseHCL, nil
	case "map":
		return cty.Map(cty.DynamicPseudoType), configs.VariableParseHCL, nil
	}

	ty, diags := typeexpr.TypeConstraint(expr)
	if diags.HasErrors() {
		return cty.DynamicPseudoType, configs.VariableParseHCL, diags
	}

	switch {
	case ty.IsPrimitiveType():
		// Primitive types use literal parsing.
		return ty, configs.VariableParseLiteral, diags
	default:
		// Everything else uses HCL parsing
		return ty, configs.VariableParseHCL, diags
	}
}

func decodeLocalsBlock(block *hcl.Block) ([]*configs.Local, hcl.Diagnostics) {
	attrs, diags := block.Body.JustAttributes()
	if len(attrs) == 0 {
		return nil, diags
	}

	locals := make([]*configs.Local, 0, len(attrs))
	for name, attr := range attrs {
		if !hclsyntax.ValidIdentifier(name) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid local value name",
				Detail:   badIdentifierDetail,
				Subject:  &attr.NameRange,
			})
		}

		locals = append(locals, &configs.Local{
			Name:      name,
			Expr:      attr.Expr,
			DeclRange: attr.Range,
		})
	}
	return locals, diags
}

func decodeOutputBlock(block *hcl.Block, override bool) (*configs.Output, hcl.Diagnostics) {
	o := &configs.Output{
		Name:      block.Labels[0],
		DeclRange: block.DefRange,
	}

	schema := outputBlockSchema
	if override {
		schema = schemaForOverrides(schema)
	}

	content, diags := block.Body.Content(schema)

	if !hclsyntax.ValidIdentifier(o.Name) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid output name",
			Detail:   badIdentifierDetail,
			Subject:  &block.LabelRanges[0],
		})
	}

	if attr, exists := content.Attributes["description"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &o.Description)
		diags = append(diags, valDiags...)
		o.DescriptionSet = true
	}

	if attr, exists := content.Attributes["value"]; exists {
		o.Expr = attr.Expr
	}

	if attr, exists := content.Attributes["sensitive"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &o.Sensitive)
		diags = append(diags, valDiags...)
		o.SensitiveSet = true
	}

	if attr, exists := content.Attributes["depends_on"]; exists {
		deps, depsDiags := decodeDependsOn(attr)
		diags = append(diags, depsDiags...)
		o.DependsOn = append(o.DependsOn, deps...)
	}

	return o, diags
}

var variableBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: "description",
		},
		{
			Name: "default",
		},
		{
			Name: "type",
		},
	},
}

var outputBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: "description",
		},
		{
			Name:     "value",
			Required: true,
		},
		{
			Name: "depends_on",
		},
		{
			Name: "sensitive",
		},
	},
}

// https://github.com/hashicorp/terraform/blob/v0.12.6/configs/depends_on.go
func decodeDependsOn(attr *hcl.Attribute) ([]hcl.Traversal, hcl.Diagnostics) {
	var ret []hcl.Traversal
	exprs, diags := hcl.ExprList(attr.Expr)

	for _, expr := range exprs {
		expr, shimDiags := shimTraversalInString(expr, false)
		diags = append(diags, shimDiags...)

		traversal, travDiags := hcl.AbsTraversalForExpr(expr)
		diags = append(diags, travDiags...)
		if len(traversal) != 0 {
			ret = append(ret, traversal)
		}
	}

	return ret, diags
}

// https://github.com/hashicorp/terraform/blob/v0.12.6/configs/module_call.go
func decodeModuleBlock(block *hcl.Block, override bool) (*configs.ModuleCall, hcl.Diagnostics) {
	mc := &configs.ModuleCall{
		Name:      block.Labels[0],
		DeclRange: block.DefRange,
	}

	schema := moduleBlockSchema
	if override {
		schema = schemaForOverrides(schema)
	}

	content, remain, diags := block.Body.PartialContent(schema)
	mc.Config = remain

	if !hclsyntax.ValidIdentifier(mc.Name) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid module instance name",
			Detail:   badIdentifierDetail,
			Subject:  &block.LabelRanges[0],
		})
	}

	if attr, exists := content.Attributes["source"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &mc.SourceAddr)
		diags = append(diags, valDiags...)
		mc.SourceAddrRange = attr.Expr.Range()
		mc.SourceSet = true
	}

	if attr, exists := content.Attributes["version"]; exists {
		var versionDiags hcl.Diagnostics
		mc.Version, versionDiags = decodeVersionConstraint(attr)
		diags = append(diags, versionDiags...)
	}

	if attr, exists := content.Attributes["count"]; exists {
		mc.Count = attr.Expr

		// We currently parse this, but don't yet do anything with it.
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Reserved argument name in module block",
			Detail:   fmt.Sprintf("The name %q is reserved for use in a future version of Terraform.", attr.Name),
			Subject:  &attr.NameRange,
		})
	}

	if attr, exists := content.Attributes["for_each"]; exists {
		mc.ForEach = attr.Expr

		// We currently parse this, but don't yet do anything with it.
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Reserved argument name in module block",
			Detail:   fmt.Sprintf("The name %q is reserved for use in a future version of Terraform.", attr.Name),
			Subject:  &attr.NameRange,
		})
	}

	if attr, exists := content.Attributes["depends_on"]; exists {
		deps, depsDiags := decodeDependsOn(attr)
		diags = append(diags, depsDiags...)
		mc.DependsOn = append(mc.DependsOn, deps...)

		// We currently parse this, but don't yet do anything with it.
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Reserved argument name in module block",
			Detail:   fmt.Sprintf("The name %q is reserved for use in a future version of Terraform.", attr.Name),
			Subject:  &attr.NameRange,
		})
	}

	if attr, exists := content.Attributes["providers"]; exists {
		seen := make(map[string]hcl.Range)
		pairs, pDiags := hcl.ExprMap(attr.Expr)
		diags = append(diags, pDiags...)
		for _, pair := range pairs {
			key, keyDiags := decodeProviderConfigRef(pair.Key, "providers")
			diags = append(diags, keyDiags...)
			value, valueDiags := decodeProviderConfigRef(pair.Value, "providers")
			diags = append(diags, valueDiags...)
			if keyDiags.HasErrors() || valueDiags.HasErrors() {
				continue
			}

			matchKey := key.String()
			if prev, exists := seen[matchKey]; exists {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Duplicate provider address",
					Detail:   fmt.Sprintf("A provider configuration was already passed to %s at %s. Each child provider configuration can be assigned only once.", matchKey, prev),
					Subject:  pair.Value.Range().Ptr(),
				})
				continue
			}

			rng := hcl.RangeBetween(pair.Key.Range(), pair.Value.Range())
			seen[matchKey] = rng
			mc.Providers = append(mc.Providers, configs.PassedProviderConfig{
				InChild:  key,
				InParent: value,
			})
		}
	}

	// Reserved block types (all of them)
	for _, block := range content.Blocks {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Reserved block type name in module block",
			Detail:   fmt.Sprintf("The block type name %q is reserved for use by Terraform in a future version.", block.Type),
			Subject:  &block.TypeRange,
		})
	}

	return mc, diags
}

var moduleBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "source",
			Required: true,
		},
		{
			Name: "version",
		},
		{
			Name: "count",
		},
		{
			Name: "for_each",
		},
		{
			Name: "depends_on",
		},
		{
			Name: "providers",
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		// These are all reserved for future use.
		{Type: "lifecycle"},
		{Type: "locals"},
		{Type: "provider", LabelNames: []string{"type"}},
	},
}

// https://github.com/hashicorp/terraform/blob/v0.12.6/configs/resource.go
func decodeResourceBlock(block *hcl.Block) (*configs.Resource, hcl.Diagnostics) {
	r := &configs.Resource{
		Mode:      addrs.ManagedResourceMode,
		Type:      block.Labels[0],
		Name:      block.Labels[1],
		DeclRange: block.DefRange,
		TypeRange: block.LabelRanges[0],
		Managed:   &configs.ManagedResource{},
	}

	content, remain, diags := block.Body.PartialContent(resourceBlockSchema)
	r.Config = remain

	if !hclsyntax.ValidIdentifier(r.Type) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid resource type name",
			Detail:   badIdentifierDetail,
			Subject:  &block.LabelRanges[0],
		})
	}
	if !hclsyntax.ValidIdentifier(r.Name) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid resource name",
			Detail:   badIdentifierDetail,
			Subject:  &block.LabelRanges[1],
		})
	}

	if attr, exists := content.Attributes["count"]; exists {
		r.Count = attr.Expr
	}

	if attr, exists := content.Attributes["for_each"]; exists {
		r.ForEach = attr.Expr
		// Cannot have count and for_each on the same resource block
		if r.Count != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  `Invalid combination of "count" and "for_each"`,
				Detail:   `The "count" and "for_each" meta-arguments are mutually-exclusive, only one should be used to be explicit about the number of resources to be created.`,
				Subject:  &attr.NameRange,
			})
		}
	}

	if attr, exists := content.Attributes["provider"]; exists {
		var providerDiags hcl.Diagnostics
		r.ProviderConfigRef, providerDiags = decodeProviderConfigRef(attr.Expr, "provider")
		diags = append(diags, providerDiags...)
	}

	if attr, exists := content.Attributes["depends_on"]; exists {
		deps, depsDiags := decodeDependsOn(attr)
		diags = append(diags, depsDiags...)
		r.DependsOn = append(r.DependsOn, deps...)
	}

	var seenLifecycle *hcl.Block
	var seenConnection *hcl.Block
	for _, block := range content.Blocks {
		switch block.Type {
		case "lifecycle":
			if seenLifecycle != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Duplicate lifecycle block",
					Detail:   fmt.Sprintf("This resource already has a lifecycle block at %s.", seenLifecycle.DefRange),
					Subject:  &block.DefRange,
				})
				continue
			}
			seenLifecycle = block

			lcContent, lcDiags := block.Body.Content(resourceLifecycleBlockSchema)
			diags = append(diags, lcDiags...)

			if attr, exists := lcContent.Attributes["create_before_destroy"]; exists {
				valDiags := gohcl.DecodeExpression(attr.Expr, nil, &r.Managed.CreateBeforeDestroy)
				diags = append(diags, valDiags...)
				r.Managed.CreateBeforeDestroySet = true
			}

			if attr, exists := lcContent.Attributes["prevent_destroy"]; exists {
				valDiags := gohcl.DecodeExpression(attr.Expr, nil, &r.Managed.PreventDestroy)
				diags = append(diags, valDiags...)
				r.Managed.PreventDestroySet = true
			}

			if attr, exists := lcContent.Attributes["ignore_changes"]; exists {

				// ignore_changes can either be a list of relative traversals
				// or it can be just the keyword "all" to ignore changes to this
				// resource entirely.
				//   ignore_changes = [ami, instance_type]
				//   ignore_changes = all
				// We also allow two legacy forms for compatibility with earlier
				// versions:
				//   ignore_changes = ["ami", "instance_type"]
				//   ignore_changes = ["*"]

				kw := hcl.ExprAsKeyword(attr.Expr)

				switch {
				case kw == "all":
					r.Managed.IgnoreAllChanges = true
				default:
					exprs, listDiags := hcl.ExprList(attr.Expr)
					diags = append(diags, listDiags...)

					var ignoreAllRange hcl.Range

					for _, expr := range exprs {

						// our expr might be the literal string "*", which
						// we accept as a deprecated way of saying "all".
						if shimIsIgnoreChangesStar(expr) {
							r.Managed.IgnoreAllChanges = true
							ignoreAllRange = expr.Range()
							diags = append(diags, &hcl.Diagnostic{
								Severity: hcl.DiagWarning,
								Summary:  "Deprecated ignore_changes wildcard",
								Detail:   "The [\"*\"] form of ignore_changes wildcard is deprecated. Use \"ignore_changes = all\" to ignore changes to all attributes.",
								Subject:  attr.Expr.Range().Ptr(),
							})
							continue
						}

						expr, shimDiags := shimTraversalInString(expr, false)
						diags = append(diags, shimDiags...)

						traversal, travDiags := hcl.RelTraversalForExpr(expr)
						diags = append(diags, travDiags...)
						if len(traversal) != 0 {
							r.Managed.IgnoreChanges = append(r.Managed.IgnoreChanges, traversal)
						}
					}

					if r.Managed.IgnoreAllChanges && len(r.Managed.IgnoreChanges) != 0 {
						diags = append(diags, &hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  "Invalid ignore_changes ruleset",
							Detail:   "Cannot mix wildcard string \"*\" with non-wildcard references.",
							Subject:  &ignoreAllRange,
							Context:  attr.Expr.Range().Ptr(),
						})
					}

				}

			}

		case "connection":
			if seenConnection != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Duplicate connection block",
					Detail:   fmt.Sprintf("This resource already has a connection block at %s.", seenConnection.DefRange),
					Subject:  &block.DefRange,
				})
				continue
			}
			seenConnection = block

			r.Managed.Connection = &configs.Connection{
				Config:    block.Body,
				DeclRange: block.DefRange,
			}

		case "provisioner":
			pv, pvDiags := decodeProvisionerBlock(block)
			diags = append(diags, pvDiags...)
			if pv != nil {
				r.Managed.Provisioners = append(r.Managed.Provisioners, pv)
			}

		default:
			// Any other block types are ones we've reserved for future use,
			// so they get a generic message.
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Reserved block type name in resource block",
				Detail:   fmt.Sprintf("The block type name %q is reserved for use by Terraform in a future version.", block.Type),
				Subject:  &block.TypeRange,
			})
		}
	}

	return r, diags
}

func decodeDataBlock(block *hcl.Block) (*configs.Resource, hcl.Diagnostics) {
	r := &configs.Resource{
		Mode:      addrs.DataResourceMode,
		Type:      block.Labels[0],
		Name:      block.Labels[1],
		DeclRange: block.DefRange,
		TypeRange: block.LabelRanges[0],
	}

	content, remain, diags := block.Body.PartialContent(dataBlockSchema)
	r.Config = remain

	if !hclsyntax.ValidIdentifier(r.Type) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid data source name",
			Detail:   badIdentifierDetail,
			Subject:  &block.LabelRanges[0],
		})
	}
	if !hclsyntax.ValidIdentifier(r.Name) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid data resource name",
			Detail:   badIdentifierDetail,
			Subject:  &block.LabelRanges[1],
		})
	}

	if attr, exists := content.Attributes["count"]; exists {
		r.Count = attr.Expr
	}

	if attr, exists := content.Attributes["for_each"]; exists {
		r.ForEach = attr.Expr
		// Cannot have count and for_each on the same data block
		if r.Count != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  `Invalid combination of "count" and "for_each"`,
				Detail:   `The "count" and "for_each" meta-arguments are mutually-exclusive, only one should be used to be explicit about the number of resources to be created.`,
				Subject:  &attr.NameRange,
			})
		}
	}

	if attr, exists := content.Attributes["provider"]; exists {
		var providerDiags hcl.Diagnostics
		r.ProviderConfigRef, providerDiags = decodeProviderConfigRef(attr.Expr, "provider")
		diags = append(diags, providerDiags...)
	}

	if attr, exists := content.Attributes["depends_on"]; exists {
		deps, depsDiags := decodeDependsOn(attr)
		diags = append(diags, depsDiags...)
		r.DependsOn = append(r.DependsOn, deps...)
	}

	for _, block := range content.Blocks {
		// All of the block types we accept are just reserved for future use, but some get a specialized error message.
		switch block.Type {
		case "lifecycle":
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unsupported lifecycle block",
				Detail:   "Data resources do not have lifecycle settings, so a lifecycle block is not allowed.",
				Subject:  &block.DefRange,
			})
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Reserved block type name in data block",
				Detail:   fmt.Sprintf("The block type name %q is reserved for use by Terraform in a future version.", block.Type),
				Subject:  &block.TypeRange,
			})
		}
	}

	return r, diags
}

func decodeProviderConfigRef(expr hcl.Expression, argName string) (*configs.ProviderConfigRef, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	var shimDiags hcl.Diagnostics
	expr, shimDiags = shimTraversalInString(expr, false)
	diags = append(diags, shimDiags...)

	traversal, travDiags := hcl.AbsTraversalForExpr(expr)

	// AbsTraversalForExpr produces only generic errors, so we'll discard
	// the errors given and produce our own with extra context. If we didn't
	// get any errors then we might still have warnings, though.
	if !travDiags.HasErrors() {
		diags = append(diags, travDiags...)
	}

	if len(traversal) < 1 || len(traversal) > 2 {
		// A provider reference was given as a string literal in the legacy
		// configuration language and there are lots of examples out there
		// showing that usage, so we'll sniff for that situation here and
		// produce a specialized error message for it to help users find
		// the new correct form.
		if exprIsNativeQuotedString(expr) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid provider configuration reference",
				Detail:   "A provider configuration reference must not be given in quotes.",
				Subject:  expr.Range().Ptr(),
			})
			return nil, diags
		}

		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid provider configuration reference",
			Detail:   fmt.Sprintf("The %s argument requires a provider type name, optionally followed by a period and then a configuration alias.", argName),
			Subject:  expr.Range().Ptr(),
		})
		return nil, diags
	}

	ret := &configs.ProviderConfigRef{
		Name:      traversal.RootName(),
		NameRange: traversal[0].SourceRange(),
	}

	if len(traversal) > 1 {
		aliasStep, ok := traversal[1].(hcl.TraverseAttr)
		if !ok {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid provider configuration reference",
				Detail:   "Provider name must either stand alone or be followed by a period and then a configuration alias.",
				Subject:  traversal[1].SourceRange().Ptr(),
			})
			return ret, diags
		}

		ret.Alias = aliasStep.Name
		ret.AliasRange = aliasStep.SourceRange().Ptr()
	}

	return ret, diags
}

var commonResourceAttributes = []hcl.AttributeSchema{
	{
		Name: "count",
	},
	{
		Name: "for_each",
	},
	{
		Name: "provider",
	},
	{
		Name: "depends_on",
	},
}

var resourceBlockSchema = &hcl.BodySchema{
	Attributes: commonResourceAttributes,
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "locals"}, // reserved for future use
		{Type: "lifecycle"},
		{Type: "connection"},
		{Type: "provisioner", LabelNames: []string{"type"}},
	},
}

var dataBlockSchema = &hcl.BodySchema{
	Attributes: commonResourceAttributes,
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "lifecycle"}, // reserved for future use
		{Type: "locals"},    // reserved for future use
	},
}

var resourceLifecycleBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: "create_before_destroy",
		},
		{
			Name: "prevent_destroy",
		},
		{
			Name: "ignore_changes",
		},
	},
}

// https://github.com/hashicorp/terraform/blob/v0.12.6/configs/provisioner.go
func decodeProvisionerBlock(block *hcl.Block) (*configs.Provisioner, hcl.Diagnostics) {
	pv := &configs.Provisioner{
		Type:      block.Labels[0],
		TypeRange: block.LabelRanges[0],
		DeclRange: block.DefRange,
		When:      configs.ProvisionerWhenCreate,
		OnFailure: configs.ProvisionerOnFailureFail,
	}

	content, config, diags := block.Body.PartialContent(provisionerBlockSchema)
	pv.Config = config

	if attr, exists := content.Attributes["when"]; exists {
		expr, shimDiags := shimTraversalInString(attr.Expr, true)
		diags = append(diags, shimDiags...)

		switch hcl.ExprAsKeyword(expr) {
		case "create":
			pv.When = configs.ProvisionerWhenCreate
		case "destroy":
			pv.When = configs.ProvisionerWhenDestroy
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid \"when\" keyword",
				Detail:   "The \"when\" argument requires one of the following keywords: create or destroy.",
				Subject:  expr.Range().Ptr(),
			})
		}
	}

	if attr, exists := content.Attributes["on_failure"]; exists {
		expr, shimDiags := shimTraversalInString(attr.Expr, true)
		diags = append(diags, shimDiags...)

		switch hcl.ExprAsKeyword(expr) {
		case "continue":
			pv.OnFailure = configs.ProvisionerOnFailureContinue
		case "fail":
			pv.OnFailure = configs.ProvisionerOnFailureFail
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid \"on_failure\" keyword",
				Detail:   "The \"on_failure\" argument requires one of the following keywords: continue or fail.",
				Subject:  attr.Expr.Range().Ptr(),
			})
		}
	}

	var seenConnection *hcl.Block
	for _, block := range content.Blocks {
		switch block.Type {

		case "connection":
			if seenConnection != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Duplicate connection block",
					Detail:   fmt.Sprintf("This provisioner already has a connection block at %s.", seenConnection.DefRange),
					Subject:  &block.DefRange,
				})
				continue
			}
			seenConnection = block

			//conn, connDiags := decodeConnectionBlock(block)
			//diags = append(diags, connDiags...)
			pv.Connection = &configs.Connection{
				Config:    block.Body,
				DeclRange: block.DefRange,
			}

		default:
			// Any other block types are ones we've reserved for future use,
			// so they get a generic message.
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Reserved block type name in provisioner block",
				Detail:   fmt.Sprintf("The block type name %q is reserved for use by Terraform in a future version.", block.Type),
				Subject:  &block.TypeRange,
			})
		}
	}

	return pv, diags
}

var provisionerBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "when"},
		{Name: "on_failure"},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "connection"},
		{Type: "lifecycle"}, // reserved for future use
	},
}

// https://github.com/hashicorp/terraform/blob/v0.12.6/configs/util.go#L15
func exprIsNativeQuotedString(expr hcl.Expression) bool {
	_, ok := expr.(*hclsyntax.TemplateExpr)
	return ok
}

func schemaForOverrides(schema *hcl.BodySchema) *hcl.BodySchema {
	ret := &hcl.BodySchema{
		Attributes: make([]hcl.AttributeSchema, len(schema.Attributes)),
		Blocks:     schema.Blocks,
	}

	for i, attrS := range schema.Attributes {
		ret.Attributes[i] = attrS
		ret.Attributes[i].Required = false
	}

	return ret
}

// https://github.com/hashicorp/terraform/blob/v0.12.6/configs/compat_shim.go
func shimTraversalInString(expr hcl.Expression, wantKeyword bool) (hcl.Expression, hcl.Diagnostics) {
	// ObjectConsKeyExpr is a special wrapper type used for keys on object
	// constructors to deal with the fact that naked identifiers are normally
	// handled as "bareword" strings rather than as variable references. Since
	// we know we're interpreting as a traversal anyway (and thus it won't
	// matter whether it's a string or an identifier) we can safely just unwrap
	// here and then process whatever we find inside as normal.
	if ocke, ok := expr.(*hclsyntax.ObjectConsKeyExpr); ok {
		expr = ocke.Wrapped
	}

	if !exprIsNativeQuotedString(expr) {
		return expr, nil
	}

	strVal, diags := expr.Value(nil)
	if diags.HasErrors() || strVal.IsNull() || !strVal.IsKnown() {
		// Since we're not even able to attempt a shim here, we'll discard
		// the diagnostics we saw so far and let the caller's own error
		// handling take care of reporting the invalid expression.
		return expr, nil
	}

	// The position handling here isn't _quite_ right because it won't
	// take into account any escape sequences in the literal string, but
	// it should be close enough for any error reporting to make sense.
	srcRange := expr.Range()
	startPos := srcRange.Start // copy
	startPos.Column++          // skip initial quote
	startPos.Byte++            // skip initial quote

	traversal, tDiags := hclsyntax.ParseTraversalAbs(
		[]byte(strVal.AsString()),
		srcRange.Filename,
		startPos,
	)
	diags = append(diags, tDiags...)

	// For initial release our deprecation warnings are disabled to allow
	// a period where modules can be compatible with both old and new
	// conventions.
	// FIXME: Re-enable these deprecation warnings in a release prior to
	// Terraform 0.13 and then remove the shims altogether for 0.13.
	/*
		if wantKeyword {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  "Quoted keywords are deprecated",
				Detail:   "In this context, keywords are expected literally rather than in quotes. Previous versions of Terraform required quotes, but that usage is now deprecated. Remove the quotes surrounding this keyword to silence this warning.",
				Subject:  &srcRange,
			})
		} else {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  "Quoted references are deprecated",
				Detail:   "In this context, references are expected literally rather than in quotes. Previous versions of Terraform required quotes, but that usage is now deprecated. Remove the quotes surrounding this reference to silence this warning.",
				Subject:  &srcRange,
			})
		}
	*/

	return &hclsyntax.ScopeTraversalExpr{
		Traversal: traversal,
		SrcRange:  srcRange,
	}, diags
}

func shimIsIgnoreChangesStar(expr hcl.Expression) bool {
	val, valDiags := expr.Value(nil)
	if valDiags.HasErrors() {
		return false
	}
	if val.Type() != cty.String || val.IsNull() || !val.IsKnown() {
		return false
	}
	return val.AsString() == "*"
}
