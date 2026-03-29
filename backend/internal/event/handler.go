package event

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"jetistik/internal/audit"
	"jetistik/internal/organization"
	"jetistik/internal/platform/middleware"
	"jetistik/internal/platform/response"
)

// Handler holds event HTTP handlers.
type Handler struct {
	svc      *Service
	orgSvc   *organization.Service
	auditSvc *audit.Service
}

// NewHandler creates a new event handler.
func NewHandler(svc *Service, orgSvc *organization.Service, auditSvc *audit.Service) *Handler {
	return &Handler{svc: svc, orgSvc: orgSvc, auditSvc: auditSvc}
}

// StaffRoutes registers staff-level event routes.
func (h *Handler) StaffRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

// Create handles POST /api/v1/staff/events
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())

	org, err := h.orgSvc.GetUserOrg(r.Context(), uc.UserID)
	if err != nil {
		response.Error(w, http.StatusForbidden, "NO_ORG", "you are not a member of any organization")
		return
	}

	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}
	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	event, err := h.svc.Create(r.Context(), org.ID, uc.UserID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to create event")
		return
	}

	h.auditSvc.Log(r.Context(), uc.UserID, audit.ActionEventCreate, "event", strconv.FormatInt(event.ID, 10), map[string]interface{}{"title": req.Title})
	response.JSON(w, http.StatusCreated, event)
}

// GetByID handles GET /api/v1/staff/events/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid event id")
		return
	}
	event, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "event not found")
		return
	}
	response.JSON(w, http.StatusOK, event)
}

// List handles GET /api/v1/staff/events
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())

	org, err := h.orgSvc.GetUserOrg(r.Context(), uc.UserID)
	if err != nil {
		response.Error(w, http.StatusForbidden, "NO_ORG", "you are not a member of any organization")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	events, total, err := h.svc.ListByOrg(r.Context(), org.ID, page, perPage)
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

// Update handles PATCH /api/v1/staff/events/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid event id")
		return
	}
	var req UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}
	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}
	event, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to update event")
		return
	}
	h.auditSvc.Log(r.Context(), uc.UserID, audit.ActionEventUpdate, "event", strconv.FormatInt(id, 10), nil)
	response.JSON(w, http.StatusOK, event)
}

// Delete handles DELETE /api/v1/staff/events/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid event id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to delete event")
		return
	}
	h.auditSvc.Log(r.Context(), uc.UserID, audit.ActionEventDelete, "event", strconv.FormatInt(id, 10), nil)
	w.WriteHeader(http.StatusNoContent)
}
