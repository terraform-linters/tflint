// Package logger provides a global logger interface for logging from plugins.
//
// This package is a wrapper for hclog, and it initializes the global logger on import.
// You can freely write logs from anywhere via the public API according to the log level.
// The log by hclog is interpreted as a structured log by go-plugin, and the log level
// can be handled correctly.
package logger
