package plugin

import (
	"os"
	"os/exec"

	hclog "github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	"github.com/wata727/tflint/plugin/discovery"
)

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "TFLINT_PLUGIN_COOKIE",
	MagicCookieValue: "OTFhOGZkMzU3NDVmYTIyZDM4NGQwOTBhMmZlOGMzYmIzOGU1N2Y5NDY2MzVhYmZm",
}

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

func Client(m discovery.PluginMeta) *plugin.Client {
	return plugin.NewClient(ClientConfig(m))
}
