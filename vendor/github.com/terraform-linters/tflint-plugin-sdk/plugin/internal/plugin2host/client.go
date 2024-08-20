package plugin2host

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	hcljson "github.com/hashicorp/hcl/v2/json"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/internal"
	"github.com/terraform-linters/tflint-plugin-sdk/logger"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/internal/fromproto"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/internal/proto"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/internal/toproto"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"github.com/zclconf/go-cty/cty/json"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCClient is a plugin-side implementation. Plugin can send requests through the client to host's gRPC server.
type GRPCClient struct {
	Client     proto.RunnerClient
	Fixer      *internal.Fixer
	FixEnabled bool
}

var _ tflint.Runner = &GRPCClient{}

// GetOriginalwd gets the original working directory.
func (c *GRPCClient) GetOriginalwd() (string, error) {
	resp, err := c.Client.GetOriginalwd(context.Background(), &proto.GetOriginalwd_Request{})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.Unimplemented {
			// Originalwd is available in TFLint v0.44+
			// Fallback to os.Getwd() because it equals the current directory in earlier versions.
			return os.Getwd()
		}
		return "", fromproto.Error(err)
	}
	return resp.Path, err
}

// GetModulePath gets the current module path address.
func (c *GRPCClient) GetModulePath() (addrs.Module, error) {
	resp, err := c.Client.GetModulePath(context.Background(), &proto.GetModulePath_Request{})
	if err != nil {
		return nil, fromproto.Error(err)
	}
	return resp.Path, err
}

// GetResourceContent gets the contents of resources based on the schema.
// This is shorthand of GetModuleContent for resources
func (c *GRPCClient) GetResourceContent(name string, inner *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
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
func (c *GRPCClient) GetProviderContent(name string, inner *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
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

// GetModuleContent gets the contents of the module based on the schema.
func (c *GRPCClient) GetModuleContent(schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	if opts == nil {
		opts = &tflint.GetModuleContentOption{}
	}

	req := &proto.GetModuleContent_Request{
		Schema: toproto.BodySchema(schema),
		Option: toproto.GetModuleContentOption(opts),
	}
	resp, err := c.Client.GetModuleContent(context.Background(), req)
	if err != nil {
		return nil, fromproto.Error(err)
	}

	body, diags := fromproto.BodyContent(resp.Content)
	if diags.HasErrors() {
		err = diags
	}
	return body, err
}

// GetFile returns hcl.File based on the passed file name.
func (c *GRPCClient) GetFile(file string) (*hcl.File, error) {
	resp, err := c.Client.GetFile(context.Background(), &proto.GetFile_Request{Name: file})
	if err != nil {
		return nil, fromproto.Error(err)
	}

	var f *hcl.File
	var diags hcl.Diagnostics
	if strings.HasSuffix(file, ".tf") {
		f, diags = hclsyntax.ParseConfig(resp.File, file, hcl.InitialPos)
	} else {
		f, diags = hcljson.Parse(resp.File, file)
	}

	if diags.HasErrors() {
		err = diags
	}
	return f, err
}

// GetFiles returns bytes of hcl.File in the self module context.
func (c *GRPCClient) GetFiles() (map[string]*hcl.File, error) {
	resp, err := c.Client.GetFiles(context.Background(), &proto.GetFiles_Request{})
	if err != nil {
		return nil, fromproto.Error(err)
	}

	files := map[string]*hcl.File{}
	var f *hcl.File
	var diags hcl.Diagnostics
	for name, bytes := range resp.Files {
		var d hcl.Diagnostics
		if strings.HasSuffix(name, ".tf") {
			f, d = hclsyntax.ParseConfig(bytes, name, hcl.InitialPos)
		} else {
			f, d = hcljson.Parse(bytes, name)
		}
		diags = diags.Extend(d)

		files[name] = f
	}

	if diags.HasErrors() {
		return files, diags
	}
	return files, nil
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
func (c *GRPCClient) WalkExpressions(walker tflint.ExprWalker) hcl.Diagnostics {
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
		attrs, jsonDiags := file.Body.JustAttributes()
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

// DecodeRuleConfig guesses the schema of the rule config from the passed interface and sends the schema to GRPC server.
// Content retrieved based on the schema is decoded into the passed interface.
func (c *GRPCClient) DecodeRuleConfig(name string, ret interface{}) error {
	resp, err := c.Client.GetRuleConfigContent(context.Background(), &proto.GetRuleConfigContent_Request{
		Name:   name,
		Schema: toproto.BodySchema(hclext.ImpliedBodySchema(ret)),
	})
	if err != nil {
		return fromproto.Error(err)
	}

	content, diags := fromproto.BodyContent(resp.Content)
	if diags.HasErrors() {
		return diags
	}
	if content.IsEmpty() {
		return nil
	}

	diags = hclext.DecodeBody(content, nil, ret)
	if diags.HasErrors() {
		return diags
	}
	return nil
}

var errRefTy = reflect.TypeOf((*error)(nil)).Elem()

// EvaluateExpr evals the passed expression based on the type.
// Passing a callback function instead of a value as the target will invoke the callback,
// passing the evaluated value to the argument.
func (c *GRPCClient) EvaluateExpr(expr hcl.Expression, target interface{}, opts *tflint.EvaluateExprOption) error {
	rval := reflect.ValueOf(target)
	rty := rval.Type()

	var callback bool
	switch rty.Kind() {
	case reflect.Func:
		// Callback must meet the following requirements:
		//   - It must be a function
		//   - It must take an argument
		//   - It must return an error
		if !(rty.NumIn() == 1 && rty.NumOut() == 1 && rty.Out(0).Implements(errRefTy)) {
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
		if errors.Is(err, tflint.ErrUnknownValue) || errors.Is(err, tflint.ErrNullValue) || errors.Is(err, tflint.ErrSensitive) || errors.Is(err, tflint.ErrUnevaluable) {
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

func (c *GRPCClient) evaluateExpr(expr hcl.Expression, target interface{}, opts *tflint.EvaluateExprOption) error {
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
	tyby, err := json.MarshalType(ty)
	if err != nil {
		return err
	}

	file, err := c.GetFile(expr.Range().Filename)
	if err != nil {
		return err
	}

	resp, err := c.Client.EvaluateExpr(
		context.Background(),
		&proto.EvaluateExpr_Request{
			Expression: toproto.Expression(expr, file.Bytes),
			Option:     &proto.EvaluateExpr_Option{Type: tyby, ModuleCtx: toproto.ModuleCtxType(opts.ModuleCtx)},
		},
	)
	if err != nil {
		return fromproto.Error(err)
	}

	val, err := fromproto.Value(resp.Value, ty, resp.Marks)
	if err != nil {
		return err
	}

	if ty == cty.DynamicPseudoType {
		return gocty.FromCtyValue(val, target)
	}

	// Returns an error if the value cannot be decoded to a Go value (e.g. unknown, null, sensitive).
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
		if v.IsMarked() {
			logger.Debug(fmt.Sprintf("sensitive value found in %s", expr.Range()))
			return false, tflint.ErrSensitive
		}
		return true, nil
	})
	if err != nil {
		return err
	}

	return gocty.FromCtyValue(val, target)
}

// EmitIssue emits the issue with the passed rule, message, location
func (c *GRPCClient) EmitIssue(rule tflint.Rule, message string, location hcl.Range) error {
	_, err := c.Client.EmitIssue(context.Background(), &proto.EmitIssue_Request{Rule: toproto.Rule(rule), Message: message, Range: toproto.Range(location)})
	if err != nil {
		return fromproto.Error(err)
	}
	return nil
}

// EmitIssueWithFix emits the issue with the passed rule, message, location.
// Invoke the fix function and add the changes to the fixer.
// If the fix function returns ErrFixNotSupported, the emitted issue will not
// be marked as fixable.
func (c *GRPCClient) EmitIssueWithFix(rule tflint.Rule, message string, location hcl.Range, fixFunc func(f tflint.Fixer) error) error {
	var fixable bool
	var fixErr error

	path, err := c.GetModulePath()
	if err != nil {
		return fromproto.Error(err)
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

	resp, err := c.Client.EmitIssue(context.Background(), &proto.EmitIssue_Request{Rule: toproto.Rule(rule), Message: message, Range: toproto.Range(location), Fixable: fixable})
	if err != nil {
		return fromproto.Error(err)
	}

	if !c.FixEnabled || !fixable || !resp.Applied {
		c.Fixer.PopChangesFromStash()
		return nil
	}
	return fixErr
}

// ApplyChanges applies the changes in the fixer to the server
func (c *GRPCClient) ApplyChanges() error {
	_, err := c.Client.ApplyChanges(context.Background(), &proto.ApplyChanges_Request{Changes: c.Fixer.Changes()})
	if err != nil {
		return fromproto.Error(err)
	}
	c.Fixer.ApplyChanges()
	return nil
}

// EnsureNoError is a helper for error handling. Depending on the type of error generated by EvaluateExpr,
// determine whether to exit, skip, or continue. If it is continued, the passed function will be executed.
//
// Deprecated: Use errors.Is() instead to determine which errors can be ignored.
func (*GRPCClient) EnsureNoError(err error, proc func() error) error {
	if err == nil {
		return proc()
	}

	if errors.Is(err, tflint.ErrUnevaluable) || errors.Is(err, tflint.ErrNullValue) || errors.Is(err, tflint.ErrUnknownValue) || errors.Is(err, tflint.ErrSensitive) {
		return nil
	}
	return err
}
