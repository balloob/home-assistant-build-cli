package client

// Error code constants for structured error output.
// REST-originated codes are already used by rest.go handleError().
// The remaining codes are used by the centralized error handler in cmd/root.go.
const (
	// REST API codes (existing, used in rest.go)
	ErrCodeAuthenticationError = "AUTHENTICATION_ERROR"
	ErrCodePermissionDenied    = "PERMISSION_DENIED"
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeValidationError     = "VALIDATION_ERROR"
	ErrCodeAPIError            = "API_ERROR"

	// Centralized handler codes
	ErrCodeAuthRequired    = "AUTH_REQUIRED"    // auth.ErrNotAuthenticated sentinel
	ErrCodeConnectionError = "CONNECTION_ERROR" // websocket/network failures
	ErrCodeTimeout         = "TIMEOUT"          // command timed out
	ErrCodeCancelled       = "CANCELLED"        // user cancelled (e.g. deletion prompt)
	ErrCodeInputError      = "INPUT_ERROR"      // invalid JSON/YAML input, parse failures
	ErrCodeUnknownError    = "UNKNOWN_ERROR"    // fallback for unclassified errors
)

// NewError creates a new APIError with the given code and message.
func NewError(code, message string) *APIError {
	return &APIError{Code: code, Message: message}
}
