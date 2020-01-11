package tflint

import "fmt"

const (
	// EvaluationError is an error when interpolation failed (unexpected)
	EvaluationError string = "E:Evaluation"
	// UnknownValueError is an error when an unknown value is referenced
	UnknownValueError string = "W:UnknownValue"
	// NullValueError is an error when null value is referenced
	NullValueError string = "W:NullValue"
	// TypeConversionError is an error when type conversion of cty.Value failed
	TypeConversionError string = "E:TypeConversion"
	// TypeMismatchError is an error when a type of cty.Value is not as expected
	TypeMismatchError string = "E:TypeMismatch"
	// UnevaluableError is an error when a received expression has unevaluable references.
	UnevaluableError string = "W:Unevaluable"
	// UnexpectedAttributeError is an error when handle unexpected attributes (e.g. block)
	UnexpectedAttributeError string = "E:UnexpectedAttribute"
	// ExternalAPIError is an error when calling the external API (e.g. AWS SDK)
	ExternalAPIError string = "E:ExternalAPI"
	// ContextError is pseudo error code for propagating runtime context.
	ContextError string = "I:Context"

	// FatalLevel is a recorverable error, it cause panic
	FatalLevel string = "Fatal"
	// ErrorLevel is a user-level error, it display and feedback error information
	ErrorLevel string = "Error"
	// WarningLevel is a user-level warning. Although it is an error, it has no effect on execution.
	WarningLevel string = "Warning"
)

// Error is application error object. It has own error code
// for processing according to a type of error.
type Error struct {
	Code    string
	Level   string
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

// NewContextError makes a new context error
func NewContextError(msg string, cause error) *Error {
	return &Error{
		Code:    ContextError,
		Level:   ErrorLevel,
		Message: msg,
		Cause:   cause,
	}
}
