package template

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"jetistik/internal/audit"
	"jetistik/internal/platform/middleware"
	"jetistik/internal/platform/response"
)

// Handler holds template HTTP handlers.
type Handler struct {
	svc      *Service
	auditSvc *audit.Service
}

// NewHandler creates a new template handler.
func NewHandler(svc *Service, auditSvc *audit.Service) *Handler {
	return &Handler{svc: svc, auditSvc: auditSvc}
}

// Upload handles POST /api/v1/staff/events/{id}/template
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())
	eventID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid event id")
		return
	}

	// Max 50 MB
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		response.Error(w, http.StatusBadRequest, "FILE_TOO_LARGE", "file must be under 50MB")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "MISSING_FILE", "file field is required")
		return
	}
	defer file.Close()

	// Validate file extension
	ext := storageExt(header.Filename)
	if ext != ".pptx" {
		response.Error(w, http.StatusBadRequest, "INVALID_FILE", "only .pptx files are allowed")
		return
	}

	tmpl, err := h.svc.Upload(r.Context(), eventID, header.Filename, file, header.Size)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to upload template")
		return
	}

	h.auditSvc.Log(r.Context(), uc.UserID, audit.ActionTemplateUpload, "event", strconv.FormatInt(eventID, 10), map[string]interface{}{
		"filename": header.Filename,
		"tokens":   tmpl.Tokens,
	})
	response.JSON(w, http.StatusCreated, tmpl)
}

// Delete handles DELETE /api/v1/staff/events/{id}/template
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())
	eventID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid event id")
		return
	}
	if err := h.svc.Delete(r.Context(), eventID); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to delete template")
		return
	}
	h.auditSvc.Log(r.Context(), uc.UserID, audit.ActionTemplateDelete, "event", strconv.FormatInt(eventID, 10), nil)
	w.WriteHeader(http.StatusNoContent)
}

// GetByEvent handles GET /api/v1/staff/events/{id}/template
func (h *Handler) GetByEvent(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid event id")
		return
	}
	tmpl, err := h.svc.GetByEventID(r.Context(), eventID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "no template found for this event")
		return
	}
	response.JSON(w, http.StatusOK, tmpl)
}

// storageExt extracts file extension (simplified inline to avoid import cycle).
func storageExt(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i:]
		}
	}
	return ""
}
