package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"jetistik/internal/platform/response"
)

// Handler holds auth HTTP handlers.
type Handler struct {
	svc          *Service
	refreshTTL   time.Duration
	secureCookie bool
}

// NewHandler creates a new auth handler.
func NewHandler(svc *Service, refreshTTL time.Duration, secureCookie bool) *Handler {
	return &Handler{
		svc:          svc,
		refreshTTL:   refreshTTL,
		secureCookie: secureCookie,
	}
}

// Routes registers auth routes on the given router.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/login", h.Login)
	r.Post("/register", h.Register)
	r.Post("/register/org", h.RegisterOrg)
	r.Post("/refresh", h.Refresh)
	r.Post("/logout", h.Logout)
	return r
}

// Login handles POST /api/v1/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	resp, rawRefresh, err := h.svc.Login(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			response.Error(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid username or password")
		case errors.Is(err, ErrUserNotActive):
			response.Error(w, http.StatusForbidden, "USER_INACTIVE", "user account is not active")
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL", "login failed")
		}
		return
	}

	h.setRefreshCookie(w, rawRefresh)
	response.JSON(w, http.StatusOK, resp)
}

// Register handles POST /api/v1/auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	resp, rawRefresh, err := h.svc.Register(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrUsernameExists):
			response.Error(w, http.StatusConflict, "USERNAME_EXISTS", "username already taken")
		case errors.Is(err, ErrEmailExists):
			response.Error(w, http.StatusConflict, "EMAIL_EXISTS", "email already taken")
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL", "registration failed")
		}
		return
	}

	h.setRefreshCookie(w, rawRefresh)
	response.JSON(w, http.StatusCreated, resp)
}

// RegisterOrg handles POST /api/v1/auth/register/org
func (h *Handler) RegisterOrg(w http.ResponseWriter, r *http.Request) {
	var req RegisterOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	resp, rawRefresh, err := h.svc.RegisterOrg(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrUsernameExists):
			response.Error(w, http.StatusConflict, "USERNAME_EXISTS", "username already taken")
		case errors.Is(err, ErrEmailExists):
			response.Error(w, http.StatusConflict, "EMAIL_EXISTS", "email already taken")
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL", "registration failed")
		}
		return
	}

	h.setRefreshCookie(w, rawRefresh)
	response.JSON(w, http.StatusCreated, resp)
}

// Refresh handles POST /api/v1/auth/refresh
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		response.Error(w, http.StatusUnauthorized, "NO_REFRESH_TOKEN", "refresh token not found")
		return
	}

	resp, rawRefresh, err := h.svc.Refresh(r.Context(), cookie.Value)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRefreshToken):
			h.clearRefreshCookie(w)
			response.Error(w, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "invalid or expired refresh token")
		case errors.Is(err, ErrUserNotActive):
			h.clearRefreshCookie(w)
			response.Error(w, http.StatusForbidden, "USER_INACTIVE", "user account is not active")
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL", "token refresh failed")
		}
		return
	}

	h.setRefreshCookie(w, rawRefresh)
	response.JSON(w, http.StatusOK, resp)
}

// Logout handles POST /api/v1/auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err == nil && cookie.Value != "" {
		_ = h.svc.Logout(r.Context(), cookie.Value)
	}

	h.clearRefreshCookie(w)
	response.JSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *Handler) setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(h.refreshTTL.Seconds()),
	})
}

func (h *Handler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}
