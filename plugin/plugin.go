package plugin

import (
	"github.com/hashicorp/go-plugin"
)

/*
The expected version for plugins.
*/
var VersionedPlugins = map[int]plugin.PluginSet{
	1: {
		"rules": &RuleCollectionPlugin{},
	},
}
