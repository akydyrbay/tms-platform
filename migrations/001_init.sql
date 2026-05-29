CREATE TYPE "priority_enum" AS ENUM (
  'low',
  'medium',
  'high'
);

CREATE TYPE "test_run_status_enum" AS ENUM (
  'pending',
  'in_progress',
  'completed',
  'cancelled'
);

CREATE TYPE "test_result_status_enum" AS ENUM (
  'passed',
  'failed',
  'blocked',
  'skipped',
  'in_progress',
  'not_run'
);

CREATE TABLE "users" (
  "id" uuid PRIMARY KEY,
  "username" varchar NOT NULL,
  "role" varchar NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "projects" (
  "id" uuid PRIMARY KEY,
  "name" text NOT NULL,
  "description" text,
  "created_by" uuid NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "test_cases" (
  "id" uuid PRIMARY KEY,
  "project_id" uuid NOT NULL,
  "title" text NOT NULL,
  "description" text,
  "priority" priority_enum NOT NULL DEFAULT 'medium',
  "created_by" uuid NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "test_case_steps" (
  "test_case_id" uuid NOT NULL,
  "step_number" int NOT NULL,
  "action" text NOT NULL,
  "expected_result" text,
  PRIMARY KEY ("test_case_id", "step_number")
);

CREATE TABLE "test_suites" (
  "id" uuid PRIMARY KEY,
  "project_id" uuid NOT NULL,
  "name" text NOT NULL,
  "description" text,
  "created_by" uuid NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "suite_test_cases" (
  "suite_id" uuid NOT NULL,
  "test_case_id" uuid NOT NULL,
  PRIMARY KEY ("suite_id", "test_case_id")
);

CREATE TABLE "test_runs" (
  "id" uuid PRIMARY KEY,
  "project_id" uuid NOT NULL,
  "suite_id" uuid NOT NULL,
  "name" text NOT NULL,
  "status" test_run_status_enum NOT NULL DEFAULT 'pending',
  "created_by" uuid NOT NULL,
  "started_at" timestamp,
  "completed_at" timestamp,
  "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "test_run_results" (
  "id" uuid PRIMARY KEY,
  "test_run_id" uuid NOT NULL,
  "test_case_id" uuid NOT NULL,
  "status" test_result_status_enum NOT NULL,
  "comment" text,
  "executed_by" uuid NOT NULL,
  "executed_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "bugs" (
  "id" uuid PRIMARY KEY,
  "test_run_result_id" uuid NOT NULL,
  "tracker" varchar NOT NULL,
  "external_id" varchar NOT NULL,
  "url" text,
  "title" text,
  "created_by" uuid NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now())
);

ALTER TABLE "projects" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "test_cases" ADD FOREIGN KEY ("project_id") REFERENCES "projects" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "test_cases" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "test_case_steps" ADD FOREIGN KEY ("test_case_id") REFERENCES "test_cases" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "test_suites" ADD FOREIGN KEY ("project_id") REFERENCES "projects" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "test_suites" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "suite_test_cases" ADD FOREIGN KEY ("suite_id") REFERENCES "test_suites" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "suite_test_cases" ADD FOREIGN KEY ("test_case_id") REFERENCES "test_cases" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "test_runs" ADD FOREIGN KEY ("project_id") REFERENCES "projects" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "test_runs" ADD FOREIGN KEY ("suite_id") REFERENCES "test_suites" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "test_runs" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "test_run_results" ADD FOREIGN KEY ("test_run_id") REFERENCES "test_runs" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "test_run_results" ADD FOREIGN KEY ("test_case_id") REFERENCES "test_cases" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "test_run_results" ADD FOREIGN KEY ("executed_by") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "bugs" ADD FOREIGN KEY ("test_run_result_id") REFERENCES "test_run_results" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "bugs" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;

-- Defer constraint checking for INSERT
BEGIN;
SET CONSTRAINTS ALL DEFERRED;

INSERT INTO "users" ("id", "username", "role", "created_at")
VALUES
  (CAST('11111111-1111-1111-1111-111111111111' AS uuid), 'alice', 'admin', '2026-05-29T09:00:00'),
  (CAST('22222222-2222-2222-2222-222222222222' AS uuid), 'bob', 'qa', '2026-05-29T09:05:00'),
  (CAST('33333333-3333-3333-3333-333333333333' AS uuid), 'diana', 'dev', '2026-05-29T09:10:00');
INSERT INTO "projects" ("id", "name", "description", "created_by", "created_at")
VALUES
  (CAST('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS uuid), 'E-Commerce App', 'Main release testing project', CAST('11111111-1111-1111-1111-111111111111' AS uuid), '2026-05-29T10:00:00'),
  (CAST('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb' AS uuid), 'Mobile Banking', 'Regression testing for banking app', CAST('11111111-1111-1111-1111-111111111111' AS uuid), '2026-05-29T10:10:00');
INSERT INTO "test_cases" ("id", "project_id", "title", "description", "priority", "created_by", "created_at", "updated_at")
VALUES
  (CAST('f1111111-1111-1111-1111-111111111111' AS uuid), CAST('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS uuid), 'User can login', 'Verify login with valid credentials', 'high', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T10:15:00', '2026-05-29T10:15:00'),
  (CAST('f2222222-2222-2222-2222-222222222222' AS uuid), CAST('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS uuid), 'User can checkout', 'Verify checkout flow with card payment', 'high', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T10:16:00', '2026-05-29T10:16:00'),
  (CAST('f3333333-3333-3333-3333-333333333333' AS uuid), CAST('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb' AS uuid), 'User can reset password', 'Verify password reset flow', 'medium', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T10:17:00', '2026-05-29T10:17:00');
INSERT INTO "test_case_steps" ("test_case_id", "step_number", "action", "expected_result")
VALUES
  (CAST('f1111111-1111-1111-1111-111111111111' AS uuid), 1, 'Navigate to the login page', 'Login form is displayed'),
  (CAST('f1111111-1111-1111-1111-111111111111' AS uuid), 2, 'Enter valid credentials and submit', 'User is redirected to the dashboard'),
  (CAST('f2222222-2222-2222-2222-222222222222' AS uuid), 1, 'Add an item to the cart and proceed to checkout', 'Checkout page is displayed'),
  (CAST('f2222222-2222-2222-2222-222222222222' AS uuid), 2, 'Enter card details and confirm payment', 'Order confirmation is displayed'),
  (CAST('f3333333-3333-3333-3333-333333333333' AS uuid), 1, 'Click "Forgot password" and submit the account email', 'Password reset email is sent');
INSERT INTO "test_suites" ("id", "project_id", "name", "description", "created_by", "created_at")
VALUES
  (CAST('e1111111-1111-1111-1111-111111111111' AS uuid), CAST('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS uuid), 'Regression Suite', 'Core regression tests before release', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T10:20:00'),
  (CAST('e2222222-2222-2222-2222-222222222222' AS uuid), CAST('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb' AS uuid), 'Smoke Suite', 'Quick critical checks', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T10:21:00');
INSERT INTO "suite_test_cases" ("suite_id", "test_case_id")
VALUES
  (CAST('e1111111-1111-1111-1111-111111111111' AS uuid), CAST('f1111111-1111-1111-1111-111111111111' AS uuid)),
  (CAST('e1111111-1111-1111-1111-111111111111' AS uuid), CAST('f2222222-2222-2222-2222-222222222222' AS uuid)),
  (CAST('e2222222-2222-2222-2222-222222222222' AS uuid), CAST('f3333333-3333-3333-3333-333333333333' AS uuid));
INSERT INTO "test_runs" ("id", "project_id", "suite_id", "name", "status", "created_by", "started_at", "completed_at", "created_at")
VALUES
  (CAST('d1111111-1111-1111-1111-111111111111' AS uuid), CAST('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS uuid), CAST('e1111111-1111-1111-1111-111111111111' AS uuid), 'Release 1.0 Regression Run', 'completed', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T11:00:00', '2026-05-29T12:00:00', '2026-05-29T11:00:00'),
  (CAST('d2222222-2222-2222-2222-222222222222' AS uuid), CAST('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb' AS uuid), CAST('e2222222-2222-2222-2222-222222222222' AS uuid), 'Nightly Smoke Run', 'in_progress', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T12:30:00', NULL, '2026-05-29T12:30:00');
INSERT INTO "test_run_results" ("id", "test_run_id", "test_case_id", "status", "comment", "executed_by", "executed_at")
VALUES
  (CAST('c1111111-1111-1111-1111-111111111111' AS uuid), CAST('d1111111-1111-1111-1111-111111111111' AS uuid), CAST('f1111111-1111-1111-1111-111111111111' AS uuid), 'passed', 'Login works as expected', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T11:10:00'),
  (CAST('c2222222-2222-2222-2222-222222222222' AS uuid), CAST('d1111111-1111-1111-1111-111111111111' AS uuid), CAST('f2222222-2222-2222-2222-222222222222' AS uuid), 'failed', 'Payment gateway returned 500', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T11:20:00'),
  (CAST('c3333333-3333-3333-3333-333333333333' AS uuid), CAST('d2222222-2222-2222-2222-222222222222' AS uuid), CAST('f3333333-3333-3333-3333-333333333333' AS uuid), 'passed', 'Password reset email sent', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T12:40:00');
INSERT INTO "bugs" ("id", "test_run_result_id", "tracker", "external_id", "url", "title", "created_by", "created_at")
VALUES
  (CAST('b1111111-1111-1111-1111-111111111111' AS uuid), CAST('c2222222-2222-2222-2222-222222222222' AS uuid), 'jira', 'ECOM-501', 'https://jira.example.com/browse/ECOM-501', 'Checkout payment gateway returns 500', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T11:25:00');

SET CONSTRAINTS ALL IMMEDIATE;
COMMIT;
