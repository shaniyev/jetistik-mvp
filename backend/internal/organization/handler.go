package organization

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"jetistik/internal/platform/response"
)

// Handler holds organization HTTP handlers.
type Handler struct {
	svc *Service
}

// NewHandler creates a new organization handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// AdminRoutes registers admin-level organization routes.
func (h *Handler) AdminRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Get("/{id}/members", h.ListMembers)
	r.Post("/{id}/members", h.AddMember)
	r.Delete("/{id}/members/{uid}", h.RemoveMember)
	return r
}

// Create handles POST /api/v1/admin/organizations
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}
	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}
	org, err := h.svc.Create(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to create organization")
		return
	}
	response.JSON(w, http.StatusCreated, org)
}

// GetByID handles GET /api/v1/admin/organizations/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid organization id")
		return
	}
	org, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "organization not found")
		return
	}
	response.JSON(w, http.StatusOK, org)
}

// List handles GET /api/v1/admin/organizations
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	orgs, total, err := h.svc.List(r.Context(), page, perPage)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to list organizations")
		return
	}
	response.Paginated(w, orgs, response.Pagination{
		Page:    page,
		PerPage: perPage,
		Total:   int(total),
	})
}

// Update handles PATCH /api/v1/admin/organizations/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid organization id")
		return
	}
	var req UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}
	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}
	org, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to update organization")
		return
	}
	response.JSON(w, http.StatusOK, org)
}

// Delete handles DELETE /api/v1/admin/organizations/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid organization id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to delete organization")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListMembers handles GET /api/v1/admin/organizations/{id}/members
func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid organization id")
		return
	}
	members, err := h.svc.ListMembers(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to list members")
		return
	}
	response.JSON(w, http.StatusOK, members)
}

// AddMember handles POST /api/v1/admin/organizations/{id}/members
func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid organization id")
		return
	}
	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}
	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}
	member, err := h.svc.AddMember(r.Context(), id, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to add member")
		return
	}
	response.JSON(w, http.StatusCreated, member)
}

// RemoveMember handles DELETE /api/v1/admin/organizations/{id}/members/{uid}
func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid organization id")
		return
	}
	uid, err := strconv.ParseInt(chi.URLParam(r, "uid"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid user id")
		return
	}
	if err := h.svc.RemoveMember(r.Context(), id, uid); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to remove member")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
