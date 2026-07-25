package middleware

import (
	"net/http"
	"time"

	"github.com/ismailtemuroglu/discord/internal/platform/logger"
	"go.opentelemetry.io/otel/trace"
)

// AccessLog logs a structured access record after each request completes.
// Fields: method, path, status, duration_ms, request_id, trace_id.
//
// Place inside SpanTracker so trace_id is available when OTel is active.
func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := newStatusResponseWriter(w)

		next.ServeHTTP(sw, r)

		requestID, _ := r.Context().Value(RequestIDKey).(string)

		traceID := ""
		if spanCtx := trace.SpanContextFromContext(r.Context()); spanCtx.IsValid() {
			traceID = spanCtx.TraceID().String()
		}

		logger.FromContext(r.Context()).Infow("http_request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", requestID,
			"trace_id", traceID,
		)
	})
}
