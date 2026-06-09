package testrun

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
	SuiteID string `json:"suite_id"`
	Name    string `json:"name"`
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
	if _, err := uuid.Parse(req.SuiteID); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid suite_id")
		return
	}
	run, err := h.svc.Create(r.Context(), CreateInput{
		SuiteID:   req.SuiteID,
		Name:      req.Name,
		CreatedBy: userID,
	})
	if writeError(w, err) {
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, run)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	if _, err := uuid.Parse(projectID); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "valid project_id query parameter is required")
		return
	}
	runs, err := h.svc.ListByProject(r.Context(), projectID)
	if writeError(w, err) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, runs)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	run, err := h.svc.Get(r.Context(), id)
	if writeError(w, err) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, run)
}

func (h *Handler) Results(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	results, err := h.svc.Results(r.Context(), id)
	if writeError(w, err) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, results)
}

type markRequest struct {
	Status  string  `json:"status"`
	Comment *string `json:"comment"`
}

func (h *Handler) Mark(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	runID, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	caseID, ok := parseID(w, r, "caseID")
	if !ok {
		return
	}
	var req markRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	res, err := h.svc.Mark(r.Context(), runID, caseID, MarkInput{
		Status:     req.Status,
		Comment:    req.Comment,
		ExecutedBy: userID,
	})
	if writeError(w, err) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, res)
}

type statusRequest struct {
	Status string `json:"status"`
}

func (h *Handler) SetStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	var req statusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	run, err := h.svc.SetStatus(r.Context(), id, req.Status)
	if writeError(w, err) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, run)
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	stats, err := h.svc.Stats(r.Context(), id)
	if writeError(w, err) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, stats)
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
	case errors.Is(err, ErrNameRequired), errors.Is(err, ErrInvalidStatus):
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrSuiteNotFound), errors.Is(err, ErrResultNotFound):
		httputil.WriteError(w, http.StatusNotFound, err.Error())
	default:
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
	return true
}
