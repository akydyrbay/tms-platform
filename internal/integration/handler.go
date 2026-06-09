package integration

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

type attachRequest struct {
	Tracker    string  `json:"tracker"`
	ExternalID string  `json:"external_id"`
	URL        *string `json:"url"`
	Title      *string `json:"title"`
}

func (h *Handler) Attach(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	resultID, ok := parseID(w, r)
	if !ok {
		return
	}
	var req attachRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	b, err := h.svc.Attach(r.Context(), AttachInput{
		ResultID:   resultID,
		Tracker:    req.Tracker,
		ExternalID: req.ExternalID,
		URL:        req.URL,
		Title:      req.Title,
		CreatedBy:  userID,
	})
	if writeError(w, err) {
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, b)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	resultID, ok := parseID(w, r)
	if !ok {
		return
	}
	bugs, err := h.svc.ListByResult(r.Context(), resultID)
	if writeError(w, err) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, bugs)
}

func parseID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := chi.URLParam(r, "resultID")
	if _, err := uuid.Parse(id); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid result id")
		return "", false
	}
	return id, true
}

func writeError(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, ErrTrackerRequired):
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrResultNotFound):
		httputil.WriteError(w, http.StatusNotFound, err.Error())
	default:
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
	return true
}
