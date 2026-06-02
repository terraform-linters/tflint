// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package terraform

import (
	"fmt"
	"sort"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/zclconf/go-cty/cty"
)

// A Config is a node in the tree of modules within a configuration.
//
// The module tree is constructed by following ModuleCall instances recursively
// through the root module transitively into descendent modules.
type Config struct {
	// RootModule points to the Config for the root module within the same
	// module tree as this module. If this module _is_ the root module then
	// this is self-referential.
	Root *Config

	// Path is a sequence of module logical names that traverse from the root
	// module to this config. Path is empty for the root module.
	Path addrs.Module

	// ChildModules points to the Config for each of the direct child modules
	// called from this module. The keys in this map match the keys in
	// Module.ModuleCalls.
	Children map[string]*Config

	// Module points to the object describing the configuration for the
	// various elements (variables, resources, etc) defined by this module.
	Module *Module
}

// NewEmptyConfig constructs a single-node configuration tree with an empty
// root module. This is generally a pretty useless thing to do, so most callers
// should instead use BuildConfig.
func NewEmptyConfig() *Config {
	ret := &Config{}
	ret.Root = ret
	ret.Children = make(map[string]*Config)
	ret.Module = &Module{}
	return ret
}

// BuildConfig constructs a Config from a root module by loading all of its
// descendent modules via the given ModuleWalker.
//
// The result is a module tree that has so far only had basic module- and
// file-level invariants validated. If the returned diagnostics contains errors,
// the returned module tree may be incomplete but can still be used carefully
// for static analysis.
func BuildConfig(root *Module, walker ModuleWalker, originalWd string, variables ...InputValues) (*Config, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	cfg := &Config{
		Module:   root,
		Children: map[string]*Config{},
	}
	cfg.Root = cfg // Root module is self-referential.

	variableValues, diags := VariableValues(cfg, variables...)
	if diags.HasErrors() {
		return nil, diags
	}

	ctx := &Evaluator{
		Meta: &ContextMeta{
			Env:                Workspace(),
			OriginalWorkingDir: originalWd,
		},
		ModulePath:     cfg.Path.UnkeyedInstanceShim(),
		Config:         cfg.Root,
		VariableValues: variableValues,
	}
	cfg.Children, diags = buildChildModules(cfg, walker, ctx)

	return cfg, diags
}

func buildChildModules(parent *Config, walker ModuleWalker, ctx *Evaluator) (map[string]*Config, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	ret := map[string]*Config{}

	calls := parent.Module.ModuleCalls

	// We'll sort the calls by their local names so that they'll appear in a
	// predictable order in any logging that's produced during the walk.
	callNames := make([]string, 0, len(calls))
	for k := range calls {
		callNames = append(callNames, k)
	}
	sort.Strings(callNames)

	for _, callName := range callNames {
		call := calls[callName]
		path := make([]string, len(parent.Path)+1)
		copy(path, parent.Path)
		path[len(path)-1] = call.Name

		// Return an error for nesting too deep to avoid infinite loops due to circular references.
		if len(path) > 10 {
			return ret, diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Module stack level too deep",
				Detail:   fmt.Sprintf("This configuration has nested modules more than 10 levels deep. This is mainly caused by circular references. current path: %s", parent.Path),
				Subject:  &call.DeclRange,
			})
		}

		addr, sourceDiags := evalModuleSource(call, ctx)
		if sourceDiags.HasErrors() {
			diags = append(diags, sourceDiags...)
			continue
		}
		if addr == nil {
			// skip if the source address is not available,
			// which can happen when the source attribute is missing,
			// contains unknown values.
			continue
		}

		req := ModuleRequest{
			Name:       call.Name,
			Path:       path,
			SourceAddr: addr,
			Parent:     parent,
			CallRange:  call.DeclRange,
		}

		mod, _, modDiags := walker.LoadModule(&req)
		diags = append(diags, modDiags...)
		if mod == nil {
			// nil can be returned if the source address was invalid and so
			// nothing could be loaded whatsoever. LoadModule should've
			// returned at least one error diagnostic in that case.
			continue
		}

		child := &Config{
			Root:     parent.Root,
			Path:     path,
			Module:   mod,
			Children: map[string]*Config{},
		}
		// To perform evaluation in a child module context,
		// each loaded module must be set to the children immediately.
		parent.Children[call.Name] = child

		childCtx, ctxDiags := buildChildCtx(parent, ctx, call.Name, child)
		if ctxDiags.HasErrors() {
			diags = append(diags, ctxDiags...)
			continue
		}
		if childCtx == nil {
			// skip if the child context cannot be built,
			// which can happen when the module disappears after expansion (e.g. `count = 0`).
			continue
		}
		child.Children, modDiags = buildChildModules(child, walker, childCtx)
		diags = append(diags, modDiags...)

		ret[call.Name] = child
	}

	return ret, diags
}

func evalModuleSource(call *ModuleCall, ctx *Evaluator) (addrs.ModuleSource, hcl.Diagnostics) {
	// skip if the call doesn't have a source attribute
	if call.SourceExpr == nil {
		return nil, nil
	}

	val, diags := ctx.EvaluateExpr(call.SourceExpr, cty.String)
	if diags.HasErrors() {
		return nil, diags
	}
	// skip if the source is unknown, null, or sensitive.
	// In Terraform, only constant variables can be used in module source, so this is not an issue.
	if !val.IsKnown() || val.IsNull() || val.IsMarked() {
		return nil, nil
	}

	rawSource := val.AsString()
	addr, err := addrs.ParseModuleSource(rawSource)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid module source address",
			Detail:   fmt.Sprintf("Failed to parse module source address: %s", err),
			Subject:  call.SourceExpr.Range().Ptr(),
		})
		return nil, diags
	}

	// FIXME: Save the resolved source for ignore module checks. However, this relies on a complex call graph
	//        and is expected to be removed in the future.
	call.SourceResolved = rawSource

	return addr, diags
}

func buildChildCtx(parent *Config, parentCtx *Evaluator, callName string, child *Config) (*Evaluator, hcl.Diagnostics) {
	moduleCallSchema := &hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "module",
				LabelNames: []string{"name"},
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{},
				},
			},
		},
	}
	for _, v := range child.Module.Variables {
		attr := hclext.AttributeSchema{Name: v.Name}
		moduleCallSchema.Blocks[0].Body.Attributes = append(moduleCallSchema.Blocks[0].Body.Attributes, attr)
	}

	moduleCalls, diags := parent.Module.PartialContent(moduleCallSchema, parentCtx)
	if diags.HasErrors() {
		return nil, diags
	}
	var moduleCallBodies []*hclext.BodyContent
	for _, block := range moduleCalls.Blocks {
		if callName == block.Labels[0] {
			moduleCallBodies = append(moduleCallBodies, block.Body)
		}
	}
	// In cases where the module disappears after expansion, such as `count = 0`,
	// the module arguments should not be evaluated and is therefore skipped.
	if len(moduleCallBodies) == 0 {
		return nil, nil
	}
	// When `count = 2` expands into multiple module calls, the only differences
	// between them are the values ​​of `count.index` and `each.*`.
	// Since these are not values ​​available when the module is loaded, any module call can be used.
	moduleCallBody := moduleCallBodies[0]

	inputs := InputValues{}
	for varName, attribute := range moduleCallBody.Attributes {
		val, evalDiags := parentCtx.EvaluateExpr(attribute.Expr, cty.DynamicPseudoType)
		if evalDiags.HasErrors() {
			diags = append(diags, evalDiags...)
			continue
		}
		inputs[varName] = &InputValue{Value: val}
	}
	if diags.HasErrors() {
		return nil, diags
	}

	variableValues, diags := VariableValues(child, inputs)
	if diags.HasErrors() {
		return nil, diags
	}

	return &Evaluator{
		Meta: &ContextMeta{
			Env:                Workspace(),
			OriginalWorkingDir: parentCtx.Meta.OriginalWorkingDir,
		},
		ModulePath:     child.Path.UnkeyedInstanceShim(),
		Config:         parent.Root,
		VariableValues: variableValues,
	}, nil
}

// DescendentForInstance returns the descendent config that has the given instance path
// beneath the receiver, or nil if there is no such module.
func (c *Config) DescendentForInstance(path addrs.ModuleInstance) *Config {
	current := c
	for _, step := range path {
		current = current.Children[step.Name]
		if current == nil {
			return nil
		}
	}
	return current
}

// A ModuleWalker knows how to find and load a child module given details about
// the module to be loaded and a reference to its partially-loaded parent
// Config.
type ModuleWalker interface {
	// LoadModule finds and loads a requested child module.
	//
	// If errors are detected during loading, implementations should return them
	// in the diagnostics object. If the diagnostics object contains any errors
	// then the caller will tolerate the returned module being nil or incomplete.
	// If no errors are returned, it should be non-nil and complete.
	//
	// Full validation need not have been performed but an implementation should
	// ensure that the basic file- and module-validations performed by the
	// LoadConfigDir function (valid syntax, no namespace collisions, etc) have
	// been performed before returning a module.
	LoadModule(req *ModuleRequest) (*Module, *version.Version, hcl.Diagnostics)
}

// ModuleWalkerFunc is an implementation of ModuleWalker that directly wraps
// a callback function, for more convenient use of that interface.
type ModuleWalkerFunc func(req *ModuleRequest) (*Module, *version.Version, hcl.Diagnostics)

// LoadModule implements ModuleWalker.
func (f ModuleWalkerFunc) LoadModule(req *ModuleRequest) (*Module, *version.Version, hcl.Diagnostics) {
	return f(req)
}

// ModuleRequest is used with the ModuleWalker interface to describe a child
// module that must be loaded.
type ModuleRequest struct {
	// Name is the "logical name" of the module call within configuration.
	// This is provided in case the name is used as part of a storage key
	// for the module, but implementations must otherwise treat it as an
	// opaque string.
	Name string

	// Path is a list of logical names that traverse from the root module to
	// this module. This can be used, for example, to form a lookup key for
	// each distinct module call in a configuration, allowing for multiple
	// calls with the same name at different points in the tree.
	Path addrs.Module

	// SourceAddr is the source address string provided by the user in
	// configuration.
	SourceAddr addrs.ModuleSource

	// Parent is the partially-constructed module tree node that the loaded
	// module will be added to. Callers may refer to any field of this
	// structure except Children, which is still under construction when
	// ModuleRequest objects are created and thus has undefined content.
	// The main reason this is provided is to build the full path for the module.
	Parent *Config

	// CallRange is the source range for the header of the "module" block
	// in configuration that prompted this request. This can be used as the
	// subject of an error diagnostic that relates to the module call itself.
	CallRange hcl.Range
}
