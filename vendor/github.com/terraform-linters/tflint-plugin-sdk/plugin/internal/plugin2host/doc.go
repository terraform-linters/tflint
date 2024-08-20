// Package plugin2host contains a gRPC server (host) and client (plugin).
//
// Communication from the plugin to the host is the second one that occurs.
// To understand what happens first, see the host2plugin package first.
// The gRPC client used by the plugin is implicitly initialized by the host2plugin
// package and hidden in the tflint.Runner interface. Normally, plugin developers
// do not need to be aware of the details of this client.
//
// The host starts a gRPC server as goroutine to respond from the plugin side
// when calling Check function in host2plugin. Please note that the gRPC server
// and client startup in plugin2host is not due to go-plugin.
package plugin2host
