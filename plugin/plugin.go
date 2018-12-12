package plugin

import (
	"github.com/hashicorp/go-plugin"
)

var VersionedPlugins = map[int]plugin.PluginSet{
	1: {
		"rules": &RuleCollectionPlugin{},
	},
}
