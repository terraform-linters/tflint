// Package plugin contains the implementations needed to make
// the built binary act as a plugin.
//
// A plugin is implemented as an gRPC server and the host acts
// as the client, sending analysis requests to the plugin.
//
// See internal/host2plugin for implementation details.
package plugin
