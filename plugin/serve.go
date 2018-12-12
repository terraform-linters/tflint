package plugin

import (
	"github.com/hashicorp/go-plugin"
)

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
