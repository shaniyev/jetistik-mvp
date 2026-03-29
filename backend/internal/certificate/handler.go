package certificate

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"jetistik/internal/audit"
	"jetistik/internal/platform/middleware"
	"jetistik/internal/platform/response"
)

// Handler holds certificate HTTP handlers.
type Handler struct {
	svc      *Service
	auditSvc *audit.Service
}

// NewHandler creates a new certificate handler.
func NewHandler(svc *Service, auditSvc *audit.Service) *Handler {
	return &Handler{svc: svc, auditSvc: auditSvc}
}

// PublicRoutes registers public certificate routes.
func (h *Handler) PublicRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/verify/{code}", h.Verify)
	r.Get("/certificates/search", h.Search)
	r.Get("/certificates/{code}/download", h.PublicDownload)
	return r
}

// StaffCertificateRoutes registers staff certificate routes (nested under events).
func (h *Handler) StaffCertificateRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListByEvent)
	return r
}

// StaffCertificateItemRoutes registers staff routes for individual certificates.
func (h *Handler) StaffCertificateItemRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{id}/download", h.Download)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/revoke", h.Revoke)
	r.Post("/{id}/unrevoke", h.Unrevoke)
	return r
}

// Verify handles GET /api/v1/verify/{code}
func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		response.Error(w, http.StatusBadRequest, "MISSING_CODE", "verification code is required")
		return
	}
	result, err := h.svc.Verify(r.Context(), code)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "verification failed")
		return
	}
	response.JSON(w, http.StatusOK, result)
}

// Search handles GET /api/v1/certificates/search?iin=...
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	iin := r.URL.Query().Get("iin")
	if iin == "" || len(iin) != 12 {
		response.Error(w, http.StatusBadRequest, "INVALID_IIN", "IIN must be exactly 12 digits")
		return
	}
	results, err := h.svc.SearchByIIN(r.Context(), iin)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "search failed")
		return
	}
	response.JSON(w, http.StatusOK, results)
}

// PublicDownload handles GET /api/v1/certificates/{code}/download
func (h *Handler) PublicDownload(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	verifyResult, err := h.svc.Verify(r.Context(), code)
	if err != nil || !verifyResult.Valid {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "certificate not found")
		return
	}
	url, err := h.svc.DownloadURLByCode(r.Context(), code)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "download failed")
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// ListByEvent handles GET /api/v1/staff/events/{id}/certificates
func (h *Handler) ListByEvent(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid event id")
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

	certs, total, err := h.svc.ListByEvent(r.Context(), eventID, page, perPage)
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

// Download handles GET /api/v1/staff/certificates/{id}/download
func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid certificate id")
		return
	}
	url, err := h.svc.DownloadURL(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "download failed")
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Update handles PATCH /api/v1/staff/certificates/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid certificate id")
		return
	}
	var req UpdateCertificateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}
	if req.Status == nil {
		response.Error(w, http.StatusBadRequest, "MISSING_STATUS", "status is required")
		return
	}
	cert, err := h.svc.UpdateStatus(r.Context(), id, *req.Status)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to update certificate")
		return
	}
	response.JSON(w, http.StatusOK, cert)
}

// Delete handles DELETE /api/v1/staff/certificates/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid certificate id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to delete certificate")
		return
	}
	h.auditSvc.Log(r.Context(), uc.UserID, audit.ActionCertificateDelete, "certificate", strconv.FormatInt(id, 10), nil)
	w.WriteHeader(http.StatusNoContent)
}

// Revoke handles POST /api/v1/staff/certificates/{id}/revoke
func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid certificate id")
		return
	}
	var req RevokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}
	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}
	cert, err := h.svc.Revoke(r.Context(), id, req.Reason)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to revoke certificate")
		return
	}
	h.auditSvc.Log(r.Context(), uc.UserID, audit.ActionCertificateRevoke, "certificate", strconv.FormatInt(id, 10), map[string]interface{}{"reason": req.Reason})
	response.JSON(w, http.StatusOK, cert)
}

// Unrevoke handles POST /api/v1/staff/certificates/{id}/unrevoke
func (h *Handler) Unrevoke(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid certificate id")
		return
	}
	cert, err := h.svc.Unrevoke(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to unrevoke certificate")
		return
	}
	h.auditSvc.Log(r.Context(), uc.UserID, audit.ActionCertificateUnrevoke, "certificate", strconv.FormatInt(id, 10), nil)
	response.JSON(w, http.StatusOK, cert)
}
