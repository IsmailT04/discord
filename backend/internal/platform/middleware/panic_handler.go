package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/ismailtemuroglu/discord/internal/platform/httpx"
	"github.com/ismailtemuroglu/discord/internal/platform/logger"
)

// Recovery catches panics, logs them with zap, and returns a structured 500.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.FromContext(r.Context()).Errorw("panic recovered",
					"panic", rec,
					"path", r.URL.Path,
					"method", r.Method,
					"stack", string(debug.Stack()),
				)
				_ = httpx.WriteError(w, httpx.ErrInternalServer)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
