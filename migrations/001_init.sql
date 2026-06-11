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
