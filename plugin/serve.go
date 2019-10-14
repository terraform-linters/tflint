package plugin

import (
	plugin "github.com/hashicorp/go-plugin"
	"github.com/wata727/tflint/rules"
)

// ServeOpts is an option provided by plug-ins
type ServeOpts struct {
	RuleSet rules.RuleSet
}

// Serve is a wrapper of plugin.Serve. This is entrypoint of all plug-ins
func Serve(opts *ServeOpts) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"ruleset": &RuleSetPlugin{RuleSet: opts.RuleSet},
		},
	})
}
