// Package host2plugin contains a gRPC server (plugin) and client (host).
//
// In the plugin system, this communication is the first thing that happens,
// and a plugin must use this package to provide a gRPC server.
// However, the detailed implementation is hidden in the tflint.RuleSet interface,
// and plugin developers usually don't need to be aware of gRPC server behavior.
//
// When the host initializes a gRPC client, go-plugin starts a gRPC server
// on the plugin side as another process. This package acts as a wrapper for go-plugin.
// Separately, the Check function initializes a new gRPC client for plugin-to-host
// communication. See the plugin2host package for details.
package host2plugin
