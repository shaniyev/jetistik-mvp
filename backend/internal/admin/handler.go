package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"jetistik/internal/platform/response"
)

// Handler holds admin HTTP handlers.
type Handler struct {
	svc *Service
}

// NewHandler creates a new admin handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes registers admin-specific routes.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/stats", h.Stats)
	r.Get("/events", h.ListEvents)
	r.Get("/certificates", h.ListCertificates)
	r.Get("/users", h.ListUsers)
	r.Get("/users/{id}", h.GetUser)
	r.Patch("/users/{id}", h.UpdateUser)
	r.Delete("/users/{id}", h.DeleteUser)
	return r
}

// Stats handles GET /api/v1/admin/stats
func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.Stats(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to get stats")
		return
	}
	response.JSON(w, http.StatusOK, stats)
}

// ListEvents handles GET /api/v1/admin/events
func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	events, total, err := h.svc.ListEvents(r.Context(), page, perPage)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to list events")
		return
	}
	response.Paginated(w, events, response.Pagination{
		Page:    page,
		PerPage: perPage,
		Total:   int(total),
	})
}

// ListCertificates handles GET /api/v1/admin/certificates
func (h *Handler) ListCertificates(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	certs, total, err := h.svc.ListCertificates(r.Context(), page, perPage)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to list certificates")
		return
	}
	response.Paginated(w, certs, response.Pagination{
		Page:    page,
		PerPage: perPage,
		Total:   int(total),
	})
}

// ListUsers handles GET /api/v1/admin/users
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	users, total, err := h.svc.ListUsers(r.Context(), page, perPage)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to list users")
		return
	}
	response.Paginated(w, users, response.Pagination{
		Page:    page,
		PerPage: perPage,
		Total:   int(total),
	})
}

// GetUser handles GET /api/v1/admin/users/{id}
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid user id")
		return
	}
	user, err := h.svc.GetUser(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "user not found")
		return
	}
	response.JSON(w, http.StatusOK, user)
}

// UpdateUser handles PATCH /api/v1/admin/users/{id}
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid user id")
		return
	}
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}
	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}
	user, err := h.svc.UpdateUser(r.Context(), id, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to update user")
		return
	}
	response.JSON(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /api/v1/admin/users/{id}
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid user id")
		return
	}
	if err := h.svc.DeleteUser(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to delete user")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
