package auth

import (
	"context"
	"net/http"
	"strings"

	"tms-platform/pkg/httputil"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

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
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDContextKey).(string)
	return id, ok
}
