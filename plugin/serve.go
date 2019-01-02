package plugin

import (
	"github.com/hashicorp/go-plugin"
)

/*
Serve helps plugins serve their contents without having to worry about the implementation details
of the host/client communication configurations.
*/
func Serve(ruleCollection RuleCollection) {
	plugins := map[int]plugin.PluginSet{
		1: map[string]plugin.Plugin{
			"rules": &RuleCollectionPlugin{
				Impl: ruleCollection,
			},
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig:  Handshake,
		VersionedPlugins: plugins,
	})
}
