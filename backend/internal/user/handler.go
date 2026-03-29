package user

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"jetistik/internal/platform/middleware"
	"jetistik/internal/platform/response"
)

// Handler holds user HTTP handlers.
type Handler struct {
	svc *Service
}

// NewHandler creates a new user handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ProfileRoutes registers profile-related routes.
func (h *Handler) ProfileRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetProfile)
	r.Patch("/", h.UpdateProfile)
	return r
}

// TeacherStudentRoutes registers teacher-student routes.
func (h *Handler) TeacherStudentRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireRole("teacher"))
	r.Get("/", h.ListStudents)
	r.Post("/", h.AddStudent)
	r.Delete("/{iin}", h.RemoveStudent)
	return r
}

// GetProfile handles GET /api/v1/profile
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())

	profile, err := h.svc.GetProfile(r.Context(), uc.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to get profile")
		return
	}

	response.JSON(w, http.StatusOK, profile)
}

// UpdateProfile handles PATCH /api/v1/profile
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	profile, err := h.svc.UpdateProfile(r.Context(), uc.UserID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to update profile")
		return
	}

	response.JSON(w, http.StatusOK, profile)
}

// ListStudents handles GET /api/v1/teacher/students
func (h *Handler) ListStudents(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())

	students, err := h.svc.ListStudents(r.Context(), uc.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to list students")
		return
	}

	response.JSON(w, http.StatusOK, students)
}

// AddStudent handles POST /api/v1/teacher/students
func (h *Handler) AddStudent(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())

	var req AddStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	ts, err := h.svc.AddStudent(r.Context(), uc.UserID, req.StudentIIN)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to add student")
		return
	}

	response.JSON(w, http.StatusCreated, ts)
}

// RemoveStudent handles DELETE /api/v1/teacher/students/{iin}
func (h *Handler) RemoveStudent(w http.ResponseWriter, r *http.Request) {
	uc := middleware.MustGetUser(r.Context())
	iin := chi.URLParam(r, "iin")

	if err := h.svc.RemoveStudent(r.Context(), uc.UserID, iin); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", "failed to remove student")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
