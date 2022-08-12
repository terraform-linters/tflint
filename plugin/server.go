package plugin

import (
	"errors"
	"fmt"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/tflint"
	"github.com/zclconf/go-cty/cty"
)

// GRPCServer is a gRPC server for responding to requests from plugins.
type GRPCServer struct {
	runner     *tflint.Runner
	rootRunner *tflint.Runner
	files      map[string]*hcl.File
}

// NewGRPCServer initializes a gRPC server for plugins.
func NewGRPCServer(runner *tflint.Runner, rootRunner *tflint.Runner, files map[string]*hcl.File) *GRPCServer {
	return &GRPCServer{runner: runner, rootRunner: rootRunner, files: files}
}

// GetModulePath returns the current module path.
func (s *GRPCServer) GetModulePath() []string {
	return s.runner.TFConfig.Path
}

// GetModuleContent returns module content based on the passed schema and options.
func (s *GRPCServer) GetModuleContent(bodyS *hclext.BodySchema, opts sdk.GetModuleContentOption) (*hclext.BodyContent, hcl.Diagnostics) {
	switch opts.ModuleCtx {
	case sdk.SelfModuleCtxType:
		return s.runner.GetModuleContent(bodyS, opts)
	case sdk.RootModuleCtxType:
		return s.rootRunner.GetModuleContent(bodyS, opts)
	default:
		panic(fmt.Sprintf("unknown module ctx: %s", opts.ModuleCtx))
	}
}

// GetFile returns the hcl.File based on passed the file name.
func (s *GRPCServer) GetFile(name string) (*hcl.File, error) {
	return s.files[name], nil
}

// GetFiles returns all hcl.File in the module.
func (s *GRPCServer) GetFiles(ty sdk.ModuleCtxType) map[string][]byte {
	switch ty {
	case sdk.SelfModuleCtxType:
		return s.runner.Sources()
	case sdk.RootModuleCtxType:
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
		return nil, s.runner.ConfigSources(), fmt.Errorf("rule `%s` is not found in config", name)
	}
	// If you enable the rule through the CLI instead of the file, its hcl.Body will be nil.
	if config.Body == nil {
		return nil, s.runner.ConfigSources(), errors.New("This rule cannot be enabled with the `--enable-rule` option because it lacks the required configuration")
	}

	body, diags := hclext.Content(config.Body, bodyS)
	if diags.HasErrors() {
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
		runner = s.rootRunner
	}
	val, err := runner.EvaluateExpr(expr, *opts.WantType)
	return val, err
}

// EmitIssue stores an issue in the server based on passed rule, message, and location.
// If the range associated with the issue is an expression, it propagates to the runner
// that the issue found in that expression. This allows you to determine if the issue was caused
// by a module argument in the case of module inspection.
func (s *GRPCServer) EmitIssue(rule sdk.Rule, message string, location hcl.Range) error {
	file := s.runner.File(location.Filename)
	if file == nil {
		s.runner.EmitIssue(rule, message, location)
		return nil
	}
	expr, diags := hclext.ParseExpression(location.SliceBytes(file.Bytes), location.Filename, location.Start)
	if diags.HasErrors() {
		s.runner.EmitIssue(rule, message, location)
		return nil
	}
	return s.runner.WithExpressionContext(expr, func() error {
		s.runner.EmitIssue(rule, message, location)
		return nil
	})
}
