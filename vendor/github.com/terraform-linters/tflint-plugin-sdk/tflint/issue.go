package tflint

// Severity indicates the severity of the issue.
type Severity int32

const (
	// ERROR is possible errors
	ERROR Severity = iota
	// WARNING doesn't cause problem immediately, but not good
	WARNING
	// NOTICE is not important, it's mentioned
	NOTICE
)

// String returns the string representation of the severity.
func (s Severity) String() string {
	switch s {
	case ERROR:
		return "Error"
	case WARNING:
		return "Warning"
	case NOTICE:
		return "Notice"
	}

	return "Unknown"
}
