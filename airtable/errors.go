package airtable

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Operation represents the API operation that failed.
type Operation string

const (
	OpList      Operation = "List"
	OpGet       Operation = "Get"
	OpCreate    Operation = "Create"
	OpUpdate    Operation = "Update"
	OpReplace   Operation = "Replace"
	OpDestroy   Operation = "Destroy"
	OpGetFields Operation = "GetFields"
	OpSave      Operation = "Save"
	OpFind      Operation = "Find"
)

// ErrorType represents Airtable API error types.
type ErrorType string

const (
	ErrTypeInvalidPermissions    ErrorType = "INVALID_PERMISSIONS_OR_MODEL_NOT_FOUND"
	ErrTypeInvalidRequest        ErrorType = "INVALID_REQUEST"
	ErrTypeInvalidRequestUnknown ErrorType = "INVALID_REQUEST_UNKNOWN"
	ErrTypeNotFound              ErrorType = "NOT_FOUND"
	ErrTypeInvalidValue          ErrorType = "INVALID_VALUE"
	ErrTypeAuthRequired          ErrorType = "AUTHENTICATION_REQUIRED"
	ErrTypeUnauthorized          ErrorType = "UNAUTHORIZED"
	ErrTypeRateLimited           ErrorType = "RATE_LIMITED"
	ErrTypeUnknown               ErrorType = "UNKNOWN"
)

// Sentinel errors for common conditions.
var (
	ErrNotConfigured   = errors.New("airtable client not configured")
	ErrMissingRecordID = errors.New("record ID required")
	ErrNotFound        = errors.New("resource not found")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrRateLimited     = errors.New("rate limited")
	ErrValidation      = errors.New("validation error")
)

// Error is the base error type for all airtable errors.
type Error struct {
	Op         Operation // Operation that failed (List, Get, Create, etc.)
	StatusCode int       // HTTP status code (0 if not applicable)
	Type       ErrorType // Airtable error type string
	Message    string    // Airtable error message or description
	Err        error     // Underlying error
}

func (e *Error) Error() string {
	if e.StatusCode != 0 && e.Type != "" {
		return fmt.Sprintf("airtable.%s: [%d] %s: %s", e.Op, e.StatusCode, e.Type, e.Message)
	}
	if e.StatusCode != 0 {
		return fmt.Sprintf("airtable.%s: [%d] %s", e.Op, e.StatusCode, e.Message)
	}
	if e.Message != "" {
		return fmt.Sprintf("airtable.%s: %s", e.Op, e.Message)
	}
	if e.Err != nil {
		return fmt.Sprintf("airtable.%s: %v", e.Op, e.Err)
	}
	return fmt.Sprintf("airtable.%s: unknown error", e.Op)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// Is enables errors.Is() checking against sentinel errors.
func (e *Error) Is(target error) bool {
	switch target {
	case ErrNotConfigured:
		return e.Err == ErrNotConfigured
	case ErrMissingRecordID:
		return e.Err == ErrMissingRecordID
	case ErrNotFound:
		return e.StatusCode == 404 || e.Type == ErrTypeNotFound
	case ErrUnauthorized:
		return e.StatusCode == 401 || e.StatusCode == 403 ||
			e.Type == ErrTypeUnauthorized || e.Type == ErrTypeAuthRequired ||
			e.Type == ErrTypeInvalidPermissions
	case ErrRateLimited:
		return e.StatusCode == 429 || e.Type == ErrTypeRateLimited
	case ErrValidation:
		return e.Type == ErrTypeInvalidValue || e.Type == ErrTypeInvalidRequest
	}
	return false
}

// APIError represents an error response from the Airtable API.
type APIError struct {
	Op         Operation
	StatusCode int
	Type       ErrorType
	Message    string
	Err        error
}

func (e *APIError) Error() string {
	if e.StatusCode != 0 && e.Type != "" {
		return fmt.Sprintf("airtable.%s: [%d] %s: %s", e.Op, e.StatusCode, e.Type, e.Message)
	}
	if e.StatusCode != 0 {
		return fmt.Sprintf("airtable.%s: [%d] %s", e.Op, e.StatusCode, e.Message)
	}
	if e.Message != "" {
		return fmt.Sprintf("airtable.%s: %s", e.Op, e.Message)
	}
	return fmt.Sprintf("airtable.%s: unknown error", e.Op)
}

func (e *APIError) Unwrap() error {
	return e.Err
}

func (e *APIError) Is(target error) bool {
	switch target {
	case ErrNotFound:
		return e.StatusCode == 404 || e.Type == ErrTypeNotFound
	case ErrUnauthorized:
		return e.StatusCode == 401 || e.StatusCode == 403 ||
			e.Type == ErrTypeUnauthorized || e.Type == ErrTypeAuthRequired ||
			e.Type == ErrTypeInvalidPermissions
	case ErrRateLimited:
		return e.StatusCode == 429 || e.Type == ErrTypeRateLimited
	case ErrValidation:
		return e.Type == ErrTypeInvalidValue || e.Type == ErrTypeInvalidRequest
	}
	return false
}

// HTTPError represents an HTTP-level error (network, timeout, etc.).
type HTTPError struct {
	Op         Operation
	StatusCode int
	Message    string
	Err        error
}

func (e *HTTPError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("airtable.%s: [%d] %s", e.Op, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("airtable.%s: HTTP %d", e.Op, e.StatusCode)
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

// ValidationError represents a local validation failure.
type ValidationError struct {
	Op      Operation
	Message string
	Field   string // Optional: which field failed validation
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("airtable.%s: %s (field: %s)", e.Op, e.Message, e.Field)
	}
	return fmt.Sprintf("airtable.%s: %s", e.Op, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

func (e *ValidationError) Is(target error) bool {
	if target == ErrValidation {
		return true
	}
	if target == ErrMissingRecordID {
		return e.Err == ErrMissingRecordID
	}
	return false
}

// ConfigError represents a configuration error.
type ConfigError struct {
	Op      Operation
	Message string
	Err     error
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("airtable.%s: %s", e.Op, e.Message)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

func (e *ConfigError) Is(target error) bool {
	return target == ErrNotConfigured
}

// NewAPIError creates an error from an Airtable API response.
func NewAPIError(op Operation, statusCode int, errType, message string) *APIError {
	return &APIError{
		Op:         op,
		StatusCode: statusCode,
		Type:       ErrorType(errType),
		Message:    message,
	}
}

// NewHTTPError creates an error from an HTTP failure.
func NewHTTPError(op Operation, statusCode int, err error) *HTTPError {
	return &HTTPError{
		Op:         op,
		StatusCode: statusCode,
		Message:    fmt.Sprintf("HTTP %d", statusCode),
		Err:        err,
	}
}

// NewValidationError creates a local validation error.
func NewValidationError(op Operation, message string) *ValidationError {
	return &ValidationError{
		Op:      op,
		Message: message,
		Err:     ErrValidation,
	}
}

// NewValidationErrorWithField creates a local validation error with field info.
func NewValidationErrorWithField(op Operation, message, field string) *ValidationError {
	return &ValidationError{
		Op:      op,
		Message: message,
		Field:   field,
		Err:     ErrValidation,
	}
}

// NewConfigError creates a configuration error.
func NewConfigError(op Operation, message string) *ConfigError {
	return &ConfigError{
		Op:      op,
		Message: message,
		Err:     ErrNotConfigured,
	}
}

// WrapError wraps an existing error with operation context.
func WrapError(op Operation, err error) *Error {
	if err == nil {
		return nil
	}

	// If it's already our error type, just update the operation if needed
	var airtableErr *Error
	if errors.As(err, &airtableErr) {
		if airtableErr.Op == "" {
			airtableErr.Op = op
		}
		return airtableErr
	}

	return &Error{
		Op:  op,
		Err: err,
	}
}

// airtableErrorResponse represents the error structure from Airtable API.
type airtableErrorResponse struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// ParseAPIError parses an Airtable error response body.
func ParseAPIError(op Operation, statusCode int, body []byte) *APIError {
	var errResp airtableErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return &APIError{
			Op:         op,
			StatusCode: statusCode,
			Type:       ErrTypeUnknown,
			Message:    string(body),
		}
	}

	errType := errResp.Error.Type
	if errType == "" {
		errType = string(ErrTypeUnknown)
	}

	return NewAPIError(op, statusCode, errType, errResp.Error.Message)
}
