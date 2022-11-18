package plugin

import (
	"errors"
	"fmt"
	"log"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/terraform"
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
	var module *terraform.Module
	var ctx *terraform.Evaluator

	switch opts.ModuleCtx {
	case sdk.SelfModuleCtxType:
		module = s.runner.TFConfig.Module
		ctx = s.runner.Ctx
	case sdk.RootModuleCtxType:
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

	//nolint:staticcheck
	if opts.IncludeNotCreated || opts.ExpandMode == sdk.ExpandModeNone {
		ctx = nil
	}

	return module.PartialContent(bodyS, ctx)
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
			return nil, s.runner.ConfigSources(), errors.New("This rule cannot be enabled with the `--enable-rule` option because it lacks the required configuration")
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
		runner = s.rootRunner
	}

	val, diags := runner.Ctx.EvaluateExpr(expr, *opts.WantType)
	if diags.HasErrors() {
		return val, diags
	}

	if val.ContainsMarked() {
		err := fmt.Errorf(
			"sensitive value found in %s:%d%w",
			expr.Range().Filename,
			expr.Range().Start.Line,
			sdk.ErrSensitive,
		)
		log.Printf("[INFO] %s. TFLint ignores expressions with sensitive values.", err)
		return cty.NullVal(cty.NilType), err
	}

	if *opts.WantType == cty.DynamicPseudoType {
		return val, nil
	}

	err := cty.Walk(val, func(path cty.Path, v cty.Value) (bool, error) {
		if !v.IsKnown() {
			err := fmt.Errorf(
				"unknown value found in %s:%d%w",
				expr.Range().Filename,
				expr.Range().Start.Line,
				sdk.ErrUnknownValue,
			)
			log.Printf("[INFO] %s. TFLint can only evaluate provided variables and skips dynamic values.", err)
			return false, err
		}

		if v.IsNull() {
			err := fmt.Errorf(
				"null value found in %s:%d%w",
				expr.Range().Filename,
				expr.Range().Start.Line,
				sdk.ErrNullValue,
			)
			log.Printf("[INFO] %s. TFLint ignores expressions with null values.", err)
			return false, err
		}

		return true, nil
	})
	if err != nil {
		return cty.NullVal(cty.NilType), err
	}

	return val, nil
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
