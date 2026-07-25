package middleware

import (
	"net/http"

	"github.com/ismailtemuroglu/discord/internal/platform/auth"
	"github.com/ismailtemuroglu/discord/internal/platform/httpx"
	"github.com/ismailtemuroglu/discord/internal/platform/logger"
)

// LoadSession attempts to load the session from the access_token or session_id
// cookie and populate the request context with the authenticated user.
//
// Missing, invalid, or expired sessions are ignored — the request continues
// unauthenticated. Use on public/optional-auth route groups.
//
// Pair with RequireAuth on protected route groups.
func LoadSession(sessions auth.SessionLookup) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := auth.SessionTokenFromRequest(r)
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			user, err := sessions.Lookup(r.Context(), token)
			if err != nil {
				logger.FromContext(r.Context()).Warnw("session lookup failed",
					"error", err,
				)
				next.ServeHTTP(w, r)
				return
			}
			if user == nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := auth.WithUser(r.Context(), user)
			ctx = auth.WithSessionID(ctx, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth aborts with 401 Unauthorized when no user is in the request
// context. Run after LoadSession on authenticated route groups.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth.UserFromContext(r.Context()) == nil {
			_ = httpx.WriteError(w, httpx.ErrUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
