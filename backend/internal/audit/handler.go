package audit

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"jetistik/internal/platform/response"
)

// Handler holds audit HTTP handlers.
type Handler struct {
	svc *Service
}

// NewHandler creates a new audit handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes registers audit log routes.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	return r
}

// List handles GET /api/v1/staff/audit-log or /api/v1/admin/audit-log
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	action := r.URL.Query().Get("action")
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	logs, total, err := h.svc.List(r.Context(), page, perPage, action)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to list audit logs")
		return
	}
	response.Paginated(w, logs, response.Pagination{
		Page:    page,
		PerPage: perPage,
		Total:   int(total),
	})
}
