package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ContextKey string

const (
	RequestIDKey    ContextKey = "request_id"
	RequestIDHeader string     = "X-Request-ID"
)

// RequestID ensures every request has an ID in context and echoes it on the response.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		w.Header().Set(RequestIDHeader, requestID)

		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
