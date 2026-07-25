package middleware

import "net/http"

// Chain wraps h with middlewares. The first middleware is outermost
// (sees the request first), matching the documented stack order.
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
