package plugin

import (
	plugin "github.com/hashicorp/go-plugin"
	tfplugin "github.com/terraform-linters/tflint-plugin-sdk/plugin"
)

// PluginRoot is the root directory of the plugins
// This variable is exposed for testing.
var PluginRoot = "~/.tflint.d/plugins"
var localPluginRoot = "./.tflint.d/plugins"

// Plugin is an object handling plugins
// Basically, it is a wrapper for go-plugin and provides an API to handle them collectively.
type Plugin struct {
	RuleSets map[string]*tfplugin.Client

	clients map[string]*plugin.Client
}

// Clean is a helper for ending plugin processes
func (p *Plugin) Clean() {
	for _, client := range p.clients {
		client.Kill()
	}
}
