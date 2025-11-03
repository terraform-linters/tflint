package plugin

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/hashicorp/go-version"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/plugin2host"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang/marks"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/terraform"
	"github.com/terraform-linters/tflint/tflint"
	"github.com/zclconf/go-cty/cty"
)

// GRPCServer is a gRPC server for responding to requests from plugins.
type GRPCServer struct {
	mu               sync.Mutex
	runner           *tflint.Runner
	rootRunner       *tflint.Runner
	files            map[string]*hcl.File
	clientSDKVersion *version.Version
}

var _ plugin2host.Server = (*GRPCServer)(nil)

// NewGRPCServer initializes a gRPC server for plugins.
func NewGRPCServer(runner *tflint.Runner, rootRunner *tflint.Runner, files map[string]*hcl.File, sdkVersion *version.Version) *GRPCServer {
	return &GRPCServer{runner: runner, rootRunner: rootRunner, files: files, clientSDKVersion: sdkVersion}
}

// GetOriginalwd returns the original working directory.
func (s *GRPCServer) GetOriginalwd() string {
	return s.runner.Ctx.Meta.OriginalWorkingDir
}

// GetModulePath returns the current module path.
func (s *GRPCServer) GetModulePath() []string {
	return s.runner.TFConfig.Path
}

// GetModuleContent returns module content based on the passed schema and options.
func (s *GRPCServer) GetModuleContent(bodyS *hclext.BodySchema, opts sdk.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
	var module *terraform.Module
	var ctx *terraform.Evaluator

	switch opts.ModuleCtx {
	case sdk.SelfModuleCtxType:
		module = s.runner.TFConfig.Module
		ctx = s.runner.Ctx
	case sdk.RootModuleCtxType:
		s.mu.Lock()
		defer s.mu.Unlock()
		module = s.rootRunner.TFConfig.Module
		ctx = s.rootRunner.Ctx
	default:
		panic(fmt.Sprintf("unknown module ctx: %s", opts.ModuleCtx))
	}

	// For performance, determine in advance whether the target resource exists.
	if opts.Hint.ResourceType != "" {
		if _, exists := module.Resources[opts.Hint.ResourceType]; !exists {
			return &hclext.BodyContent{}, nil
		}
	}

	if opts.ExpandMode == sdk.ExpandModeNone {
		ctx = nil
	}

	return module.PartialContent(bodyS, ctx)
}

// GetFile returns the hcl.File based on passed the file name.
func (s *GRPCServer) GetFile(name string) (*hcl.File, error) {
	// Considering that autofix has been applied, prioritize returning the value of runner.Files().
	if file, exists := s.runner.Files()[name]; exists {
		return file, nil
	}
	// If the file is not found in the current module, it may be in other modules (e.g. root module).
	log.Printf(`[DEBUG] The file "%s" is not found in the current module. Fall back to global caches.`, name)
	return s.files[name], nil
}

// GetFiles returns all hcl.File in the module.
func (s *GRPCServer) GetFiles(ty sdk.ModuleCtxType) map[string][]byte {
	switch ty {
	case sdk.SelfModuleCtxType:
		return s.runner.Sources()
	case sdk.RootModuleCtxType:
		// HINT: This is an operation on the root runner,
		//       but it works without locking since it is obviously readonly.
		return s.rootRunner.Sources()
	default:
		panic(fmt.Sprintf("invalid ModuleCtxType: %s", ty))
	}
}

// GetRuleConfigContent extracts the rule config based on the schema.
// It returns an extracted body content and sources.
// The reason for returning sources is to encode the expression, and there is room for improvement here.
func (s *GRPCServer) GetRuleConfigContent(name string, bodyS *hclext.BodySchema) (*hclext.BodyContent, map[string][]byte, error) {
	config := s.runner.RuleConfig(name)
	if config == nil {
		return &hclext.BodyContent{}, s.runner.ConfigSources(), nil
	}

	enabledByCLI := false
	configBody := config.Body
	// If you enable the rule through the CLI instead of the file, its hcl.Body will be nil.
	if config.Body == nil {
		enabledByCLI = true
		configBody = hcl.EmptyBody()
	}

	body, diags := hclext.Content(configBody, bodyS)
	if diags.HasErrors() {
		if enabledByCLI {
			return nil, s.runner.ConfigSources(), errors.New("This rule cannot be enabled with the --enable-rule option because it lacks the required configuration")
		}
		return body, s.runner.ConfigSources(), diags
	}
	return body, s.runner.ConfigSources(), nil
}

// EvaluateExpr returns the value of the passed expression.
func (s *GRPCServer) EvaluateExpr(expr hcl.Expression, opts sdk.EvaluateExprOption) (cty.Value, error) {
	var runner *tflint.Runner
	switch opts.ModuleCtx {
	case sdk.SelfModuleCtxType:
		runner = s.runner
	case sdk.RootModuleCtxType:
		s.mu.Lock()
		defer s.mu.Unlock()
		runner = s.rootRunner
	}

	val, diags := runner.Ctx.EvaluateExpr(expr, *opts.WantType)
	if diags.HasErrors() {
		return val, diags
	}

	// If an ephemeral mark is contained, cty.Value will not be returned
	// unless the plugin is built with SDK 0.22+ which supports ephemeral marks.
	if !marks.Contains(val, marks.Ephemeral) || SupportsEphemeralMarks(s.clientSDKVersion) {
		return val, nil
	}

	// Plugins that do not support ephemeral marks will return ErrSensitive to prevent secrets from being exposed.
	// Do not return ErrEphemeral as it is not supported by plugins.
	err := fmt.Errorf(
		"ephemeral value found in %s:%d%w",
		expr.Range().Filename,
		expr.Range().Start.Line,
		sdk.ErrSensitive,
	)
	log.Printf("[INFO] %s. TFLint ignores ephemeral values for plugins built with SDK versions earlier than v0.22.", err)
	return cty.NullVal(cty.NilType), err
}

// EmitIssue stores an issue in the server based on passed rule, message, and location.
// It attempts to detect whether the issue range represents an expression and emits it based on that context.
// However, some ranges may be syntactically valid but not actually represent an expression.
// In these cases, the "expression" is still provided as context and the client should ignore any errors when attempting to evaluate it.
func (s *GRPCServer) EmitIssue(rule sdk.Rule, message string, location hcl.Range, fixable bool) (bool, error) {
	// If the issue range represents an expression, it is emitted based on that context.
	// This is required to emit issues in called modules.
	expr, err := s.getExprFromRange(location)
	if err != nil {
		// If the range does not represent an expression, just emit it without context.
		return s.runner.EmitIssue(rule, message, location, fixable), nil
	}

	var applied bool
	err = s.runner.WithExpressionContext(expr, func() error {
		applied = s.runner.EmitIssue(rule, message, location, fixable)
		return nil
	})
	return applied, err
}

func (s *GRPCServer) getExprFromRange(location hcl.Range) (hcl.Expression, error) {
	file := s.runner.File(location.Filename)
	if file == nil {
		return nil, errors.New("file not found")
	}
	expr, diags := hclext.ParseExpression(location.SliceBytes(file.Bytes), location.Filename, location.Start)
	if diags.HasErrors() {
		return nil, diags
	}
	return expr, nil
}

// ApplyChanges applies the autofix changes to the runner.
func (s *GRPCServer) ApplyChanges(changes map[string][]byte) error {
	diags := s.runner.ApplyChanges(changes)
	if diags.HasErrors() {
		return diags
	}
	return nil
}
