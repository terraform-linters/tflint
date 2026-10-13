package plugin

import (
	plugin "github.com/hashicorp/go-plugin"
	"github.com/hashicorp/go-version"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/plugin2host"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
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
	RuleSets map[string]RuleSet

	clients map[string]*plugin.Client
}

type RuleSet interface {
	RuleSetName() (string, error)
	RuleSetVersion() (string, error)
	RuleNames() ([]string, error)
	VersionConstraints() (version.Constraints, error)
	SDKVersion() (*version.Version, error)
	ConfigSchema() (*hclext.BodySchema, error)
	ApplyGlobalConfig(*sdk.Config) error
	ApplyConfig(*hclext.BodyContent, map[string][]byte) error
	Check(plugin2host.Server) error
}

// Clean is a helper for ending plugin processes
func (p *Plugin) Clean() {
	for _, client := range p.clients {
		client.Kill()
	}
}
