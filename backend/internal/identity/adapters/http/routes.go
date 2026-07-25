package httpadapter

import (
	"net/http"

	"github.com/ismailtemuroglu/discord/internal/platform/middleware"
)

// Mount registers identity routes on mux.
//
// Public:  POST /auth/register, POST /auth/login, POST /auth/refresh
// Protected (RequireAuth): POST /auth/logout, GET /users/me
func Mount(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /auth/register", h.Register)
	mux.HandleFunc("POST /auth/login", h.Login)
	mux.HandleFunc("POST /auth/refresh", h.Refresh)

	withAuth := func(next http.Handler) http.Handler {
		return middleware.Chain(next, middleware.RequireAuth)
	}

	mux.Handle("POST /auth/logout", withAuth(http.HandlerFunc(h.Logout)))
	mux.Handle("GET /users/me", withAuth(http.HandlerFunc(h.Me)))
}
