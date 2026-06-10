package auth

import (
	"context"
	"net/http"
	"strings"

	"tms-platform/pkg/httputil"
)

type contextKey string

const (
	userIDContextKey contextKey = "user_id"
	roleContextKey   contextKey = "role"
)

const (
	RoleViewer    = "viewer"
	RoleDeveloper = "developer"
	RoleQA        = "qa"
	RoleQALead    = "qa_lead"
	RoleAdmin     = "admin"
)

var WriteRoles = []string{RoleQA, RoleQALead, RoleAdmin}

var ManageRoles = []string{RoleQALead, RoleAdmin}

var AllRoles = []string{RoleViewer, RoleDeveloper, RoleQA, RoleQALead, RoleAdmin}

func IsValidRole(role string) bool {
	for _, r := range AllRoles {
		if r == role {
			return true
		}
	}
	return false
}

func (m *JWTManager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || strings.TrimSpace(token) == "" {
			httputil.WriteError(w, http.StatusUnauthorized, "missing or malformed authorization header")
			return
		}

		claims, err := m.Validate(token)
		if err != nil {
			httputil.WriteError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID)
		ctx = context.WithValue(ctx, roleContextKey, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireRoles(allowed ...string) func(http.Handler) http.Handler {
	set := make(map[string]bool, len(allowed))
	for _, r := range allowed {
		set[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := RoleFromContext(r.Context())
			if !set[role] {
				httputil.WriteError(w, http.StatusForbidden, "insufficient permissions for this action")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDContextKey).(string)
	return id, ok
}

func RoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(roleContextKey).(string)
	return role, ok
}
