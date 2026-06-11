# TMS Platform
Corporate **Test Management System** (TMS) for the centralized storage of test cases, their versions, their execution in test runs, bug links to external trackers, and CI/CD importing.

- **Backend:** Go (chi, sqlx, pgx, jwt, bcrypt) - a modular monolith.
- **Frontend:** React + TypeScript + Vite (dependency free).
- **Database:** PostgreSQL.

---

## Architecture

```
Frontend (React)
    ↓
API Gateway
    ↓
TMS Core Service
├── Test Case Module
├── Execution Module
├── Reporting Module
├── Integration Module
    ↓
PostgreSQL
```

Each module is three files with one job each:

| Layer | File | Responsibility |
|---|---|---|
| Handler | `handler.go` | HTTP: decode request -> call service -> write JSON |
| Service | `service.go` | Business rules, validation, sentinel errors |
| Repository | `repository.go` | SQL only |

---

## To Run it

```bash
# 1. Config (JWT_SECRET, DB_* and PORT live here)
cp .env.example .env

# 2. Start API + Postgres + Front
docker compose up --build   
```

Handy targets: `make docker-run`, `make docker-down`, `make test`.

---

## Core logic

A **suite** is a reusable set of test cases (e.g. Smoke, Regression). A **test run** is created from a suite and snapshots the current version of each case at that moment.

- On `POST /test-runs`, a single transaction writes one `test_run_results` row per case with an explicit `test_case_version_id` 
- Editing a case afterwards inserts a new version row - old versions are never mutated.
- Marking / re-marking / CI-importing a result never touches `test_case_version_id`.

---

## RBAC (roles & access)

The role lives in the JWT and is enforced by `RequireRoles` middleware after auth.

| Role | Access |
|---|---|
| `viewer`, `developer` | **Read-only** (all `GET`) |
| `qa` | Read + create/edit cases, suites, runs, execute & import |
| `qa_lead`, `admin` | **Full access**, including destructive deletes |

- Reads (`GET`) any authenticated user.
- Mutations (`POST/PUT/PATCH`) -> `qa`, `qa_lead`, `admin` -> otherwise **403**.
- Destructive deletes -> `qa_lead`, `admin` only.

---

## API summary

```
POST /api/v1/auth/register | /login                 (public)

GET  /api/v1/projects                               POST (writer)
GET  /api/v1/folders?project_id=                    POST (writer)
GET  /api/v1/test-cases?project_id= | /{id}         POST/PUT (writer)
GET  /api/v1/test-cases/{id}/versions[/{n}]
GET  /api/v1/suites?project_id= | /{id}             POST cases (writer), DELETE cases (manager)
GET  /api/v1/test-runs?project_id= | /{id}          POST (writer)
GET  /api/v1/test-runs/{id}/results | /stats
POST /api/v1/test-runs/{id}/results/{caseID}        mark/re-mark (writer)
POST /api/v1/test-runs/{id}/ci-import               GitLab CI import (writer)
GET  /api/v1/run-results/{resultID}/bugs            POST attach bug (writer)
```

### GitLab CI import

`POST /api/v1/test-runs/{id}/ci-import` ingests a CI pipeline report and applies pass/fail onto the **frozen** run:

```json
{
  "source": "gitlab-ci",
  "pipeline_url": "https://gitlab.example.com/pipelines/42",
  "results": [
    { "test_case_id": "…", "status": "passed",  "comment": "job ok" },
    { "test_case_id": "…", "status": "failed",  "comment": "assertion x" }
  ]
}
```

Each entry is applied via the same `MarkResult` path (version stays frozen); the source + pipeline URL are stamped into the result comment. The response summarises `imported` / `skipped` (cases not in the run) and returns fresh stats.

---

## Testing

```bash
go test ./...                      # all tests (needs Docker for containers)
go test ./internal/server/ -v      # end-to-end: version freeze + RBAC + CI import
```

Integration tests use **testcontainers-go** to run a real Postgres with the migration applied - no mocks for the DB layer.
