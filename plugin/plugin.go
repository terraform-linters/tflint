package plugin

import (
	"net/rpc"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/wata727/tflint/rules"
	"github.com/wata727/tflint/tflint"
)

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "TFLINT_RULESET_PLUGIN",
	MagicCookieValue: "5adSn1bX8nrDfgBqiAqqEkC6OE1h3iD8SqbMc5UUONx8x3xCF0KlPDsBRNDjoYDP",
}

type Response struct {
	Issues tflint.Issues
	Error  error
}

type RPCClient struct {
	raw *rpc.Client
}

func (c *RPCClient) Check(runner *tflint.Runner) (tflint.Issues, error) {
	var resp Response
	err := c.raw.Call("Plugin.Check", runner, &resp)
	if err != nil {
		return tflint.Issues{}, err
	}

	return resp.Issues, resp.Error
}

type RPCServer struct {
	RuleSet rules.RuleSet
}

func (s *RPCServer) Check(runner *tflint.Runner, resp *Response) error {
	issues, err := s.RuleSet.Check(runner)
	*resp = Response{Issues: issues, Error: err}
	return nil
}

type RuleSetPlugin struct {
	RuleSet rules.RuleSet
}

func (p *RuleSetPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &RPCServer{RuleSet: p.RuleSet}, nil
}

func (RuleSetPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &RPCClient{raw: c}, nil
}
