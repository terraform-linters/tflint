// Package host2plugin exposes a gRPC client for use on a host (TFLint).
//
// The implementation details are hidden in internal/host2plugin and
// the exposed ones are minimal. They are not intended to be used by plugins.
// For that reason, this package is subject to breaking changes without notice,
// and the changes do not follow the SDK versioning policy.
package host2plugin
