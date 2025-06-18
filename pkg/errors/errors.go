package errors

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	// Configuration errors
	ErrorTypeConfig         ErrorType = "CONFIG"
	ErrorTypeValidation     ErrorType = "VALIDATION"
	ErrorTypeProfile        ErrorType = "PROFILE"
	
	// Network and connectivity errors
	ErrorTypeNetwork        ErrorType = "NETWORK"
	ErrorTypeEndpoint       ErrorType = "ENDPOINT"
	ErrorTypeConnection     ErrorType = "CONNECTION"
	ErrorTypeTimeout        ErrorType = "TIMEOUT"
	
	// Load test execution errors
	ErrorTypeLoadTest       ErrorType = "LOADTEST"
	ErrorTypeTransaction    ErrorType = "TRANSACTION"
	ErrorTypeClientFactory  ErrorType = "CLIENT_FACTORY"
	ErrorTypeBroadcast      ErrorType = "BROADCAST"
	
	// File system and I/O errors
	ErrorTypeFileSystem     ErrorType = "FILESYSTEM"
	ErrorTypePermission     ErrorType = "PERMISSION"
	ErrorTypeSerialization  ErrorType = "SERIALIZATION"
	
	// Internal errors
	ErrorTypeInternal       ErrorType = "INTERNAL"
	ErrorTypeUnknown        ErrorType = "UNKNOWN"
)

// LoadTestError represents a structured error with context
type LoadTestError struct {
	Type        ErrorType `json:"type"`
	Code        string    `json:"code"`
	Message     string    `json:"message"`
	Details     string    `json:"details,omitempty"`
	Cause       error     `json:"cause,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	StackTrace  string    `json:"stack_trace,omitempty"`
	Timestamp   string    `json:"timestamp"`
	Component   string    `json:"component"`
}

// Error implements the error interface
func (e *LoadTestError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s:%s] %s: %s", e.Type, e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s:%s] %s", e.Type, e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *LoadTestError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target
func (e *LoadTestError) Is(target error) bool {
	if t, ok := target.(*LoadTestError); ok {
		return e.Type == t.Type && e.Code == t.Code
	}
	return false
}

// WithContext adds context information to the error
func (e *LoadTestError) WithContext(key string, value interface{}) *LoadTestError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithDetails adds additional details to the error
func (e *LoadTestError) WithDetails(details string) *LoadTestError {
	e.Details = details
	return e
}

// NewError creates a new LoadTestError
func NewError(errorType ErrorType, code, message string) *LoadTestError {
	return &LoadTestError{
		Type:      errorType,
		Code:      code,
		Message:   message,
		Timestamp: fmt.Sprintf("%d", getCurrentTimestamp()),
		Component: getCallerComponent(),
	}
}

// NewErrorWithCause creates a new LoadTestError with an underlying cause
func NewErrorWithCause(errorType ErrorType, code, message string, cause error) *LoadTestError {
	return &LoadTestError{
		Type:       errorType,
		Code:       code,
		Message:    message,
		Cause:      cause,
		Timestamp:  fmt.Sprintf("%d", getCurrentTimestamp()),
		Component:  getCallerComponent(),
		StackTrace: getStackTrace(),
	}
}

// WrapError wraps an existing error with LoadTestError
func WrapError(err error, errorType ErrorType, code, message string) *LoadTestError {
	if err == nil {
		return nil
	}
	
	return &LoadTestError{
		Type:       errorType,
		Code:       code,
		Message:    message,
		Cause:      err,
		Timestamp:  fmt.Sprintf("%d", getCurrentTimestamp()),
		Component:  getCallerComponent(),
		StackTrace: getStackTrace(),
	}
}

// Predefined error constructors for common cases

// Config Errors
func NewConfigError(code, message string) *LoadTestError {
	return NewError(ErrorTypeConfig, code, message)
}

func NewValidationError(code, message string) *LoadTestError {
	return NewError(ErrorTypeValidation, code, message)
}

func NewProfileError(code, message string) *LoadTestError {
	return NewError(ErrorTypeProfile, code, message)
}

// Network Errors
func NewNetworkError(code, message string) *LoadTestError {
	return NewError(ErrorTypeNetwork, code, message)
}

func NewEndpointError(code, message string) *LoadTestError {
	return NewError(ErrorTypeEndpoint, code, message)
}

func NewConnectionError(code, message string) *LoadTestError {
	return NewError(ErrorTypeConnection, code, message)
}

func NewTimeoutError(code, message string) *LoadTestError {
	return NewError(ErrorTypeTimeout, code, message)
}

// Load Test Errors
func NewLoadTestError(code, message string) *LoadTestError {
	return NewError(ErrorTypeLoadTest, code, message)
}

func NewTransactionError(code, message string) *LoadTestError {
	return NewError(ErrorTypeTransaction, code, message)
}

func NewClientFactoryError(code, message string) *LoadTestError {
	return NewError(ErrorTypeClientFactory, code, message)
}

func NewBroadcastError(code, message string) *LoadTestError {
	return NewError(ErrorTypeBroadcast, code, message)
}

// File System Errors
func NewFileSystemError(code, message string) *LoadTestError {
	return NewError(ErrorTypeFileSystem, code, message)
}

func NewPermissionError(code, message string) *LoadTestError {
	return NewError(ErrorTypePermission, code, message)
}

func NewSerializationError(code, message string) *LoadTestError {
	return NewError(ErrorTypeSerialization, code, message)
}

// Internal Errors
func NewInternalError(code, message string) *LoadTestError {
	return NewError(ErrorTypeInternal, code, message)
}

// Error recovery and handling utilities

// IsRecoverable determines if an error is recoverable
func IsRecoverable(err error) bool {
	if loadTestErr, ok := err.(*LoadTestError); ok {
		switch loadTestErr.Type {
		case ErrorTypeTimeout, ErrorTypeConnection, ErrorTypeNetwork:
			return true
		case ErrorTypeValidation, ErrorTypeConfig, ErrorTypePermission:
			return false
		default:
			return false
		}
	}
	return false
}

// GetErrorType extracts the error type from an error
func GetErrorType(err error) ErrorType {
	if loadTestErr, ok := err.(*LoadTestError); ok {
		return loadTestErr.Type
	}
	return ErrorTypeUnknown
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) string {
	if loadTestErr, ok := err.(*LoadTestError); ok {
		return loadTestErr.Code
	}
	return "UNKNOWN"
}

// Utility functions

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

func getCallerComponent() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}
	
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	
	name := fn.Name()
	parts := strings.Split(name, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return name
}

func getStackTrace() string {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// Error codes constants
const (
	// Configuration error codes
	ErrCodeInvalidConfig     = "INVALID_CONFIG"
	ErrCodeMissingConfig     = "MISSING_CONFIG"
	ErrCodeInvalidEndpoint   = "INVALID_ENDPOINT"
	ErrCodeInvalidDuration   = "INVALID_DURATION"
	ErrCodeInvalidRate       = "INVALID_RATE"
	ErrCodeInvalidSize       = "INVALID_SIZE"
	
	// Profile error codes
	ErrCodeProfileNotFound   = "PROFILE_NOT_FOUND"
	ErrCodeProfileExists     = "PROFILE_EXISTS"
	ErrCodeProfileInvalid    = "PROFILE_INVALID"
	ErrCodeProfileSaveFailed = "PROFILE_SAVE_FAILED"
	ErrCodeProfileLoadFailed = "PROFILE_LOAD_FAILED"
	
	// Network error codes
	ErrCodeEndpointUnreachable = "ENDPOINT_UNREACHABLE"
	ErrCodeConnectionFailed    = "CONNECTION_FAILED"
	ErrCodeConnectionTimeout   = "CONNECTION_TIMEOUT"
	ErrCodeNetworkError        = "NETWORK_ERROR"
	
	// Load test error codes
	ErrCodeLoadTestFailed      = "LOADTEST_FAILED"
	ErrCodeTransactionFailed   = "TRANSACTION_FAILED"
	ErrCodeClientFactoryNotFound = "CLIENT_FACTORY_NOT_FOUND"
	ErrCodeBroadcastFailed     = "BROADCAST_FAILED"
	
	// File system error codes
	ErrCodeFileNotFound        = "FILE_NOT_FOUND"
	ErrCodeFileReadFailed      = "FILE_READ_FAILED"
	ErrCodeFileWriteFailed     = "FILE_WRITE_FAILED"
	ErrCodePermissionDenied    = "PERMISSION_DENIED"
	
	// Serialization error codes
	ErrCodeJSONMarshalFailed   = "JSON_MARSHAL_FAILED"
	ErrCodeJSONUnmarshalFailed = "JSON_UNMARSHAL_FAILED"
	ErrCodeYAMLMarshalFailed   = "YAML_MARSHAL_FAILED"
	ErrCodeYAMLUnmarshalFailed = "YAML_UNMARSHAL_FAILED"
	
	// Internal error codes
	ErrCodeInternalError       = "INTERNAL_ERROR"
	ErrCodePanicRecovered      = "PANIC_RECOVERED"
	ErrCodeUnexpectedError     = "UNEXPECTED_ERROR"
) 