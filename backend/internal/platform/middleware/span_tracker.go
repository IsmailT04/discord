package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// SpanTracker wraps the handler with OpenTelemetry HTTP instrumentation.
// It creates a server span, propagates context, and records request attributes.
func SpanTracker(next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, "http")
}
