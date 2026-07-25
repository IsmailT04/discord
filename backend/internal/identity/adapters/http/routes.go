package httpadapter

import (
	"net/http"

	"github.com/ismailtemuroglu/discord/internal/platform/middleware"
)

// APIPrefix is the versioned HTTP prefix for identity routes.
const APIPrefix = "/api/v1"

// Mount registers identity routes on mux under /api/v1.
//
// Public:  POST /api/v1/auth/register|login|refresh
// Protected (RequireAuth): POST /api/v1/auth/logout, GET /api/v1/users/me
func Mount(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST "+APIPrefix+"/auth/register", h.Register)
	mux.HandleFunc("POST "+APIPrefix+"/auth/login", h.Login)
	mux.HandleFunc("POST "+APIPrefix+"/auth/refresh", h.Refresh)

	withAuth := func(next http.Handler) http.Handler {
		return middleware.Chain(next, middleware.RequireAuth)
	}

	mux.Handle("POST "+APIPrefix+"/auth/logout", withAuth(http.HandlerFunc(h.Logout)))
	mux.Handle("GET "+APIPrefix+"/users/me", withAuth(http.HandlerFunc(h.Me)))
}
