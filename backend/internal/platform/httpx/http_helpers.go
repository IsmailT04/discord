package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Default max payload sizes
const (
	DefaultMaxBodyBytes int64 = 1024 * 1024 // 1 MB (ideal for standard JSON APIs)
)

// APIError represents a standard, structured error response.
type APIError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Status  int               `json:"status"`
	Details map[string]string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Pre-defined standard platform errors
var (
	ErrInternalServer = &APIError{Code: "INTERNAL_ERROR", Message: "An unexpected error occurred", Status: http.StatusInternalServerError}
	ErrNotFound       = &APIError{Code: "NOT_FOUND", Message: "The requested resource was not found", Status: http.StatusNotFound}
	ErrUnauthorized   = &APIError{Code: "UNAUTHORIZED", Message: "Authentication is required", Status: http.StatusUnauthorized}
	ErrForbidden      = &APIError{Code: "FORBIDDEN", Message: "You do not have permission to access this resource", Status: http.StatusForbidden}
	ErrTooManyReqs    = &APIError{Code: "TOO_MANY_REQUESTS", Message: "Rate limit exceeded", Status: http.StatusTooManyRequests}
)

// WriteJSON encodes data as JSON with the given status code.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// WriteError sends a structured APIError response to the client.
func WriteError(w http.ResponseWriter, apiErr *APIError) error {
	return WriteJSON(w, apiErr.Status, apiErr)
}

// WriteValidationError sends a 400 Bad Request with field-level details.
func WriteValidationError(w http.ResponseWriter, fieldErrors map[string]string) error {
	return WriteError(w, &APIError{
		Code:    "VALIDATION_FAILED",
		Message: "The request payload failed validation checks.",
		Status:  http.StatusBadRequest,
		Details: fieldErrors,
	})
}

// RequireContentType enforces that incoming requests match expected MIME types.
func RequireContentType(r *http.Request, expectedType string) error {
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		return &APIError{Code: "MISSING_CONTENT_TYPE", Message: "Content-Type header is required", Status: http.StatusUnsupportedMediaType}
	}
	if !strings.HasPrefix(strings.ToLower(ct), strings.ToLower(expectedType)) {
		return &APIError{
			Code:    "UNSUPPORTED_MEDIA_TYPE",
			Message: fmt.Sprintf("Content-Type must be %s", expectedType),
			Status:  http.StatusUnsupportedMediaType,
		}
	}
	return nil
}

// ReadJSON securely decodes a JSON body with strict limits & checks:
// 1. Enforces Max Body Bytes (prevents OOM / DoS).
// 2. Rejects unknown/unmapped fields (disallowUnknownFields).
// 3. Ensures only ONE JSON value exists in the request body.
func ReadJSON(w http.ResponseWriter, r *http.Request, v interface{}, maxBytes int64) error {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBodyBytes
	}

	// 1. Enforce JSON Content-Type
	if err := RequireContentType(r, "application/json"); err != nil {
		return err
	}

	// 2. Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	// 3. Setup strict decoder
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // Catch client typos (e.g. "usrename" instead of "username")

	if err := dec.Decode(v); err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalErr *json.UnmarshalTypeError
		var maxBytesErr *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxErr):
			return &APIError{Code: "BAD_REQUEST", Message: fmt.Sprintf("Malformed JSON at position %d", syntaxErr.Offset), Status: http.StatusBadRequest}

		case errors.As(err, &unmarshalErr):
			return &APIError{Code: "BAD_REQUEST", Message: fmt.Sprintf("Invalid type for field %q", unmarshalErr.Field), Status: http.StatusBadRequest}

		case errors.As(err, &maxBytesErr):
			return &APIError{Code: "PAYLOAD_TOO_LARGE", Message: fmt.Sprintf("Request body exceeds maximum size of %d bytes", maxBytes), Status: http.StatusRequestEntityTooLarge}

		case errors.Is(err, io.EOF):
			return &APIError{Code: "BAD_REQUEST", Message: "Request body cannot be empty", Status: http.StatusBadRequest}

		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return &APIError{Code: "BAD_REQUEST", Message: fmt.Sprintf("Unknown field in request body: %s", fieldName), Status: http.StatusBadRequest}

		default:
			return &APIError{Code: "BAD_REQUEST", Message: "Failed to parse JSON body", Status: http.StatusBadRequest}
		}
	}

	// 4. Ensure there is only a single JSON object in the stream
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return &APIError{Code: "BAD_REQUEST", Message: "Request body must only contain a single JSON object", Status: http.StatusBadRequest}
	}

	return nil
}
