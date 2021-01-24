package plugin

import (
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform/configs"
	client "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	tfplugin "github.com/terraform-linters/tflint-plugin-sdk/tflint/client"
	"github.com/terraform-linters/tflint/tflint"
	"github.com/zclconf/go-cty/cty"
)

// Server is a RPC server for responding to requests from plugins
type Server struct {
	runner     *tflint.Runner
	rootRunner *tflint.Runner
	sources    map[string][]byte
}

// NewServer initializes a RPC server for plugins
func NewServer(runner *tflint.Runner, rootRunner *tflint.Runner, sources map[string][]byte) *Server {
	return &Server{runner: runner, rootRunner: rootRunner, sources: sources}
}

// Attributes returns corresponding hcl.Attributes
func (s *Server) Attributes(req *tfplugin.AttributesRequest, resp *tfplugin.AttributesResponse) error {
	ret := []*tfplugin.Attribute{}
	err := s.runner.WalkResourceAttributes(req.Resource, req.AttributeName, func(attr *hcl.Attribute) error {
		ret = append(ret, &tfplugin.Attribute{
			Name:      attr.Name,
			Expr:      attr.Expr.Range().SliceBytes(s.sources[attr.Expr.Range().Filename]),
			ExprRange: attr.Expr.Range(),
			Range:     attr.Range,
			NameRange: attr.NameRange,
		})
		return nil
	})
	*resp = tfplugin.AttributesResponse{Attributes: ret, Err: err}
	return nil
}

// Blocks returns corresponding hcl.Blocks
func (s *Server) Blocks(req *tfplugin.BlocksRequest, resp *tfplugin.BlocksResponse) error {
	ret := []*tfplugin.Block{}
	err := s.runner.WalkResourceBlocks(req.Resource, req.BlockType, func(block *hcl.Block) error {
		bodyRange := tflint.HCLBodyRange(block.Body, block.DefRange)
		ret = append(ret, &tfplugin.Block{
			Type:        block.Type,
			Labels:      block.Labels,
			Body:        bodyRange.SliceBytes(s.runner.File(block.DefRange.Filename).Bytes),
			BodyRange:   bodyRange,
			DefRange:    block.DefRange,
			TypeRange:   block.TypeRange,
			LabelRanges: block.LabelRanges,
		})
		return nil
	})
	*resp = tfplugin.BlocksResponse{Blocks: ret, Err: err}
	return nil
}

// Resources returns corresponding configs.Resource as tfplugin.Resource
func (s *Server) Resources(req *tfplugin.ResourcesRequest, resp *tfplugin.ResourcesResponse) error {
	var ret []*tfplugin.Resource
	err := s.runner.WalkResources(req.Name, func(resource *configs.Resource) error {
		ret = append(ret, s.encodeResource(resource))
		return nil
	})
	*resp = tfplugin.ResourcesResponse{Resources: ret, Err: err}
	return nil
}

// ModuleCalls returns all configs.ModuleCall as tfplugin.ModuleCall
func (s *Server) ModuleCalls(req *tfplugin.ModuleCallsRequest, resp *tfplugin.ModuleCallsResponse) error {
	ret := []*tfplugin.ModuleCall{}
	err := s.runner.WalkModuleCalls(func(call *configs.ModuleCall) error {
		ret = append(ret, s.encodeModuleCall(call))
		return nil
	})
	*resp = tfplugin.ModuleCallsResponse{ModuleCalls: ret, Err: err}
	return nil
}

// Backend returns corresponding configs.Backend as tfplugin.Backend
func (s *Server) Backend(req *tfplugin.BackendRequest, resp *tfplugin.BackendResponse) error {
	backend := s.runner.Backend()
	if backend == nil {
		return nil
	}

	*resp = tfplugin.BackendResponse{
		Backend: s.encodeBackend(backend),
	}

	return nil
}

// Config returns corresponding configs.Config as tfplugin.Config
func (s *Server) Config(req *tfplugin.ConfigRequest, resp *tfplugin.ConfigResponse) error {
	config, err := s.encodeConfig(s.runner.TFConfig)
	if err != nil {
		return err
	}
	*resp = tfplugin.ConfigResponse{
		Config: config,
	}

	return nil
}

// File returns the corresponding hcl.File object
func (s *Server) File(req *tfplugin.FileRequest, resp *tfplugin.FileResponse) error {
	file := s.runner.File(req.Filename)
	if file == nil {
		*resp = tfplugin.FileResponse{}
		return nil
	}

	*resp = tfplugin.FileResponse{
		Bytes: file.Bytes,
		Range: tflint.HCLBodyRange(file.Body, hcl.Range{Filename: req.Filename, Start: hcl.InitialPos}),
	}
	return nil
}

// RootProvider returns the provider configuration on the root module as tfplugin.Provider
func (s *Server) RootProvider(req *tfplugin.RootProviderRequest, resp *tfplugin.RootProviderResponse) error {
	provider, exists := s.rootRunner.TFConfig.Module.ProviderConfigs[req.Name]
	if !exists {
		return nil
	}

	*resp = tfplugin.RootProviderResponse{
		Provider: s.encodeProvider(provider),
	}
	return nil
}

// RuleConfig returns an encoded HCL format rule config
func (s *Server) RuleConfig(req *tfplugin.RuleConfigRequest, resp *tfplugin.RuleConfigResponse) error {
	config := s.runner.RuleConfig(req.Name)
	if config == nil {
		*resp = tfplugin.RuleConfigResponse{Exists: false}
		return nil
	}
	// HACK: If you enable the rule through the CLI instead of the file, its hcl.Body will not contain valid range.
	// @see https://github.com/hashicorp/hcl/blob/v2.8.0/merged.go#L132-L135
	if config.Body.MissingItemRange().Filename == "<empty>" {
		*resp = tfplugin.RuleConfigResponse{Err: client.Error{Message: "This rule cannot be enabled with the `--enable-rule` option because it lacks the required configuration"}}
		return nil
	}

	configRange := configBodyRange(config.Body)
	*resp = tfplugin.RuleConfigResponse{
		Exists: true,
		Config: configRange.SliceBytes(config.Bytes()),
		Range:  configRange,
		Err:    nil,
	}
	return nil
}

// EvalExpr returns a value of the evaluated expression
func (s *Server) EvalExpr(req *tfplugin.EvalExprRequest, resp *tfplugin.EvalExprResponse) error {
	expr, diags := tflint.ParseExpression(req.Expr, req.ExprRange.Filename, req.ExprRange.Start)
	if diags.HasErrors() {
		return diags
	}

	val, err := s.runner.EvalExpr(expr, req.Ret, cty.Type{})
	if err != nil {
		if appErr, ok := err.(*tflint.Error); ok {
			err = client.Error(*appErr)
		}
	}
	*resp = tfplugin.EvalExprResponse{Val: val, Err: err}
	return nil
}

// EvalExprOnRootCtx returns a value of the evaluated expression on the context of the root module.
func (s *Server) EvalExprOnRootCtx(req *tfplugin.EvalExprRequest, resp *tfplugin.EvalExprResponse) error {
	expr, diags := tflint.ParseExpression(req.Expr, req.ExprRange.Filename, req.ExprRange.Start)
	if diags.HasErrors() {
		return diags
	}

	val, err := s.rootRunner.EvalExpr(expr, req.Ret, cty.Type{})
	if err != nil {
		if appErr, ok := err.(*tflint.Error); ok {
			err = client.Error(*appErr)
		}
	}
	*resp = tfplugin.EvalExprResponse{Val: val, Err: err}
	return nil
}

// IsNullExpr returns the result of determining whether the expression is null or not.
func (s *Server) IsNullExpr(req *tfplugin.IsNullExprRequest, resp *tfplugin.IsNullExprResponse) error {
	expr, diags := tflint.ParseExpression(req.Expr, req.Range.Filename, req.Range.Start)
	if diags.HasErrors() {
		return diags
	}

	ret, err := s.runner.IsNullExpr(expr)
	*resp = tfplugin.IsNullExprResponse{Ret: ret, Err: err}
	return nil
}

// EmitIssue reflects a issue to the Runner
func (s *Server) EmitIssue(req *tfplugin.EmitIssueRequest, resp *interface{}) error {
	if req.Expr != nil {
		expr, diags := tflint.ParseExpression(req.Expr, req.ExprRange.Filename, req.ExprRange.Start)
		if diags.HasErrors() {
			return diags
		}

		s.runner.WithExpressionContext(expr, func() error {
			s.runner.EmitIssue(req.Rule, req.Message, req.Location)
			return nil
		})
	} else {
		s.runner.EmitIssue(req.Rule, req.Message, req.Location)
		return nil
	}
	return nil
}

func configBodyRange(body hcl.Body) hcl.Range {
	var bodyRange hcl.Range

	// Estimate the range of the body from the range of all attributes and blocks.
	hclBody := body.(*hclsyntax.Body)
	for _, attr := range hclBody.Attributes {
		if bodyRange.Empty() {
			bodyRange = attr.Range()
		} else {
			bodyRange = hcl.RangeOver(bodyRange, attr.Range())
		}
	}
	for _, block := range hclBody.Blocks {
		if bodyRange.Empty() {
			bodyRange = block.Range()
		} else {
			bodyRange = hcl.RangeOver(bodyRange, block.Range())
		}
	}
	return bodyRange
}
