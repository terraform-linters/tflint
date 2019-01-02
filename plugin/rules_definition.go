package plugin

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"github.com/wata727/tflint/rules"
	"github.com/wata727/tflint/tflint"
)

/*
The 'real' implementation of a rule collection that is transmitted over rpc.
*/
type RuleCollection interface {
	NewRules(*tflint.Config) []rules.Rule
}

/*
RPC client struct.
*/
type RuleCollectionRPC struct{ client *rpc.Client }

/*
A wrapper for the NewRules function that allows it to be called over an RPC connection.
*/
func (r *RuleCollectionRPC) NewRules(c *tflint.Config) []rules.Rule {
	var resp []rules.Rule
	err := r.client.Call("Plugin.Process", c, &resp)
	if err != nil {
		panic(err)
	}

	return resp
}

type RuleCollectionRPCServer struct {
	Impl RuleCollection
}

func (s *RuleCollectionRPCServer) Process(c *tflint.Config, resp *[]rules.Rule) error {
	*resp = s.Impl.NewRules(c)
	return nil
}

type RuleCollectionPlugin struct {
	Impl RuleCollection
}

func (p *RuleCollectionPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &RuleCollectionRPCServer{Impl: p.Impl}, nil
}

func (RuleCollectionPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &RuleCollectionRPC{client: c}, nil
}
