package plugin

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"github.com/wata727/tflint/issue"
)

type RuleCollection interface {
	Process([]string) []*issue.Issue
}

type RuleCollectionRPC struct{ client *rpc.Client }

func (r *RuleCollectionRPC) Process(files []string) []*issue.Issue {
	var resp []*issue.Issue
	err := r.client.Call("Plugin.Process", files, &resp)
	if err != nil {
		panic(err)
	}

	return resp
}

type RuleCollectionRPCServer struct {
	Impl RuleCollection
}

func (s *RuleCollectionRPCServer) Process(files []string, resp *[]*issue.Issue) error {
	*resp = s.Impl.Process(files)
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
