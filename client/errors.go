package client

import (
	"errors"
	"maps"
)

// errUnexpectedResponse is returned when an API response cannot be
// type-asserted to the expected Go type (map or slice).
var errUnexpectedResponse = errors.New("unexpected response type")

// Error code constants for structured error output.
// REST-originated codes are used by rest.go handleError().
// The remaining codes are used by the centralized error handler in cmd/root.go.
const (
	// REST API codes (existing, used in rest.go)
	ErrCodeAuthenticationError = "AUTHENTICATION_ERROR"
	ErrCodePermissionDenied    = "PERMISSION_DENIED"
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeValidationError     = "VALIDATION_ERROR"
	ErrCodeAPIError            = "API_ERROR"

	// Centralized handler codes
	ErrCodeAuthRequired         = "AUTH_REQUIRED"         // auth.ErrNotAuthenticated sentinel
	ErrCodeConnectionError      = "CONNECTION_ERROR"      // websocket/network failures
	ErrCodeTimeout              = "TIMEOUT"               // command timed out
	ErrCodeCancelled            = "CANCELLED"             // user cancelled (e.g. deletion prompt)
	ErrCodeConfirmationRequired = "CONFIRMATION_REQUIRED" // non-interactive destructive command without --force
	ErrCodeInputError           = "INPUT_ERROR"           // invalid JSON/YAML input, parse failures
	ErrCodeUnknownError         = "UNKNOWN_ERROR"         // fallback for unclassified errors
)

// APIError represents a structured API error with a machine-readable code
// and a human-readable message.  It is used by both the REST and WebSocket
// clients and can be inspected by callers via errors.As.
type APIError struct {
	Code                string
	Message             string
	Details             map[string]any
	Category            string
	Retryable           bool
	LikelyCause         string
	SuggestedFix        string
	SuggestedCommands   []string
	PrerequisiteMissing string
	Transport           string
	StatusCode          int
}

func (e *APIError) Error() string {
	return e.Message
}

// NewError creates a new APIError with the given code and message.
func NewError(code, message string) *APIError {
	return &APIError{Code: code, Message: message}
}

// NewDetailedError creates a new APIError with structured details.
func NewDetailedError(code, message string, details map[string]any) *APIError {
	return &APIError{Code: code, Message: message, Details: details}
}

// DetailsMap returns the details map enriched with structured APIError fields.
func (e *APIError) DetailsMap() map[string]any {
	if e == nil {
		return nil
	}

	var details map[string]any
	if len(e.Details) > 0 {
		details = maps.Clone(e.Details)
	} else {
		details = make(map[string]any)
	}

	if e.Category != "" {
		details["category"] = e.Category
	}
	if e.Retryable {
		details["retryable"] = true
	}
	if e.LikelyCause != "" {
		details["likely_cause"] = e.LikelyCause
	}
	if e.SuggestedFix != "" {
		details["suggested_fix"] = e.SuggestedFix
	}
	if len(e.SuggestedCommands) > 0 {
		details["suggested_commands"] = e.SuggestedCommands
	}
	if e.PrerequisiteMissing != "" {
		details["prerequisite_missing"] = e.PrerequisiteMissing
	}
	if e.Transport != "" {
		details["transport"] = e.Transport
	}
	if e.StatusCode > 0 {
		details["status_code"] = e.StatusCode
	}

	if len(details) == 0 {
		return nil
	}
	return details
}

// NewCancelledError returns a standardized user-cancelled error.
func NewCancelledError(action string) *APIError {
	details := map[string]any{}
	if action != "" {
		details["action"] = action
	}
	return &APIError{
		Code:         ErrCodeCancelled,
		Message:      "Operation cancelled by user.",
		Details:      details,
		Category:     "cancellation",
		Retryable:    true,
		SuggestedFix: "Re-run the command and confirm the prompt, or use --force when you explicitly want to skip confirmation.",
	}
}

// NewConfirmationRequiredError returns a standardized error for destructive
// commands that cannot prompt in non-interactive mode.
func NewConfirmationRequiredError(action, prompt string) *APIError {
	details := map[string]any{}
	if action != "" {
		details["action"] = action
	}
	if prompt != "" {
		details["prompt"] = prompt
	}
	return &APIError{
		Code:         ErrCodeConfirmationRequired,
		Message:      "Confirmation required in non-interactive mode.",
		Details:      details,
		Category:     "confirmation",
		Retryable:    true,
		SuggestedFix: "Re-run the command with --force if you have already validated the target, or run it interactively to confirm the prompt.",
	}
}
