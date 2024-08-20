package tflint

import (
	"errors"
)

// List of errors returned by TFLint.
var (
	// ErrUnknownValue is an error that occurs when decoding an unknown value to a Go value.
	ErrUnknownValue = errors.New("unknown value found")
	// ErrNullValue is an error that occurs when decoding null to a Go value.
	ErrNullValue = errors.New("null value found")
	// ErrUnevaluable is an error that occurs when decoding an unevaluable value to a Go value.
	//
	// Deprecated: This error is no longer returned since TFLint v0.41.
	ErrUnevaluable = errors.New("")
	// ErrSensitive is an error that occurs when decoding a sensitive value to a Go value.
	ErrSensitive = errors.New("sensitive value found")
)

var (
	// ErrFixNotSupported is an error to return if autofix is not supported.
	// This can prevent the issue from being marked as fixable by returning it
	// in FixFunc when autofix cannot be implemented, such as with JSON syntax.
	ErrFixNotSupported = errors.New("autofix is not supported")
)
