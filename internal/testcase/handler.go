package testcase

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"tms-platform/internal/auth"
	"tms-platform/internal/model"
	"tms-platform/pkg/httputil"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type stepRequest struct {
	Action         string  `json:"action"`
	ExpectedResult *string `json:"expected_result"`
}

type contentRequest struct {
	Title          string        `json:"title"`
	Description    *string       `json:"description"`
	Preconditions  *string       `json:"preconditions"`
	ExpectedResult *string       `json:"expected_result"`
	Module         *string       `json:"module"`
	Component      *string       `json:"component"`
	Priority       string        `json:"priority"`
	Type           string        `json:"type"`
	Status         string        `json:"status"`
	Steps          []stepRequest `json:"steps"`
}

type createRequest struct {
	ProjectID string  `json:"project_id"`
	FolderID  *string `json:"folder_id"`
	contentRequest
}

func (c contentRequest) toContent() Content {
	steps := make([]model.TestCaseStep, len(c.Steps))
	for i, s := range c.Steps {
		steps[i] = model.TestCaseStep{Action: s.Action, ExpectedResult: s.ExpectedResult}
	}
	return Content{
		Title:          c.Title,
		Description:    c.Description,
		Preconditions:  c.Preconditions,
		ExpectedResult: c.ExpectedResult,
		Module:         c.Module,
		Component:      c.Component,
		Priority:       c.Priority,
		Type:           c.Type,
		Status:         c.Status,
		Steps:          steps,
	}
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

	v, err := h.svc.Create(r.Context(), CreateInput{
		ProjectID: req.ProjectID,
		FolderID:  req.FolderID,
		CreatedBy: userID,
		Content:   req.contentRequest.toContent(),
	})
	if writeServiceError(w, err) {
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, v)
}

func (h *Handler) Edit(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var req contentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	v, err := h.svc.Edit(r.Context(), id, userID, req.toContent())
	if writeServiceError(w, err) {
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, v)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	if _, err := uuid.Parse(projectID); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "valid project_id query parameter is required")
		return
	}
	cases, err := h.svc.ListByProject(r.Context(), projectID)
	if writeServiceError(w, err) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, cases)
}

func (h *Handler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	v, err := h.svc.GetCurrent(r.Context(), id)
	if writeServiceError(w, err) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, v)
}

func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	versions, err := h.svc.History(r.Context(), id)
	if writeServiceError(w, err) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, versions)
}

func (h *Handler) GetVersion(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	number, err := strconv.Atoi(chi.URLParam(r, "number"))
	if err != nil || number < 1 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid version number")
		return
	}
	v, err := h.svc.GetVersion(r.Context(), id, number)
	if writeServiceError(w, err) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, v)
}

func parseID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid test case id")
		return "", false
	}
	return id, true
}

// writeServiceError maps service errors to HTTP responses. It returns true when
// a response was written (i.e. err != nil).
func writeServiceError(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, ErrTitleRequired), errors.Is(err, ErrProjectRequired):
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrVersionNotFound):
		httputil.WriteError(w, http.StatusNotFound, err.Error())
	default:
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
	return true
}
