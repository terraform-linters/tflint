// Adapted from tflint-plugin-sdk v0.25.0/plugin/internal/plugin2host/client.go.
// Licensed under MPL-2.0.
package plugin

import (
	stdjson "encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/logger"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/plugin2host"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang/marks"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/internal/sdkfixer"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type localRunner struct {
	server     plugin2host.Server
	Fixer      *sdkfixer.Fixer
	FixEnabled bool
}

var _ tflint.Runner = (*localRunner)(nil)

func (c *localRunner) GetOriginalwd() (string, error)       { return c.server.GetOriginalwd(), nil }
func (c *localRunner) GetModulePath() (addrs.Module, error) { return c.server.GetModulePath(), nil }
func (c *localRunner) GetModuleContent(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	if opts == nil {
		opts = &tflint.GetModuleContentOption{}
	}
	content, diags := c.server.GetModuleContent(schema, *opts)
	if diags.HasErrors() {
		return content, diags
	}
	return content, nil
}
func (c *localRunner) GetFile(name string) (*hcl.File, error) { return c.server.GetFile(name) }
func (c *localRunner) GetFiles() (map[string]*hcl.File, error) {
	files := map[string]*hcl.File{}
	for name := range c.server.GetFiles(tflint.SelfModuleCtxType) {
		file, err := c.server.GetFile(name)
		if err != nil {
			return nil, err
		}
		files[name] = file
	}
	return files, nil
}
func (c *localRunner) DecodeRuleConfig(name string, target interface{}) error {
	content, _, err := c.server.GetRuleConfigContent(name, hclext.ImpliedBodySchema(target))
	if err != nil {
		return err
	}
	if content.IsEmpty() {
		return nil
	}
	diags := hclext.DecodeBody(content, nil, target)
	if diags.HasErrors() {
		return diags
	}
	return nil
}
func (c *localRunner) EmitIssue(rule tflint.Rule, message string, location hcl.Range) error {
	_, err := c.server.EmitIssue(rule, message, location, false)
	return err
}
func (c *localRunner) GetResourceContent(name string, inner *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	if opts == nil {
		opts = &tflint.GetModuleContentOption{}
	}
	opts.Hint.ResourceType = name

	body, err := c.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{Type: "resource", LabelNames: []string{"type", "name"}, Body: inner},
		},
	}, opts)
	if err != nil {
		return nil, err
	}

	content := &hclext.BodyContent{Blocks: []*hclext.Block{}}
	for _, resource := range body.Blocks {
		if resource.Labels[0] != name {
			continue
		}

		content.Blocks = append(content.Blocks, resource)
	}

	return content, nil
}

// GetProviderContent gets the contents of providers based on the schema.
// This is shorthand of GetModuleContent for providers
func (c *localRunner) GetProviderContent(name string, inner *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	if opts == nil {
		opts = &tflint.GetModuleContentOption{}
	}

	body, err := c.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{Type: "provider", LabelNames: []string{"name"}, Body: inner},
		},
	}, opts)
	if err != nil {
		return nil, err
	}

	content := &hclext.BodyContent{Blocks: []*hclext.Block{}}
	for _, provider := range body.Blocks {
		if provider.Labels[0] != name {
			continue
		}

		content.Blocks = append(content.Blocks, provider)
	}

	return content, nil
}

type nativeWalker struct {
	walker tflint.ExprWalker
}

func (w *nativeWalker) Enter(node hclsyntax.Node) hcl.Diagnostics {
	if expr, ok := node.(hcl.Expression); ok {
		return w.walker.Enter(expr)
	}
	return nil
}

func (w *nativeWalker) Exit(node hclsyntax.Node) hcl.Diagnostics {
	if expr, ok := node.(hcl.Expression); ok {
		return w.walker.Exit(expr)
	}
	return nil
}

// extractJSONKeys extracts attribute names from JSON bytes using encoding/json.
// This works for both object-based JSON {"foo": ...} and array-based JSON [{"foo": ...}].
func extractJSONKeys(bytes []byte) ([]string, error) {
	// Try to unmarshal as an object first
	var obj map[string]any
	if err := stdjson.Unmarshal(bytes, &obj); err == nil {
		keys := make([]string, 0, len(obj))
		for k := range obj {
			keys = append(keys, k)
		}
		return keys, nil
	}

	// Try as an array of objects
	var arr []map[string]any
	if err := stdjson.Unmarshal(bytes, &arr); err != nil {
		return nil, err
	}

	// Collect all unique keys from all objects in the array
	keysMap := make(map[string]bool)
	for _, obj := range arr {
		for k := range obj {
			keysMap[k] = true
		}
	}

	keys := make([]string, 0, len(keysMap))
	for k := range keysMap {
		keys = append(keys, k)
	}
	return keys, nil
}

// getJSONAttributes gets all attributes from a JSON body, supporting both object
// and array-based syntax. For array-based JSON like [{"import": {...}}], it
// extracts attribute names using encoding/json and builds a schema to extract them.
func getJSONAttributes(body hcl.Body, bytes []byte) (hcl.Attributes, hcl.Diagnostics) {
	// First, try JustAttributes (works for object-based JSON)
	attrs, diags := body.JustAttributes()
	if !diags.HasErrors() {
		return attrs, nil
	}

	// Extract keys using encoding/json
	keys, err := extractJSONKeys(bytes)
	if err != nil {
		return attrs, diags // Return original JustAttributes error
	}

	// Build a schema with all discovered keys
	schema := &hcl.BodySchema{
		Attributes: make([]hcl.AttributeSchema, len(keys)),
	}
	for i, key := range keys {
		schema.Attributes[i] = hcl.AttributeSchema{Name: key}
	}

	// Use PartialContent to get proper *json.expression objects
	content, _, partialDiags := body.PartialContent(schema)
	return content.Attributes, partialDiags
}

// WalkExpressions traverses expressions in all files by the passed walker.
// Note that it behaves differently in native HCL syntax and JSON syntax.
//
// In the HCL syntax, `var.foo` and `var.bar` in `[var.foo, var.bar]` are
// also passed to the walker. In other words, it traverses expressions recursively.
// To avoid redundant checks, the walker should check the kind of expression.
//
// In the JSON syntax, only an expression of an attribute seen from the top
// level of the file is passed. In other words, it doesn't traverse expressions
// recursively. This is a limitation of JSON syntax.
func (c *localRunner) WalkExpressions(walker tflint.ExprWalker) hcl.Diagnostics {
	files, err := c.GetFiles()
	if err != nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "failed to call GetFiles()",
				Detail:   err.Error(),
			},
		}
	}

	diags := hcl.Diagnostics{}
	for _, file := range files {
		if body, ok := file.Body.(*hclsyntax.Body); ok {
			walkDiags := hclsyntax.Walk(body, &nativeWalker{walker: walker})
			diags = diags.Extend(walkDiags)
			continue
		}

		// In JSON syntax, everything can be walked as an attribute.
		attrs, jsonDiags := getJSONAttributes(file.Body, file.Bytes)
		if jsonDiags.HasErrors() {
			diags = diags.Extend(jsonDiags)
			continue
		}

		for _, attr := range attrs {
			enterDiags := walker.Enter(attr.Expr)
			diags = diags.Extend(enterDiags)
			exitDiags := walker.Exit(attr.Expr)
			diags = diags.Extend(exitDiags)
		}
	}

	return diags
}

var errUnevaluable = tflint.ErrUnevaluable //nolint:staticcheck // Preserve the SDK Runner contract for legacy callers.

var errRefTy = reflect.TypeOf((*error)(nil)).Elem()

// EvaluateExpr evals the passed expression based on the type.
// Passing a callback function instead of a value as the target will invoke the callback,
// passing the evaluated value to the argument.
func (c *localRunner) EvaluateExpr(expr hcl.Expression, target interface{}, opts *tflint.EvaluateExprOption) error {
	rval := reflect.ValueOf(target)
	rty := rval.Type()

	var callback bool
	switch rty.Kind() {
	case reflect.Func:
		// Callback must meet the following requirements:
		//   - It must be a function
		//   - It must take an argument
		//   - It must return an error
		if rty.NumIn() != 1 || rty.NumOut() != 1 || !rty.Out(0).Implements(errRefTy) {
			panic(`callback must be of type "func (v T) error"`)
		}
		callback = true
		target = reflect.New(rty.In(0)).Interface()

	case reflect.Pointer:
		// ok
	default:
		panic("target value is not a pointer or function")
	}

	err := c.evaluateExpr(expr, target, opts)
	if !callback {
		// error should be handled in the caller
		return err
	}

	if err != nil {
		// If it cannot be represented as a Go value, exit without invoking the callback rather than returning an error.
		if errors.Is(err, tflint.ErrUnknownValue) ||
			errors.Is(err, tflint.ErrNullValue) ||
			errors.Is(err, tflint.ErrSensitive) ||
			errors.Is(err, tflint.ErrEphemeral) ||
			errors.Is(err, errUnevaluable) {
			return nil
		}
		return err
	}

	rerr := rval.Call([]reflect.Value{reflect.ValueOf(target).Elem()})
	if rerr[0].IsNil() {
		return nil
	}
	return rerr[0].Interface().(error)
}

func (c *localRunner) evaluateExpr(expr hcl.Expression, target interface{}, opts *tflint.EvaluateExprOption) error {
	if opts == nil {
		opts = &tflint.EvaluateExprOption{}
	}

	var ty cty.Type
	if opts.WantType != nil {
		ty = *opts.WantType
	} else {
		switch target.(type) {
		case *string:
			ty = cty.String
		case *int:
			ty = cty.Number
		case *bool:
			ty = cty.Bool
		case *[]string:
			ty = cty.List(cty.String)
		case *[]int:
			ty = cty.List(cty.Number)
		case *[]bool:
			ty = cty.List(cty.Bool)
		case *map[string]string:
			ty = cty.Map(cty.String)
		case *map[string]int:
			ty = cty.Map(cty.Number)
		case *map[string]bool:
			ty = cty.Map(cty.Bool)
		case *cty.Value:
			ty = cty.DynamicPseudoType
		default:
			panic(fmt.Sprintf("unsupported target type: %T", target))
		}
	}
	option := *opts
	option.WantType = &ty
	val, err := c.server.EvaluateExpr(expr, option)
	if err != nil {
		return err
	}

	if ty == cty.DynamicPseudoType {
		return gocty.FromCtyValue(val, target)
	}

	// Returns an error if the value cannot be decoded to a Go value (e.g. unknown, null, marked).
	// This allows the caller to handle the value by the errors package.
	err = cty.Walk(val, func(path cty.Path, v cty.Value) (bool, error) {
		if !v.IsKnown() {
			logger.Debug(fmt.Sprintf("unknown value found in %s", expr.Range()))
			return false, tflint.ErrUnknownValue
		}
		if v.IsNull() {
			logger.Debug(fmt.Sprintf("null value found in %s", expr.Range()))
			return false, tflint.ErrNullValue
		}
		if v.HasMark(marks.Sensitive) {
			logger.Debug(fmt.Sprintf("sensitive value found in %s", expr.Range()))
			return false, tflint.ErrSensitive
		}
		if v.HasMark(marks.Ephemeral) {
			logger.Debug(fmt.Sprintf("ephemeral value found in %s", expr.Range()))
			return false, tflint.ErrEphemeral
		}
		return true, nil
	})
	if err != nil {
		return err
	}

	return gocty.FromCtyValue(val, target)
}

func (c *localRunner) EmitIssueWithFix(rule tflint.Rule, message string, location hcl.Range, fixFunc func(f tflint.Fixer) error) error {
	var fixable bool
	var fixErr error

	path, err := c.GetModulePath()
	if err != nil {
		return err
	}
	// If the issue is not in the root module, skip the fix.
	if path.IsRoot() {
		fixable = true
		c.Fixer.StashChanges()

		fixErr = fixFunc(c.Fixer)
		if errors.Is(fixErr, tflint.ErrFixNotSupported) {
			fixable = false
		}
	}

	applied, err := c.server.EmitIssue(rule, message, location, fixable)
	if err != nil {
		return err
	}

	if !c.FixEnabled || !fixable || !applied {
		c.Fixer.PopChangesFromStash()
		return nil
	}
	return fixErr
}

func (*localRunner) EnsureNoError(err error, proc func() error) error {
	if err == nil {
		return proc()
	}

	if errors.Is(err, errUnevaluable) || errors.Is(err, tflint.ErrNullValue) || errors.Is(err, tflint.ErrUnknownValue) || errors.Is(err, tflint.ErrSensitive) {
		return nil
	}
	return err
}
