package httpadapter

import (
	"errors"
	"net/http"

	"github.com/ismailtemuroglu/discord/internal/identity/application"
	"github.com/ismailtemuroglu/discord/internal/identity/domain"
	"github.com/ismailtemuroglu/discord/internal/platform/auth"
	"github.com/ismailtemuroglu/discord/internal/platform/config"
	"github.com/ismailtemuroglu/discord/internal/platform/httpx"
)

// Handler serves identity HTTP endpoints.
type Handler struct {
	svc  *application.Service
	opts auth.CookieOptions
}

// NewHandler constructs identity HTTP handlers.
func NewHandler(svc *application.Service, cfg *config.Config) *Handler {
	return &Handler{
		svc: svc,
		opts: auth.CookieOptions{
			Domain: cfg.Auth.CookieDomain,
			Secure: cfg.Auth.CookieSecure,
		},
	}
}

type registerRequest struct {
	Email       string `json:"email"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	User domain.PublicProfile `json:"user"`
}

// Register handles POST /api/v1/auth/register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := httpx.ReadJSON(w, r, &req, 1<<20); err != nil {
		writeErr(w, err)
		return
	}

	result, err := h.svc.Register(r.Context(), application.RegisterInput{
		Email:       req.Email,
		Username:    req.Username,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		writeDomainErr(w, err)
		return
	}

	if err := h.setSessionCookies(w, result); err != nil {
		_ = httpx.WriteError(w, httpx.ErrInternalServer)
		return
	}

	_ = httpx.WriteJSON(w, http.StatusCreated, authResponse{User: result.User})
}

// Login handles POST /api/v1/auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httpx.ReadJSON(w, r, &req, 1<<20); err != nil {
		writeErr(w, err)
		return
	}

	result, err := h.svc.Login(r.Context(), application.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeDomainErr(w, err)
		return
	}

	if err := h.setSessionCookies(w, result); err != nil {
		_ = httpx.WriteError(w, httpx.ErrInternalServer)
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, authResponse{User: result.User})
}

// Refresh handles POST /api/v1/auth/refresh.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken := auth.RefreshTokenFromRequest(r)
	result, err := h.svc.Refresh(r.Context(), refreshToken)
	if err != nil {
		// Clear cookies on invalid/reused refresh so the client re-authenticates.
		if errors.Is(err, domain.ErrInvalidRefresh) || errors.Is(err, domain.ErrRefreshReuse) {
			auth.ClearAuthCookies(w, h.opts)
		}
		writeDomainErr(w, err)
		return
	}

	if err := h.setSessionCookies(w, result); err != nil {
		_ = httpx.WriteError(w, httpx.ErrInternalServer)
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, authResponse{User: result.User})
}

// Logout handles POST /api/v1/auth/logout.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	access := auth.SessionIDFromContext(r.Context())
	if access == "" {
		access = auth.SessionTokenFromRequest(r)
	}
	refresh := auth.RefreshTokenFromRequest(r)

	if err := h.svc.Logout(r.Context(), application.LogoutInput{
		AccessToken:  access,
		RefreshToken: refresh,
	}); err != nil {
		_ = httpx.WriteError(w, httpx.ErrInternalServer)
		return
	}

	auth.ClearAuthCookies(w, h.opts)
	w.WriteHeader(http.StatusNoContent)
}

// Me handles GET /api/v1/users/me.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		_ = httpx.WriteError(w, httpx.ErrUnauthorized)
		return
	}

	profile, err := h.svc.Me(r.Context(), user.ID)
	if err != nil {
		writeDomainErr(w, err)
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, map[string]any{"user": profile})
}

func (h *Handler) setSessionCookies(w http.ResponseWriter, result *application.AuthResult) error {
	auth.SetAccessToken(w, result.AccessToken, result.AccessTTL, h.opts)
	auth.SetRefreshToken(w, result.RefreshToken, result.RefreshTTL, h.opts)

	csrf, err := auth.NewCSRFToken()
	if err != nil {
		return err
	}
	// CSRF outlives access so refresh still works after access expiry.
	auth.SetCSRFToken(w, csrf, result.RefreshTTL, h.opts)
	return nil
}

func writeErr(w http.ResponseWriter, err error) {
	var apiErr *httpx.APIError
	if errors.As(err, &apiErr) {
		_ = httpx.WriteError(w, apiErr)
		return
	}
	_ = httpx.WriteError(w, httpx.ErrInternalServer)
}

func writeDomainErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrInvalidUsername),
		errors.Is(err, domain.ErrInvalidPassword),
		errors.Is(err, domain.ErrInvalidDisplayName):
		_ = httpx.WriteError(w, &httpx.APIError{
			Code:    "VALIDATION_FAILED",
			Message: err.Error(),
			Status:  http.StatusBadRequest,
		})
	case errors.Is(err, domain.ErrEmailTaken):
		_ = httpx.WriteError(w, &httpx.APIError{
			Code:    "EMAIL_TAKEN",
			Message: "email already registered",
			Status:  http.StatusConflict,
		})
	case errors.Is(err, domain.ErrUsernameTaken):
		_ = httpx.WriteError(w, &httpx.APIError{
			Code:    "USERNAME_TAKEN",
			Message: "username already taken",
			Status:  http.StatusConflict,
		})
	case errors.Is(err, domain.ErrInvalidCredentials):
		_ = httpx.WriteError(w, &httpx.APIError{
			Code:    "INVALID_CREDENTIALS",
			Message: "invalid email or password",
			Status:  http.StatusUnauthorized,
		})
	case errors.Is(err, domain.ErrInvalidRefresh):
		_ = httpx.WriteError(w, &httpx.APIError{
			Code:    "INVALID_REFRESH",
			Message: "refresh token is missing or invalid",
			Status:  http.StatusUnauthorized,
		})
	case errors.Is(err, domain.ErrRefreshReuse):
		_ = httpx.WriteError(w, &httpx.APIError{
			Code:    "REFRESH_REUSE",
			Message: "refresh token reuse detected; session revoked",
			Status:  http.StatusUnauthorized,
		})
	case errors.Is(err, domain.ErrUserNotFound):
		_ = httpx.WriteError(w, httpx.ErrNotFound)
	case errors.Is(err, domain.ErrUnauthenticated):
		_ = httpx.WriteError(w, httpx.ErrUnauthorized)
	default:
		_ = httpx.WriteError(w, httpx.ErrInternalServer)
	}
}
