package plugin

import (
	"os/exec"

	plugin "github.com/hashicorp/go-plugin"
)

type ClientOpts struct {
	Cmd *exec.Cmd
}

func Client(opts *ClientOpts) *plugin.Client {
	return plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"ruleset": &RuleSetPlugin{},
		},
		Cmd: opts.Cmd,
	})
}
