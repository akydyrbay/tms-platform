package suite

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"tms-platform/internal/auth"
	"tms-platform/pkg/httputil"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type createRequest struct {
	ProjectID   string  `json:"project_id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	s, err := h.svc.Create(r.Context(), CreateInput{
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   userID,
	})
	if writeError(w, err) {
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, s)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	if _, err := uuid.Parse(projectID); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "valid project_id query parameter is required")
		return
	}
	suites, err := h.svc.ListByProject(r.Context(), projectID)
	if writeError(w, err) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, suites)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	s, err := h.svc.Get(r.Context(), id)
	if writeError(w, err) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, s)
}

type addCaseRequest struct {
	TestCaseID string `json:"test_case_id"`
}

func (h *Handler) AddCase(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	var req addCaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if _, err := uuid.Parse(req.TestCaseID); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid test_case_id")
		return
	}
	s, err := h.svc.AddCase(r.Context(), id, req.TestCaseID)
	if writeError(w, err) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, s)
}

func (h *Handler) RemoveCase(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	caseID, ok := parseID(w, r, "caseID")
	if !ok {
		return
	}
	s, err := h.svc.RemoveCase(r.Context(), id, caseID)
	if writeError(w, err) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, s)
}

func parseID(w http.ResponseWriter, r *http.Request, param string) (string, bool) {
	id := chi.URLParam(r, param)
	if _, err := uuid.Parse(id); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid id")
		return "", false
	}
	return id, true
}

func writeError(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, ErrNameRequired), errors.Is(err, ErrProjectRequired):
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrCaseNotFound):
		httputil.WriteError(w, http.StatusNotFound, err.Error())
	default:
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
	return true
}
