package plugin

import (
	hcl "github.com/hashicorp/hcl/v2"
	tfplugin "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/tflint"
)

// Server is a RPC server for responding to requests from plugins
type Server struct {
	runner *tflint.Runner
}

// NewServer initializes a RPC server for plugins
func NewServer(runner *tflint.Runner) *Server {
	return &Server{runner: runner}
}

// Attributes returns corresponding hcl.Attributes
func (s *Server) Attributes(req *tfplugin.AttributesRequest, resp *tfplugin.AttributesResponse) error {
	ret := []*hcl.Attribute{}
	err := s.runner.WalkResourceAttributes(req.Resource, req.AttributeName, func(attr *hcl.Attribute) error {
		ret = append(ret, attr)
		return nil
	})
	*resp = tfplugin.AttributesResponse{Attributes: ret, Err: err}
	return nil
}

// EvalExpr returns a value of the evaluated expression
func (s *Server) EvalExpr(req *tfplugin.EvalExprRequest, resp *tfplugin.EvalExprResponse) error {
	val, err := s.runner.EvalExpr(req.Expr, req.Ret)
	if err != nil {
		if appErr, ok := err.(*tflint.Error); ok {
			err = tfplugin.Error(*appErr)
		}
	}
	*resp = tfplugin.EvalExprResponse{Val: val, Err: err}
	return nil
}

// EmitIssue reflects a issue to the Runner
func (s *Server) EmitIssue(req *tfplugin.EmitIssueRequest, resp *interface{}) error {
	s.runner.WithExpressionContext(req.Meta.Expr, func() error {
		s.runner.EmitIssue(req.Rule, req.Message, req.Location)
		return nil
	})
	return nil
}
