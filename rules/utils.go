package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// WalkExpressions visits all expressions, including those in the file before merging.
// Note that it behaves differently in native HCL syntax and JSON syntax.
// In the HCL syntax, expressions in expressions, such as list and object are passed to
// the walker function. The walker should check the type of the expression.
// In the JSON syntax, only an expression of an attribute seen from the top level of the file
// is passed, not expressions in expressions to the walker. This is an API limitation of JSON syntax.
//
// XXX: Should be moved to SDK?
func WalkExpressions(r tflint.Runner, walker func(hcl.Expression) error) error {
	visit := func(node hclsyntax.Node) hcl.Diagnostics {
		if expr, ok := node.(hcl.Expression); ok {
			if err := walker(expr); err != nil {
				// FIXME: walker should returns hcl.Diagnostics directly
				return hcl.Diagnostics{
					{
						Severity: hcl.DiagError,
						Summary:  err.Error(),
					},
				}
			}
		}
		return hcl.Diagnostics{}
	}

	files, err := r.GetFiles()
	if err != nil {
		return err
	}
	for _, file := range files {
		if body, ok := file.Body.(*hclsyntax.Body); ok {
			diags := hclsyntax.VisitAll(body, visit)
			if diags.HasErrors() {
				return diags
			}
			continue
		}

		// In JSON syntax, everything can be walked as an attribute.
		attrs, diags := file.Body.JustAttributes()
		if diags.HasErrors() {
			return diags
		}

		for _, attr := range attrs {
			if err := walker(attr.Expr); err != nil {
				return err
			}
		}
	}

	return nil
}

type moduleCall struct {
	name        string
	defRange    hcl.Range
	source      string
	sourceAttr  *hclext.Attribute
	version     version.Constraints
	versionAttr *hclext.Attribute
}

func decodeModuleCall(block *hclext.Block) (*moduleCall, hcl.Diagnostics) {
	module := &moduleCall{
		name:     block.Labels[0],
		defRange: block.DefRange,
	}
	diags := hcl.Diagnostics{}

	if source, exists := block.Body.Attributes["source"]; exists {
		module.sourceAttr = source
		sourceDiags := gohcl.DecodeExpression(source.Expr, nil, &module.source)
		diags = diags.Extend(sourceDiags)
	}

	if versionAttr, exists := block.Body.Attributes["version"]; exists {
		module.versionAttr = versionAttr

		var versionVal string
		versionDiags := gohcl.DecodeExpression(versionAttr.Expr, nil, &versionVal)
		diags = diags.Extend(versionDiags)
		if diags.HasErrors() {
			return module, diags
		}

		constraints, err := version.NewConstraint(versionVal)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid version constraint",
				Detail:   "This string does not use correct version constraint syntax.",
				Subject:  versionAttr.Expr.Range().Ptr(),
			})
		}
		module.version = constraints
	}

	return module, diags
}

var moduleCallSchema = &hclext.BodySchema{
	Blocks: []hclext.BlockSchema{
		{
			Type:       "module",
			LabelNames: []string{"name"},
			Body: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{
					{Name: "source"},
					{Name: "version"},
				},
			},
		},
	},
}

type local struct {
	name     string
	defRange hcl.Range
}

func getLocals(runner tflint.Runner) (map[string]*local, hcl.Diagnostics) {
	locals := map[string]*local{}
	diags := hcl.Diagnostics{}

	files, err := runner.GetFiles()
	if err != nil {
		// XXX: Should we return error or diagnostics?
		return locals, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  err.Error(),
			},
		}
	}

	for _, file := range files {
		content, _, schemaDiags := file.Body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{{Type: "locals"}},
		})
		diags = diags.Extend(schemaDiags)
		if diags.HasErrors() {
			continue
		}

		for _, block := range content.Blocks {
			attrs, localsDiags := block.Body.JustAttributes()
			diags = diags.Extend(localsDiags)
			if diags.HasErrors() {
				continue
			}

			for name, attr := range attrs {
				locals[name] = &local{
					name:     attr.Name,
					defRange: attr.Range,
				}
			}
		}
	}

	return locals, diags
}

type providerRef struct {
	name     string
	defRange hcl.Range
}

func getProviderRefs(runner tflint.Runner) (map[string]*providerRef, hcl.Diagnostics) {
	providerRefs := map[string]*providerRef{}

	body, err := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "resource",
				LabelNames: []string{"type", "name"},
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{
						{Name: "provider"},
					},
				},
			},
			{
				Type:       "data",
				LabelNames: []string{"type", "name"},
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{
						{Name: "provider"},
					},
				},
			},
			{
				Type:       "provider",
				LabelNames: []string{"name"},
				Body:       &hclext.BodySchema{},
			},
			{
				Type:       "module",
				LabelNames: []string{"name"},
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{
						{Name: "providers"},
					},
				},
			},
		},
	}, &tflint.GetModuleContentOption{IncludeNotCreated: true})
	if err != nil {
		return providerRefs, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  err.Error(),
			},
		}
	}

	var diags hcl.Diagnostics
	for _, block := range body.Blocks {
		switch block.Type {
		case "resource":
			fallthrough
		case "data":
			if attr, exists := block.Body.Attributes["provider"]; exists {
				ref, decodeDiags := decodeProviderRef(attr.Expr, block.DefRange)
				diags = diags.Extend(decodeDiags)
				if diags.HasErrors() {
					continue
				}
				providerRefs[ref.name] = ref
			} else {
				providerName := block.Labels[0]
				if under := strings.Index(providerName, "_"); under != -1 {
					providerName = providerName[:under]
				}
				providerRefs[providerName] = &providerRef{
					name:     providerName,
					defRange: block.DefRange,
				}
			}
		case "provider":
			providerRefs[block.Labels[0]] = &providerRef{
				name:     block.Labels[0],
				defRange: block.DefRange,
			}
		case "module":
			if attr, exists := block.Body.Attributes["providers"]; exists {
				pairs, mapDiags := hcl.ExprMap(attr.Expr)
				diags = diags.Extend(mapDiags)
				if diags.HasErrors() {
					continue
				}

				for _, pair := range pairs {
					ref, decodeDiags := decodeProviderRef(pair.Value, block.DefRange)
					diags = diags.Extend(decodeDiags)
					if diags.HasErrors() {
						continue
					}
					providerRefs[ref.name] = ref
				}
			}
		}
	}

	return providerRefs, nil
}

func decodeProviderRef(expr hcl.Expression, defRange hcl.Range) (*providerRef, hcl.Diagnostics) {
	expr, diags := shimTraversalInString(expr)
	if diags.HasErrors() {
		return nil, diags
	}

	traversal, diags := hcl.AbsTraversalForExpr(expr)
	if diags.HasErrors() {
		return nil, diags
	}

	return &providerRef{
		name:     traversal.RootName(),
		defRange: defRange,
	}, nil
}

type reference struct {
	subject referencable
}

type referencable interface {
	referenceableSigil()
}

type inputVariableReference struct {
	referencable
	name string
}

type localValueReference struct {
	referencable
	name string
}

type terraformReference struct {
	referencable
	name string
}

type dataResourceReference struct {
	referencable
	typeName string
	name     string
}

func (d *dataResourceReference) String() string {
	return fmt.Sprintf("data.%s.%s", d.typeName, d.name)
}

// referencesInExpr extracts and returns references from hcl.Expression.
// This is roughly equivalent to ReferencesInExpr in Terraform interlan API,
// but simply ignores invalid references without returing diagnostics.
// This is because it is not the responsibility of this method to detect invalid references.
// Invalid references should be caught by terraform validate.
//
// @see https://github.com/hashicorp/terraform/blob/v1.2.5/internal/lang/references.go#L75
func referencesInExpr(expr hcl.Expression) []*reference {
	references := []*reference{}

	for _, traversal := range expr.Variables() {
		name := traversal.RootName()

		switch name {
		case "var":
			name, diags := parseSingleAttrRef(traversal)
			if diags.HasErrors() {
				continue
			}
			references = append(references, &reference{subject: inputVariableReference{name: name}})
		case "local":
			name, diags := parseSingleAttrRef(traversal)
			if diags.HasErrors() {
				continue
			}
			references = append(references, &reference{subject: localValueReference{name: name}})
		case "terraform":
			name, diags := parseSingleAttrRef(traversal)
			if diags.HasErrors() {
				continue
			}
			references = append(references, &reference{subject: terraformReference{name: name}})
		case "data":
			if len(traversal) < 3 {
				// The "data" object must be followed by two attribute names: the data source type and the resource name.
				continue
			}

			var typeName string
			switch tt := traversal[1].(type) {
			case hcl.TraverseRoot:
				typeName = tt.Name
			case hcl.TraverseAttr:
				typeName = tt.Name
			default:
				// The "data" object does not support this operation.
				continue
			}

			attrTrav, ok := traversal[2].(hcl.TraverseAttr)
			if !ok {
				// A reference to a data source must be followed by at least one attribute access, specifying the resource name.
				continue
			}

			references = append(references, &reference{subject: dataResourceReference{typeName: typeName, name: attrTrav.Name}})
		}
	}

	return references
}

// @see https://github.com/hashicorp/terraform/blob/v1.2.5/internal/addrs/parse_ref.go#L392
func parseSingleAttrRef(traversal hcl.Traversal) (string, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	root := traversal.RootName()
	rootRange := traversal[0].SourceRange()

	if len(traversal) < 2 {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid reference",
			Detail:   fmt.Sprintf("The %q object cannot be accessed directly. Instead, access one of its attributes.", root),
			Subject:  &rootRange,
		})
		return "", diags
	}
	if attrTrav, ok := traversal[1].(hcl.TraverseAttr); ok {
		return attrTrav.Name, diags
	}
	diags = diags.Append(&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "Invalid reference",
		Detail:   fmt.Sprintf("The %q object does not support this operation.", root),
		Subject:  traversal[1].SourceRange().Ptr(),
	})
	return "", diags
}

// @see https://github.com/hashicorp/terraform/blob/v1.2.5/internal/configs/compat_shim.go#L34
func shimTraversalInString(expr hcl.Expression) (hcl.Expression, hcl.Diagnostics) {
	// ObjectConsKeyExpr is a special wrapper type used for keys on object
	// constructors to deal with the fact that naked identifiers are normally
	// handled as "bareword" strings rather than as variable references. Since
	// we know we're interpreting as a traversal anyway (and thus it won't
	// matter whether it's a string or an identifier) we can safely just unwrap
	// here and then process whatever we find inside as normal.
	if ocke, ok := expr.(*hclsyntax.ObjectConsKeyExpr); ok {
		expr = ocke.Wrapped
	}

	if _, ok := expr.(*hclsyntax.TemplateExpr); !ok {
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

	return &hclsyntax.ScopeTraversalExpr{
		Traversal: traversal,
		SrcRange:  srcRange,
	}, diags
}
