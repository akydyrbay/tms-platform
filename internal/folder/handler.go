package folder

import (
	"encoding/json"
	"errors"
	"net/http"

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
	ProjectID string  `json:"project_id"`
	ParentID  *string `json:"parent_id"`
	Name      string  `json:"name"`
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

	f, err := h.svc.Create(r.Context(), CreateInput{
		ProjectID: req.ProjectID,
		ParentID:  req.ParentID,
		Name:      req.Name,
		CreatedBy: userID,
	})
	switch {
	case errors.Is(err, ErrNameRequired),
		errors.Is(err, ErrProjectRequired),
		errors.Is(err, ErrParentDifferentProject):
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	case errors.Is(err, ErrParentNotFound):
		httputil.WriteError(w, http.StatusNotFound, err.Error())
		return
	case err != nil:
		httputil.WriteError(w, http.StatusInternalServerError, "could not create folder")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, f)
}

func (h *Handler) Tree(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	if _, err := uuid.Parse(projectID); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "valid project_id query parameter is required")
		return
	}

	tree, err := h.svc.Tree(r.Context(), projectID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "could not list folders")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, tree)
}
