package project

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

	p, err := h.svc.Create(r.Context(), CreateInput{
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   userID,
	})
	switch {
	case errors.Is(err, ErrNameRequired):
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	case err != nil:
		httputil.WriteError(w, http.StatusInternalServerError, "could not create project")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, p)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.svc.List(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "could not list projects")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, projects)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	p, err := h.svc.Get(r.Context(), id)
	switch {
	case errors.Is(err, ErrNotFound):
		httputil.WriteError(w, http.StatusNotFound, "project not found")
		return
	case err != nil:
		httputil.WriteError(w, http.StatusInternalServerError, "could not get project")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, p)
}
