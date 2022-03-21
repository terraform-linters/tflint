package plugin

import (
	plugin "github.com/hashicorp/go-plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/host2plugin"
)

// PluginRoot is the root directory of the plugins
// This variable is exposed for testing.
var (
	PluginRoot      = "~/.tflint.d/plugins"
	localPluginRoot = "./.tflint.d/plugins"
)

// Plugin is an object handling plugins
// Basically, it is a wrapper for go-plugin and provides an API to handle them collectively.
type Plugin struct {
	RuleSets map[string]*host2plugin.GRPCClient

	clients map[string]*plugin.Client
}

// Clean is a helper for ending plugin processes
func (p *Plugin) Clean() {
	for _, client := range p.clients {
		client.Kill()
	}
}
