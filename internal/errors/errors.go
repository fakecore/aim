package errors

import (
	"fmt"
)

// Error represents a structured AIM error
type Error struct {
	Code        string
	Category    string
	Message     string
	Details     map[string]interface{}
	Suggestions []string
	Cause       error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// LogMessage returns the error message with code for logging
func (e *Error) LogMessage() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) ExitCode() int {
	switch e.Category {
	case "CFG":
		return 2
	case "ACC":
		return 3
	case "VEN":
		return 4
	case "TOO":
		return 5
	case "EXE":
		return 6
	case "NET":
		return 7
	case "EXT":
		return 8
	case "SYS":
		return 9
	case "USR":
		return 10
	default:
		return 1
	}
}

// Predefined errors
var (
	ErrConfigNotFound = &Error{
		Code:        "AIM-CFG-001",
		Category:    "CFG",
		Message:     "Config file not found",
		Suggestions: []string{"Run 'aim init' to create a config file"},
	}

	ErrAccountNotFound = &Error{
		Code:        "AIM-ACC-001",
		Category:    "ACC",
		Message:     "Account '%s' not found",
		Suggestions: []string{"Check available accounts with 'aim config show'"},
	}

	ErrKeyNotSet = &Error{
		Code:        "AIM-ACC-002",
		Category:    "ACC",
		Message:     "Account '%s': API key not set",
		Suggestions: []string{"Set environment variable or edit config"},
	}

	ErrVendorNotFound = &Error{
		Code:        "AIM-VEN-001",
		Category:    "VEN",
		Message:     "Vendor '%s' not found",
		Suggestions: []string{"Define vendor in config or use builtin"},
	}

	ErrProtocolNotSupported = &Error{
		Code:        "AIM-VEN-002",
		Category:    "VEN",
		Message:     "Vendor '%s' does not support '%s' protocol",
		Suggestions: []string{"Use a different vendor or tool"},
	}

	ErrToolNotFound = &Error{
		Code:        "AIM-TOO-001",
		Category:    "TOO",
		Message:     "Unknown tool '%s'",
		Suggestions: []string{"Check available tools"},
	}

	ErrCommandNotFound = &Error{
		Code:        "AIM-TOO-002",
		Category:    "TOO",
		Message:     "Command '%s' not found in PATH",
		Suggestions: []string{"Install the tool or check PATH"},
	}

	ErrExecutionTimeout = &Error{
		Code:        "AIM-EXE-003",
		Category:    "EXE",
		Message:     "Command timed out after %s",
		Suggestions: []string{"Increase timeout with --timeout flag"},
	}
)

// Wrap creates a new error with formatted message
func Wrap(err *Error, args ...interface{}) *Error {
	return &Error{
		Code:        err.Code,
		Category:    err.Category,
		Message:     fmt.Sprintf(err.Message, args...),
		Suggestions: err.Suggestions,
	}
}

// WrapWithCause creates a new error with cause
func WrapWithCause(err *Error, cause error, args ...interface{}) *Error {
	e := Wrap(err, args...)
	e.Cause = cause
	return e
}
