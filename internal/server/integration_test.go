package server

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"tms-platform/internal/auth"
	"tms-platform/internal/dto"
	"tms-platform/internal/folder"
	"tms-platform/internal/integration"
	"tms-platform/internal/model"
	"tms-platform/internal/project"
	"tms-platform/internal/suite"
	"tms-platform/internal/testcase"
	"tms-platform/internal/testrun"
)

func TestWorkflow(t *testing.T) {
	ctx := context.Background()

	migration, err := filepath.Abs("../../migrations/001_init.sql")
	if err != nil {
		t.Fatalf("abs migration path: %v", err)
	}

	pg, err := postgres.Run(ctx, "postgres:latest",
		postgres.WithInitScripts(migration),
		postgres.WithDatabase("tms"),
		postgres.WithUsername("tms"),
		postgres.WithPassword("tms"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	defer func() { _ = pg.Terminate(ctx) }()

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("conn string: %v", err)
	}
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	dbx := sqlx.NewDb(db, "pgx")

	ts := httptest.NewServer(buildTestServer(dbx).RegisterRoutes())
	defer ts.Close()

	c := &client{t: t, base: ts.URL}

	// --- users with different roles ---
	c.register("Admin", "admin@test.com", "pw123456", auth.RoleAdmin)
	adminTok := c.login("admin@test.com", "pw123456")

	c.register("Viewer", "viewer@test.com", "pw123456", auth.RoleViewer)
	viewerTok := c.login("viewer@test.com", "pw123456")

	// --- RBAC: viewer cannot create, admin can ---
	if code, _ := c.do("POST", "/api/v1/projects", viewerTok, map[string]any{"name": "X"}); code != http.StatusForbidden {
		t.Fatalf("viewer create project: want 403, got %d", code)
	}
	var proj model.Project
	c.json("POST", "/api/v1/projects", adminTok, map[string]any{"name": "E-Commerce"}, http.StatusCreated, &proj)

	// viewer CAN read (read-only access)
	if code, _ := c.do("GET", "/api/v1/projects", viewerTok, nil); code != http.StatusOK {
		t.Fatalf("viewer list projects: want 200, got %d", code)
	}

	// --- create a test case (v1) ---
	var v1 model.TestCaseVersion
	c.json("POST", "/api/v1/test-cases", adminTok, map[string]any{
		"project_id": proj.ID,
		"title":      "Login works",
		"priority":   "high",
		"steps": []map[string]any{
			{"action": "Open login page", "expected_result": "form shown"},
			{"action": "Submit valid creds", "expected_result": "dashboard"},
		},
	}, http.StatusCreated, &v1)
	if v1.VersionNumber != 1 {
		t.Fatalf("first version should be 1, got %d", v1.VersionNumber)
	}

	// --- suite + add case ---
	var st model.Suite
	c.json("POST", "/api/v1/suites", adminTok, map[string]any{
		"project_id": proj.ID, "name": "Regression",
	}, http.StatusCreated, &st)
	c.json("POST", "/api/v1/suites/"+st.ID+"/cases", adminTok, map[string]any{
		"test_case_id": v1.TestCaseID,
	}, http.StatusOK, &st)

	// --- create run: snapshots the CURRENT version (v1) ---
	var run model.TestRun
	c.json("POST", "/api/v1/test-runs", adminTok, map[string]any{
		"suite_id": st.ID, "name": "Release 1.0",
	}, http.StatusCreated, &run)

	frozen := c.results(adminTok, run.ID)
	if len(frozen) != 1 {
		t.Fatalf("run should have 1 result, got %d", len(frozen))
	}
	if frozen[0].TestCaseVersionID != v1.ID {
		t.Fatalf("run should be frozen at v1 (%s), got %s", v1.ID, frozen[0].TestCaseVersionID)
	}

	// --- edit the case → creates v2 ---
	var v2 model.TestCaseVersion
	c.json("PUT", "/api/v1/test-cases/"+v1.TestCaseID, adminTok, map[string]any{
		"title":    "Login works (updated)",
		"priority": "high",
		"steps":    []map[string]any{{"action": "New step", "expected_result": "ok"}},
	}, http.StatusCreated, &v2)
	if v2.VersionNumber != 2 || v2.ID == v1.ID {
		t.Fatalf("edit should create v2 with new id, got version=%d id=%s", v2.VersionNumber, v2.ID)
	}

	// --- KEY INVARIANT: run still points at v1, not v2 ---
	afterEdit := c.results(adminTok, run.ID)
	if afterEdit[0].TestCaseVersionID != v1.ID {
		t.Fatalf("FREEZE BROKEN: run moved to %s, expected frozen v1 %s", afterEdit[0].TestCaseVersionID, v1.ID)
	}

	// --- RBAC: viewer cannot import CI results ---
	ciBody := map[string]any{
		"source":       "gitlab-ci",
		"pipeline_url": "https://gitlab.example.com/pipelines/42",
		"results": []map[string]any{
			{"test_case_id": v1.TestCaseID, "status": "passed", "comment": "green"},
		},
	}
	if code, _ := c.do("POST", "/api/v1/test-runs/"+run.ID+"/ci-import", viewerTok, ciBody); code != http.StatusForbidden {
		t.Fatalf("viewer CI import: want 403, got %d", code)
	}

	// --- GitLab CI import (admin) applies results onto the frozen run ---
	var summary testrun.CIImportSummary
	c.json("POST", "/api/v1/test-runs/"+run.ID+"/ci-import", adminTok, ciBody, http.StatusOK, &summary)
	if summary.Imported != 1 || summary.Skipped != 0 {
		t.Fatalf("CI import summary: want imported=1 skipped=0, got imported=%d skipped=%d", summary.Imported, summary.Skipped)
	}
	if summary.Stats.Passed != 1 {
		t.Fatalf("CI import stats: want passed=1, got %d", summary.Stats.Passed)
	}

	// result is now passed AND still frozen at v1
	final := c.results(adminTok, run.ID)
	if final[0].Status != "passed" {
		t.Fatalf("result status: want passed, got %s", final[0].Status)
	}
	if final[0].TestCaseVersionID != v1.ID {
		t.Fatalf("FREEZE BROKEN after CI import: got %s, expected v1 %s", final[0].TestCaseVersionID, v1.ID)
	}
}

// buildTestServer wires every module against the test database. db is nil
// because /health (the only consumer) is never hit in this test.
func buildTestServer(dbx *sqlx.DB) *Server {
	jwtMgr := auth.NewJWTManager([]byte("test-secret"), time.Hour)
	return &Server{
		jwt:         jwtMgr,
		auth:        auth.NewAuthHandler(auth.NewAuthService(auth.NewRepository(dbx), jwtMgr)),
		project:     project.NewHandler(project.NewService(project.NewRepository(dbx))),
		folder:      folder.NewHandler(folder.NewService(folder.NewRepository(dbx))),
		testcase:    testcase.NewHandler(testcase.NewService(testcase.NewRepository(dbx))),
		suite:       suite.NewHandler(suite.NewService(suite.NewRepository(dbx))),
		testrun:     testrun.NewHandler(testrun.NewService(testrun.NewRepository(dbx))),
		integration: integration.NewHandler(integration.NewService(integration.NewRepository(dbx))),
	}
}

// --- tiny HTTP test client ---

type client struct {
	t    *testing.T
	base string
}

func (c *client) do(method, path, token string, body any) (int, []byte) {
	c.t.Helper()
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.base+path, reader)
	if err != nil {
		c.t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatalf("do %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, data
}

func (c *client) json(method, path, token string, body any, wantStatus int, out any) {
	c.t.Helper()
	code, data := c.do(method, path, token, body)
	if code != wantStatus {
		c.t.Fatalf("%s %s: want %d, got %d (%s)", method, path, wantStatus, code, string(data))
	}
	if out != nil {
		if err := json.Unmarshal(data, out); err != nil {
			c.t.Fatalf("decode %s %s: %v (%s)", method, path, err, string(data))
		}
	}
}

func (c *client) register(name, email, password, role string) {
	c.t.Helper()
	c.json("POST", "/api/v1/auth/register", "", map[string]any{
		"name": name, "email": email, "password": password, "role": role,
	}, http.StatusCreated, nil)
}

func (c *client) login(email, password string) string {
	c.t.Helper()
	var resp dto.LoginResponse
	c.json("POST", "/api/v1/auth/login", "", map[string]any{
		"email": email, "password": password,
	}, http.StatusOK, &resp)
	return resp.AccessToken
}

func (c *client) results(token, runID string) []model.TestRunResult {
	c.t.Helper()
	var results []model.TestRunResult
	c.json("GET", "/api/v1/test-runs/"+runID+"/results", token, nil, http.StatusOK, &results)
	return results
}
