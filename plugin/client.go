package plugin

import (
	"os"
	"os/exec"

	hclog "github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	"github.com/wata727/tflint/plugin/discovery"
)

/*
The Handshake variable contains values that are used to validate authenticity between a client
and a plugin.

This is more of a user experience feature that helps tflint to not run any ol' binary as a plugin.
*/
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "TFLINT_PLUGIN_COOKIE",
	MagicCookieValue: "OTFhOGZkMzU3NDVmYTIyZDM4NGQwOTBhMmZlOGMzYmIzOGU1N2Y5NDY2MzVhYmZm",
}

/*
ClientConfig returns a go-plugin configuration struct.
*/
func ClientConfig(m discovery.PluginMeta) *plugin.ClientConfig {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "tflint_plugin",
		Level:  hclog.Info,
		Output: os.Stderr,
	})

	return &plugin.ClientConfig{
		Cmd:              exec.Command(m.Path),
		HandshakeConfig:  Handshake,
		Managed:          true,
		VersionedPlugins: VersionedPlugins,
		Logger:           logger,
	}
}

/*
Client is a convenience function for establishing a relationship between a host and a plugin.
*/
func Client(m discovery.PluginMeta) *plugin.Client {
	return plugin.NewClient(ClientConfig(m))
}
