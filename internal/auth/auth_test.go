package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPasswordHashAndCheck(t *testing.T) {
	hash, err := HashPassword("s3cret-pw")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "" || hash == "s3cret-pw" {
		t.Fatalf("hash should be non-empty and not the plaintext, got %q", hash)
	}
	if !CheckPassword("s3cret-pw", hash) {
		t.Error("CheckPassword should accept the correct password")
	}
	if CheckPassword("wrong-pw", hash) {
		t.Error("CheckPassword should reject a wrong password")
	}
}

func TestJWTGenerateValidateRoundTrip(t *testing.T) {
	m := NewJWTManager([]byte("test-secret"), time.Hour)
	id := uuid.New()

	token, err := m.Generate(id, "qa@test.com", RoleQA)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	claims, err := m.Validate(token)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if claims.UserID != id.String() {
		t.Errorf("user_id: want %s, got %s", id.String(), claims.UserID)
	}
	if claims.Email != "qa@test.com" || claims.Role != RoleQA {
		t.Errorf("claims mismatch: %+v", claims)
	}
}

func TestJWTRejectsTamperedAndExpired(t *testing.T) {
	m := NewJWTManager([]byte("secret-a"), time.Hour)
	token, _ := m.Generate(uuid.New(), "a@b.com", RoleViewer)

	// Different secret must reject the signature.
	other := NewJWTManager([]byte("secret-b"), time.Hour)
	if _, err := other.Validate(token); err == nil {
		t.Error("validate should reject a token signed with a different secret")
	}

	// Already-expired token must be rejected.
	expiredMgr := NewJWTManager([]byte("secret-a"), -time.Minute)
	expired, _ := expiredMgr.Generate(uuid.New(), "a@b.com", RoleViewer)
	if _, err := m.Validate(expired); err == nil {
		t.Error("validate should reject an expired token")
	}
}

func TestIsValidRole(t *testing.T) {
	for _, r := range AllRoles {
		if !IsValidRole(r) {
			t.Errorf("%q should be valid", r)
		}
	}
	if IsValidRole("superuser") {
		t.Error("unknown role should be invalid")
	}
}

func TestMiddlewareRejectsMissingHeader(t *testing.T) {
	m := NewJWTManager([]byte("s"), time.Hour)
	h := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not run without a token")
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rr.Code)
	}
}

func TestMiddlewareStoresUserAndRole(t *testing.T) {
	m := NewJWTManager([]byte("s"), time.Hour)
	id := uuid.New()
	token, _ := m.Generate(id, "x@y.com", RoleAdmin)

	var gotID, gotRole string
	h := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID, _ = UserIDFromContext(r.Context())
		gotRole, _ = RoleFromContext(r.Context())
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if gotID != id.String() {
		t.Errorf("context user id: want %s, got %s", id.String(), gotID)
	}
	if gotRole != RoleAdmin {
		t.Errorf("context role: want %s, got %s", RoleAdmin, gotRole)
	}
}

func TestRequireRoles(t *testing.T) {
	guard := RequireRoles(WriteRoles...) // qa, qa_lead, admin
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cases := map[string]int{
		RoleAdmin:     http.StatusOK,
		RoleQALead:    http.StatusOK,
		RoleQA:        http.StatusOK,
		RoleViewer:    http.StatusForbidden,
		RoleDeveloper: http.StatusForbidden,
		"":            http.StatusForbidden,
	}
	for role, want := range cases {
		rr := httptest.NewRecorder()
		ctx := context.WithValue(context.Background(), roleContextKey, role)
		req := httptest.NewRequest("POST", "/", nil).WithContext(ctx)
		guard(next).ServeHTTP(rr, req)
		if rr.Code != want {
			t.Errorf("role %q: want %d, got %d", role, want, rr.Code)
		}
	}
}
