package terraform

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// Runner is a custom runner that provides helper functions for this ruleset.
type Runner struct {
	tflint.Runner
}

// NewRunner returns a new custom runner.
func NewRunner(runner tflint.Runner) *Runner {
	return &Runner{Runner: runner}
}

// GetModuleCalls returns all "module" blocks, including uncreated module calls.
func (r *Runner) GetModuleCalls() ([]*ModuleCall, hcl.Diagnostics) {
	calls := []*ModuleCall{}
	diags := hcl.Diagnostics{}

	body, err := r.GetModuleContent(&hclext.BodySchema{
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
	}, &tflint.GetModuleContentOption{IncludeNotCreated: true})
	if err != nil {
		return calls, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "failed to call GetModuleContent()",
				Detail:   err.Error(),
			},
		}
	}

	for _, block := range body.Blocks {
		call, decodeDiags := decodeModuleCall(block)
		diags = diags.Extend(decodeDiags)
		if decodeDiags.HasErrors() {
			continue
		}
		calls = append(calls, call)
	}

	return calls, diags
}

// GetLocals returns all entries in "locals" blocks.
func (r *Runner) GetLocals() (map[string]*Local, hcl.Diagnostics) {
	locals := map[string]*Local{}
	diags := hcl.Diagnostics{}

	files, err := r.GetFiles()
	if err != nil {
		return locals, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "failed to call GetFiles()",
				Detail:   err.Error(),
			},
		}
	}

	for _, file := range files {
		content, _, schemaDiags := file.Body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{{Type: "locals"}},
		})
		diags = diags.Extend(schemaDiags)
		if schemaDiags.HasErrors() {
			continue
		}

		for _, block := range content.Blocks {
			attrs, localsDiags := block.Body.JustAttributes()
			diags = diags.Extend(localsDiags)
			if localsDiags.HasErrors() {
				continue
			}

			for name, attr := range attrs {
				locals[name] = &Local{
					Name:     attr.Name,
					DefRange: attr.Range,
				}
			}
		}
	}

	return locals, diags
}

// GetProviderRefs returns all references to providers in resources, data, provider declarations, and module calls.
func (r *Runner) GetProviderRefs() (map[string]*ProviderRef, hcl.Diagnostics) {
	providerRefs := map[string]*ProviderRef{}

	body, err := r.GetModuleContent(&hclext.BodySchema{
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
				Summary:  "failed to call `GetModuleContent()`",
				Detail:   err.Error(),
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
				if decodeDiags.HasErrors() {
					continue
				}
				providerRefs[ref.Name] = ref
			} else {
				providerName := block.Labels[0]
				if under := strings.Index(providerName, "_"); under != -1 {
					providerName = providerName[:under]
				}
				providerRefs[providerName] = &ProviderRef{
					Name:     providerName,
					DefRange: block.DefRange,
				}
			}
		case "provider":
			providerRefs[block.Labels[0]] = &ProviderRef{
				Name:     block.Labels[0],
				DefRange: block.DefRange,
			}
		case "module":
			if attr, exists := block.Body.Attributes["providers"]; exists {
				pairs, mapDiags := hcl.ExprMap(attr.Expr)
				diags = diags.Extend(mapDiags)
				if mapDiags.HasErrors() {
					continue
				}

				for _, pair := range pairs {
					ref, decodeDiags := decodeProviderRef(pair.Value, block.DefRange)
					diags = diags.Extend(decodeDiags)
					if decodeDiags.HasErrors() {
						continue
					}
					providerRefs[ref.Name] = ref
				}
			}
		}
	}

	return providerRefs, nil
}
