package batch

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"jetistik/internal/audit"
	"jetistik/internal/platform/middleware"
	"jetistik/internal/platform/response"
	tmpl "jetistik/internal/template"
)

// Handler holds batch HTTP handlers.
type Handler struct {
	svc      *Service
	tmplSvc  *tmpl.Service
	auditSvc *audit.Service
}

// NewHandler creates a new batch handler.
func NewHandler(svc *Service, tmplSvc *tmpl.Service, auditSvc *audit.Service) *Handler {
	return &Handler{svc: svc, tmplSvc: tmplSvc, auditSvc: auditSvc}
}

// Upload handles POST /api/v1/staff/events/{id}/batches
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())
	eventID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid event id")
		return
	}

	// Max 20 MB
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		response.Error(w, http.StatusBadRequest, "FILE_TOO_LARGE", "file must be under 20MB")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "MISSING_FILE", "file field is required")
		return
	}
	defer file.Close()

	// Validate extension
	lower := strings.ToLower(header.Filename)
	if !strings.HasSuffix(lower, ".csv") && !strings.HasSuffix(lower, ".xlsx") {
		response.Error(w, http.StatusBadRequest, "INVALID_FILE", "only CSV and XLSX files are allowed")
		return
	}

	// Get template tokens for default mapping
	var templateTokens []string
	tmplResp, err := h.tmplSvc.GetByEventID(r.Context(), eventID)
	if err == nil && tmplResp != nil {
		templateTokens = tmplResp.Tokens
	}

	result, err := h.svc.Upload(r.Context(), eventID, header.Filename, file, header.Size, templateTokens)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to upload batch: "+err.Error())
		return
	}

	h.auditSvc.Log(r.Context(), uc.UserID, audit.ActionBatchUpload, "batch", strconv.FormatInt(result.Batch.ID, 10), map[string]interface{}{
		"event_id": eventID,
		"filename": header.Filename,
		"rows":     result.PreviewRows,
	})
	response.JSON(w, http.StatusCreated, result)
}

// GetByID handles GET /api/v1/staff/batches/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid batch id")
		return
	}
	batchResp, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "batch not found")
		return
	}
	response.JSON(w, http.StatusOK, batchResp)
}

// UpdateMapping handles PATCH /api/v1/staff/batches/{id}/mapping
func (h *Handler) UpdateMapping(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid batch id")
		return
	}

	var req UpdateMappingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}
	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	batchResp, err := h.svc.UpdateMapping(r.Context(), id, req.Mapping)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to update mapping")
		return
	}

	h.auditSvc.Log(r.Context(), uc.UserID, audit.ActionBatchMapping, "batch", strconv.FormatInt(id, 10), nil)
	response.JSON(w, http.StatusOK, batchResp)
}

// Delete handles DELETE /api/v1/staff/batches/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid batch id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to delete batch")
		return
	}
	h.auditSvc.Log(r.Context(), uc.UserID, audit.ActionBatchDelete, "batch", strconv.FormatInt(id, 10), nil)
	w.WriteHeader(http.StatusNoContent)
}
