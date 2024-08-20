package host2plugin

import "github.com/terraform-linters/tflint-plugin-sdk/plugin/internal/host2plugin"

// Client is a host-side implementation. Host can send requests through the client to plugin's gRPC server.
type Client = host2plugin.GRPCClient

// ClientOpts is an option for initializing a Client.
type ClientOpts = host2plugin.ClientOpts

// NewClient returns a new gRPC client for host-to-plugin communication.
var NewClient = host2plugin.NewClient
