CREATE TYPE "priority_enum" AS ENUM (
  'low',
  'medium',
  'high'
);
 
CREATE TYPE "case_status_enum" AS ENUM (
  'draft',
  'active',
  'deprecated'
);
 
CREATE TYPE "case_type_enum" AS ENUM (
  'functional',
  'smoke',
  'regression',
  'integration',
  'e2e'
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
  "name" text NOT NULL,
  "email" text UNIQUE NOT NULL,
  "password_hash" text NOT NULL,
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
 
CREATE TABLE "folders" (
  "id" uuid PRIMARY KEY,
  "project_id" uuid NOT NULL,
  "parent_id" uuid,
  "name" text NOT NULL,
  "created_by" uuid NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now())
);
 
CREATE TABLE "test_cases" (
  "id" uuid PRIMARY KEY,
  "project_id" uuid NOT NULL,
  "folder_id" uuid,
  "created_by" uuid NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now())
);
 
CREATE TABLE "test_case_versions" (
  "id" uuid PRIMARY KEY,
  "test_case_id" uuid NOT NULL,
  "version_number" int NOT NULL,
  "title" text NOT NULL,
  "description" text,
  "preconditions" text,
  "expected_result" text,
  "module" text,
  "component" text,
  "priority" priority_enum NOT NULL DEFAULT 'medium',
  "type" case_type_enum NOT NULL DEFAULT 'functional',
  "status" case_status_enum NOT NULL DEFAULT 'active',
  "created_by" uuid NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  UNIQUE ("test_case_id", "version_number")
);
 
CREATE TABLE "test_case_steps" (
  "test_case_version_id" uuid NOT NULL,
  "step_number" int NOT NULL,
  "action" text NOT NULL,
  "expected_result" text,
  PRIMARY KEY ("test_case_version_id", "step_number")
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
  "test_case_version_id" uuid NOT NULL,
  "status" test_result_status_enum NOT NULL DEFAULT 'not_run',
  "comment" text,
  "executed_by" uuid,
  "executed_at" timestamp
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
 
ALTER TABLE "folders" ADD FOREIGN KEY ("project_id") REFERENCES "projects" ("id") DEFERRABLE INITIALLY IMMEDIATE;
 
ALTER TABLE "folders" ADD FOREIGN KEY ("parent_id") REFERENCES "folders" ("id") DEFERRABLE INITIALLY IMMEDIATE;
 
ALTER TABLE "folders" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;
 
ALTER TABLE "test_cases" ADD FOREIGN KEY ("project_id") REFERENCES "projects" ("id") DEFERRABLE INITIALLY IMMEDIATE;
 
ALTER TABLE "test_cases" ADD FOREIGN KEY ("folder_id") REFERENCES "folders" ("id") DEFERRABLE INITIALLY IMMEDIATE;
 
ALTER TABLE "test_cases" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;
 
ALTER TABLE "test_case_versions" ADD FOREIGN KEY ("test_case_id") REFERENCES "test_cases" ("id") DEFERRABLE INITIALLY IMMEDIATE;
 
ALTER TABLE "test_case_versions" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;
 
ALTER TABLE "test_case_steps" ADD FOREIGN KEY ("test_case_version_id") REFERENCES "test_case_versions" ("id") DEFERRABLE INITIALLY IMMEDIATE;
 
ALTER TABLE "test_suites" ADD FOREIGN KEY ("project_id") REFERENCES "projects" ("id") DEFERRABLE INITIALLY IMMEDIATE;
 
ALTER TABLE "test_suites" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;
 
ALTER TABLE "suite_test_cases" ADD FOREIGN KEY ("suite_id") REFERENCES "test_suites" ("id") DEFERRABLE INITIALLY IMMEDIATE;
 
ALTER TABLE "suite_test_cases" ADD FOREIGN KEY ("test_case_id") REFERENCES "test_cases" ("id") DEFERRABLE INITIALLY IMMEDIATE;
 
ALTER TABLE "test_runs" ADD FOREIGN KEY ("project_id") REFERENCES "projects" ("id") DEFERRABLE INITIALLY IMMEDIATE;
 
ALTER TABLE "test_runs" ADD FOREIGN KEY ("suite_id") REFERENCES "test_suites" ("id") DEFERRABLE INITIALLY IMMEDIATE;
 
ALTER TABLE "test_runs" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;
 
ALTER TABLE "test_run_results" ADD FOREIGN KEY ("test_run_id") REFERENCES "test_runs" ("id") DEFERRABLE INITIALLY IMMEDIATE;
 
ALTER TABLE "test_run_results" ADD FOREIGN KEY ("test_case_id") REFERENCES "test_cases" ("id") DEFERRABLE INITIALLY IMMEDIATE;
 
ALTER TABLE "test_run_results" ADD FOREIGN KEY ("test_case_version_id") REFERENCES "test_case_versions" ("id") DEFERRABLE INITIALLY IMMEDIATE;
 
ALTER TABLE "test_run_results" ADD FOREIGN KEY ("executed_by") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;
 
ALTER TABLE "bugs" ADD FOREIGN KEY ("test_run_result_id") REFERENCES "test_run_results" ("id") DEFERRABLE INITIALLY IMMEDIATE;
 
ALTER TABLE "bugs" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;

BEGIN;
SET CONSTRAINTS ALL DEFERRED;
 
INSERT INTO "users" ("id", "name", "email", "password_hash", "role", "created_at")
VALUES
  (CAST('11111111-1111-1111-1111-111111111111' AS uuid), 'alice', 'alice@example.com', 'x', 'admin', '2026-05-29T09:00:00'),
  (CAST('22222222-2222-2222-2222-222222222222' AS uuid), 'bob', 'bob@example.com', 'x', 'qa', '2026-05-29T09:05:00'),
  (CAST('33333333-3333-3333-3333-333333333333' AS uuid), 'diana', 'diana@example.com', 'x', 'dev', '2026-05-29T09:10:00');
 
INSERT INTO "projects" ("id", "name", "description", "created_by", "created_at")
VALUES
  (CAST('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS uuid), 'E-Commerce App', 'Main release testing project', CAST('11111111-1111-1111-1111-111111111111' AS uuid), '2026-05-29T10:00:00'),
  (CAST('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb' AS uuid), 'Mobile Banking', 'Regression testing for banking app', CAST('11111111-1111-1111-1111-111111111111' AS uuid), '2026-05-29T10:10:00');
 
INSERT INTO "folders" ("id", "project_id", "parent_id", "name", "created_by", "created_at")
VALUES
  (CAST('91111111-1111-1111-1111-111111111111' AS uuid), CAST('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS uuid), NULL, 'Auth', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T10:12:00'),
  (CAST('92222222-2222-2222-2222-222222222222' AS uuid), CAST('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS uuid), CAST('91111111-1111-1111-1111-111111111111' AS uuid), 'Login', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T10:13:00'),
  (CAST('93333333-3333-3333-3333-333333333333' AS uuid), CAST('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb' AS uuid), NULL, 'Accounts', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T10:14:00');
 
INSERT INTO "test_cases" ("id", "project_id", "folder_id", "created_by", "created_at")
VALUES
  (CAST('f1111111-1111-1111-1111-111111111111' AS uuid), CAST('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS uuid), CAST('92222222-2222-2222-2222-222222222222' AS uuid), CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T10:15:00'),
  (CAST('f2222222-2222-2222-2222-222222222222' AS uuid), CAST('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AS uuid), NULL, CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T10:16:00'),
  (CAST('f3333333-3333-3333-3333-333333333333' AS uuid), CAST('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb' AS uuid), CAST('93333333-3333-3333-3333-333333333333' AS uuid), CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T10:17:00');
 
INSERT INTO "test_case_versions" ("id", "test_case_id", "version_number", "title", "description", "preconditions", "expected_result", "module", "component", "priority", "type", "status", "created_by", "created_at")
VALUES
  (CAST('81111111-1111-1111-1111-111111111111' AS uuid), CAST('f1111111-1111-1111-1111-111111111111' AS uuid), 1, 'User can login', 'Verify login with valid credentials', 'User is registered', 'User is logged in and redirected to the dashboard', 'Auth', 'Login', 'high', 'functional', 'active', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T10:15:00'),
  (CAST('81111112-1111-1111-1111-111111111111' AS uuid), CAST('f1111111-1111-1111-1111-111111111111' AS uuid), 2, 'User can login', 'Verify login with valid credentials', 'User is registered', 'User is logged in and redirected to the dashboard', 'Auth', 'Login', 'high', 'functional', 'active', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T12:30:00'),
  (CAST('82222222-2222-2222-2222-222222222222' AS uuid), CAST('f2222222-2222-2222-2222-222222222222' AS uuid), 1, 'User can checkout', 'Verify checkout flow with card payment', 'User has items in the cart', 'Order confirmation is displayed', 'Checkout', 'Payment', 'high', 'functional', 'active', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T10:16:00'),
  (CAST('83333333-3333-3333-3333-333333333333' AS uuid), CAST('f3333333-3333-3333-3333-333333333333' AS uuid), 1, 'User can reset password', 'Verify password reset flow', 'User account exists', 'Password reset email is sent', 'Accounts', 'Auth', 'medium', 'functional', 'active', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T10:17:00');
 
INSERT INTO "test_case_steps" ("test_case_version_id", "step_number", "action", "expected_result")
VALUES
  (CAST('81111111-1111-1111-1111-111111111111' AS uuid), 1, 'Navigate to the login page', 'Login form is displayed'),
  (CAST('81111111-1111-1111-1111-111111111111' AS uuid), 2, 'Enter valid credentials and click "Login"', 'User is redirected to the dashboard'),
  (CAST('81111112-1111-1111-1111-111111111111' AS uuid), 1, 'Navigate to the login page', 'Login form is displayed'),
  (CAST('81111112-1111-1111-1111-111111111111' AS uuid), 2, 'Enter valid credentials and click "Continue"', 'User is redirected to the dashboard'),
  (CAST('82222222-2222-2222-2222-222222222222' AS uuid), 1, 'Add an item to the cart and proceed to checkout', 'Checkout page is displayed'),
  (CAST('82222222-2222-2222-2222-222222222222' AS uuid), 2, 'Enter card details and confirm payment', 'Order confirmation is displayed'),
  (CAST('83333333-3333-3333-3333-333333333333' AS uuid), 1, 'Click "Forgot password" and submit the account email', 'Password reset email is sent');
 
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
 
INSERT INTO "test_run_results" ("id", "test_run_id", "test_case_id", "test_case_version_id", "status", "comment", "executed_by", "executed_at")
VALUES
  (CAST('c1111111-1111-1111-1111-111111111111' AS uuid), CAST('d1111111-1111-1111-1111-111111111111' AS uuid), CAST('f1111111-1111-1111-1111-111111111111' AS uuid), CAST('81111111-1111-1111-1111-111111111111' AS uuid), 'passed', 'Login works as expected', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T11:10:00'),
  (CAST('c2222222-2222-2222-2222-222222222222' AS uuid), CAST('d1111111-1111-1111-1111-111111111111' AS uuid), CAST('f2222222-2222-2222-2222-222222222222' AS uuid), CAST('82222222-2222-2222-2222-222222222222' AS uuid), 'failed', 'Payment gateway returned 500', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T11:20:00'),
  (CAST('c3333333-3333-3333-3333-333333333333' AS uuid), CAST('d2222222-2222-2222-2222-222222222222' AS uuid), CAST('f3333333-3333-3333-3333-333333333333' AS uuid), CAST('83333333-3333-3333-3333-333333333333' AS uuid), 'passed', 'Password reset email sent', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T12:40:00');
 
INSERT INTO "bugs" ("id", "test_run_result_id", "tracker", "external_id", "url", "title", "created_by", "created_at")
VALUES
  (CAST('b1111111-1111-1111-1111-111111111111' AS uuid), CAST('c2222222-2222-2222-2222-222222222222' AS uuid), 'jira', 'ECOM-501', 'https://jira.example.com/browse/ECOM-501', 'Checkout payment gateway returns 500', CAST('22222222-2222-2222-2222-222222222222' AS uuid), '2026-05-29T11:25:00');
 
SET CONSTRAINTS ALL IMMEDIATE;
COMMIT;