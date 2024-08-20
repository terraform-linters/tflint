package plugin

import (
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/internal/host2plugin"

	// Import this package to initialize the global logger
	_ "github.com/terraform-linters/tflint-plugin-sdk/logger"
)

// ServeOpts is an option for serving a plugin.
// Each plugin can pass a RuleSet that represents its own functionality.
type ServeOpts = host2plugin.ServeOpts

// Serve is a wrapper of plugin.Serve. This is entrypoint of all plugins.
var Serve = host2plugin.Serve

// SDKVersion is the SDK version.
const SDKVersion = host2plugin.SDKVersion
