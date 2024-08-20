package logger

import (
	"os"

	"github.com/hashicorp/go-hclog"
)

// internalLogger is intended to be called via the public methods of the package.
// So the output line will be the caller of this package.
var internalLogger hclog.Logger

// logger is inteded to be called directly.
// It is mainly assumed to be used by go-plugin.
var logger hclog.Logger

// Use the init process to set the global logger.
// It is expected to be initialized when the plugin starts
// and you need to import the package in the proper order.
func init() {
	level := os.Getenv("TFLINT_LOG")
	if level == "" {
		// Do not emit logs by default
		level = "off"
	}

	internalLogger = hclog.New(&hclog.LoggerOptions{
		Level:                    hclog.LevelFromString(level),
		Output:                   os.Stderr,
		TimeFormat:               "15:04:05",
		IncludeLocation:          true,
		AdditionalLocationOffset: 1,
	})
	logger = hclog.New(&hclog.LoggerOptions{
		Level:           hclog.LevelFromString(level),
		Output:          os.Stderr,
		TimeFormat:      "15:04:05",
		IncludeLocation: true,
	})
}

// Logger returns hcl.Logger
func Logger() hclog.Logger {
	return logger
}

// Trace emits a message at the TRACE level
func Trace(msg string, args ...interface{}) {
	if internalLogger == nil {
		return
	}
	internalLogger.Trace(msg, args...)
}

// Debug emits a message at the DEBUG level
func Debug(msg string, args ...interface{}) {
	if internalLogger == nil {
		return
	}
	internalLogger.Debug(msg, args...)
}

// Info emits a message at the INFO level
func Info(msg string, args ...interface{}) {
	if internalLogger == nil {
		return
	}
	internalLogger.Info(msg, args...)
}

// Warn emits a message at the WARN level
func Warn(msg string, args ...interface{}) {
	if internalLogger == nil {
		return
	}
	internalLogger.Warn(msg, args...)
}

// Error emits a message at the ERROR level
func Error(msg string, args ...interface{}) {
	if internalLogger == nil {
		return
	}
	internalLogger.Error(msg, args...)
}
