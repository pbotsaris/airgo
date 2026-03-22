package airtable

import (
	"errors"
	"testing"

	. "github.com/Antfood/airgo/testutils/testutils"
)

func TestErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name: "full API error",
			err: &Error{
				Op:         OpList,
				StatusCode: 401,
				Type:       ErrTypeUnauthorized,
				Message:    "Invalid API key",
			},
			expected: "airtable.List: [401] UNAUTHORIZED: Invalid API key",
		},
		{
			name: "error with status code only",
			err: &Error{
				Op:         OpCreate,
				StatusCode: 500,
				Message:    "Internal server error",
			},
			expected: "airtable.Create: [500] Internal server error",
		},
		{
			name: "error with message only",
			err: &Error{
				Op:      OpUpdate,
				Message: "client not configured",
			},
			expected: "airtable.Update: client not configured",
		},
		{
			name: "error with wrapped error",
			err: &Error{
				Op:  OpGet,
				Err: errors.New("connection refused"),
			},
			expected: "airtable.Get: connection refused",
		},
		{
			name: "minimal error",
			err: &Error{
				Op: OpDestroy,
			},
			expected: "airtable.Destroy: unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Equals(t, tt.err.Error(), tt.expected)
		})
	}
}

func TestErrorIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected bool
	}{
		{
			name: "404 status code",
			err: &Error{
				Op:         OpGet,
				StatusCode: 404,
			},
			expected: true,
		},
		{
			name: "NOT_FOUND type",
			err: &Error{
				Op:   OpGet,
				Type: ErrTypeNotFound,
			},
			expected: true,
		},
		{
			name: "other error",
			err: &Error{
				Op:         OpGet,
				StatusCode: 500,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.err, ErrNotFound)
			Assert(t, result == tt.expected, "expected errors.Is(err, ErrNotFound) = %v, got %v", tt.expected, result)
		})
	}
}

func TestErrorIsUnauthorized(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected bool
	}{
		{
			name: "401 status code",
			err: &Error{
				Op:         OpList,
				StatusCode: 401,
			},
			expected: true,
		},
		{
			name: "403 status code",
			err: &Error{
				Op:         OpList,
				StatusCode: 403,
			},
			expected: true,
		},
		{
			name: "UNAUTHORIZED type",
			err: &Error{
				Op:   OpList,
				Type: ErrTypeUnauthorized,
			},
			expected: true,
		},
		{
			name: "INVALID_PERMISSIONS type",
			err: &Error{
				Op:   OpList,
				Type: ErrTypeInvalidPermissions,
			},
			expected: true,
		},
		{
			name: "other error",
			err: &Error{
				Op:         OpList,
				StatusCode: 500,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.err, ErrUnauthorized)
			Assert(t, result == tt.expected, "expected errors.Is(err, ErrUnauthorized) = %v, got %v", tt.expected, result)
		})
	}
}

func TestErrorIsRateLimited(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected bool
	}{
		{
			name: "429 status code",
			err: &Error{
				Op:         OpList,
				StatusCode: 429,
			},
			expected: true,
		},
		{
			name: "RATE_LIMITED type",
			err: &Error{
				Op:   OpList,
				Type: ErrTypeRateLimited,
			},
			expected: true,
		},
		{
			name: "other error",
			err: &Error{
				Op:         OpList,
				StatusCode: 500,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.err, ErrRateLimited)
			Assert(t, result == tt.expected, "expected errors.Is(err, ErrRateLimited) = %v, got %v", tt.expected, result)
		})
	}
}

func TestErrorIsValidation(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected bool
	}{
		{
			name: "INVALID_VALUE type",
			err: &Error{
				Op:   OpCreate,
				Type: ErrTypeInvalidValue,
			},
			expected: true,
		},
		{
			name: "INVALID_REQUEST type",
			err: &Error{
				Op:   OpCreate,
				Type: ErrTypeInvalidRequest,
			},
			expected: true,
		},
		{
			name: "other error",
			err: &Error{
				Op:         OpCreate,
				StatusCode: 500,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.err, ErrValidation)
			Assert(t, result == tt.expected, "expected errors.Is(err, ErrValidation) = %v, got %v", tt.expected, result)
		})
	}
}

func TestErrorIsNotConfigured(t *testing.T) {
	err := &Error{
		Op:  OpList,
		Err: ErrNotConfigured,
	}

	Assert(t, errors.Is(err, ErrNotConfigured), "expected errors.Is(err, ErrNotConfigured) = true")
}

func TestErrorIsMissingRecordID(t *testing.T) {
	err := &Error{
		Op:  OpGet,
		Err: ErrMissingRecordID,
	}

	Assert(t, errors.Is(err, ErrMissingRecordID), "expected errors.Is(err, ErrMissingRecordID) = true")
}

func TestErrorUnwrap(t *testing.T) {
	inner := errors.New("inner error")
	err := &Error{
		Op:  OpList,
		Err: inner,
	}

	unwrapped := err.Unwrap()
	Assert(t, unwrapped == inner, "expected Unwrap() to return inner error")
}

func TestErrorAs(t *testing.T) {
	t.Run("APIError", func(t *testing.T) {
		err := NewAPIError(OpList, 403, "INVALID_PERMISSIONS_OR_MODEL_NOT_FOUND", "Access denied")

		var apiErr *APIError
		Assert(t, errors.As(err, &apiErr), "expected errors.As to match APIError")
		Equals(t, apiErr.Op, OpList)
		Equals(t, apiErr.StatusCode, 403)
		Equals(t, string(apiErr.Type), "INVALID_PERMISSIONS_OR_MODEL_NOT_FOUND")
		Equals(t, apiErr.Message, "Access denied")
	})

	t.Run("ValidationError", func(t *testing.T) {
		err := NewValidationErrorWithField(OpCreate, "invalid value", "Name")

		var valErr *ValidationError
		Assert(t, errors.As(err, &valErr), "expected errors.As to match ValidationError")
		Equals(t, valErr.Op, OpCreate)
		Equals(t, valErr.Field, "Name")
	})

	t.Run("ConfigError", func(t *testing.T) {
		err := NewConfigError(OpList, "client not configured")

		var cfgErr *ConfigError
		Assert(t, errors.As(err, &cfgErr), "expected errors.As to match ConfigError")
		Equals(t, cfgErr.Op, OpList)
		Assert(t, errors.Is(cfgErr, ErrNotConfigured), "ConfigError should wrap ErrNotConfigured")
	})
}

func TestParseAPIError(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		body           string
		expectedType   ErrorType
		expectedMsg    string
	}{
		{
			name:         "standard Airtable error",
			statusCode:   403,
			body:         `{"error":{"type":"INVALID_PERMISSIONS_OR_MODEL_NOT_FOUND","message":"Access denied"}}`,
			expectedType: ErrTypeInvalidPermissions,
			expectedMsg:  "Access denied",
		},
		{
			name:         "rate limit error",
			statusCode:   429,
			body:         `{"error":{"type":"RATE_LIMITED","message":"Too many requests"}}`,
			expectedType: ErrTypeRateLimited,
			expectedMsg:  "Too many requests",
		},
		{
			name:         "invalid JSON",
			statusCode:   500,
			body:         `invalid json`,
			expectedType: ErrTypeUnknown,
			expectedMsg:  "invalid json",
		},
		{
			name:         "empty error type",
			statusCode:   400,
			body:         `{"error":{"message":"Bad request"}}`,
			expectedType: ErrTypeUnknown,
			expectedMsg:  "Bad request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ParseAPIError(OpList, tt.statusCode, []byte(tt.body))

			Equals(t, err.Op, OpList)
			Equals(t, err.StatusCode, tt.statusCode)
			Equals(t, err.Type, tt.expectedType)
			Equals(t, err.Message, tt.expectedMsg)
		})
	}
}

func TestNewAPIError(t *testing.T) {
	err := NewAPIError(OpCreate, 400, "INVALID_VALUE", "Field value is invalid")

	Equals(t, err.Op, OpCreate)
	Equals(t, err.StatusCode, 400)
	Equals(t, string(err.Type), "INVALID_VALUE")
	Equals(t, err.Message, "Field value is invalid")
}

func TestNewHTTPError(t *testing.T) {
	inner := errors.New("connection timeout")
	err := NewHTTPError(OpList, 500, inner)

	Equals(t, err.Op, OpList)
	Equals(t, err.StatusCode, 500)
	Assert(t, errors.Is(err, inner), "HTTPError should wrap inner error")
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError(OpCreate, "record ID required")

	Equals(t, err.Op, OpCreate)
	Equals(t, err.Message, "record ID required")
	Assert(t, errors.Is(err, ErrValidation), "ValidationError should wrap ErrValidation")
}

func TestNewConfigError(t *testing.T) {
	err := NewConfigError(OpList, "client not configured")

	Equals(t, err.Op, OpList)
	Equals(t, err.Message, "client not configured")
	Assert(t, errors.Is(err, ErrNotConfigured), "ConfigError should wrap ErrNotConfigured")
}

func TestWrapError(t *testing.T) {
	t.Run("wraps standard error", func(t *testing.T) {
		inner := errors.New("something went wrong")
		err := WrapError(OpList, inner)

		Equals(t, err.Op, OpList)
		Assert(t, errors.Is(err, inner), "wrapped error should contain inner error")
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		err := WrapError(OpList, nil)
		Assert(t, err == nil, "WrapError(nil) should return nil")
	})

	t.Run("preserves existing airtable error", func(t *testing.T) {
		inner := &Error{
			Op:      OpGet,
			Message: "original error",
		}
		err := WrapError(OpList, inner)

		// Should return the same error since it's already an *Error
		Equals(t, err.Op, OpGet)
		Equals(t, err.Message, "original error")
	})
}
