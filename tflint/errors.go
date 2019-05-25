package tflint

import "fmt"

const (
	// EvaluationError is an error when interpolation failed (unexpected)
	EvaluationError int = 0
	// UnknownValueError is an error when an unknown value is referenced
	UnknownValueError int = 1 + iota
	// NullValueError is an error when null value is referenced
	NullValueError
	// TypeConversionError is an error when type conversion of cty.Value failed
	TypeConversionError
	// TypeMismatchError is an error when a type of cty.Value is not as expected
	TypeMismatchError
	// UnevaluableError is an error when a received expression has unevaluable references.
	UnevaluableError
	// UnexpectedAttributeError is an error when handle unexpected attributes (e.g. block)
	UnexpectedAttributeError
	// ExternalAPIError is an error when calling the external API (e.g. AWS SDK)
	ExternalAPIError

	// FatalLevel is a recorverable error, it cause panic
	FatalLevel int = 0
	// ErrorLevel is a user-level error, it display and feedback error information
	ErrorLevel int = 1 + iota
	// WarningLevel is a user-level warning. Although it is an error, it has no effect on execution.
	WarningLevel
)

// Error is application error object. It has own error code
// for processing according to a type of error.
type Error struct {
	Code    int
	Level   int
	Message string
	Cause   error
}

// Error shows error message. This must be implemented for error interface.
func (e *Error) Error() string {
	if e.Message != "" && e.Cause != nil {
		return fmt.Sprintf("%s; %s", e.Message, e.Cause)
	}

	if e.Message == "" && e.Cause != nil {
		return e.Cause.Error()
	}

	return e.Message
}
