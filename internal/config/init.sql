-- 001_init.sql
-- For UUID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ---------------------------
-- projects
-- ---------------------------
CREATE TABLE projects (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  slug        text NOT NULL UNIQUE,
  name        text NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now()
);

-- ---------------------------
-- environments (dev/staging/prod)
-- ---------------------------
CREATE TABLE environments (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id  uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  slug        text NOT NULL,
  name        text NOT NULL,
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  UNIQUE (project_id, slug)
);

CREATE INDEX idx_env_project_id ON environments(project_id);

-- ---------------------------
-- environment_settings (sane limits without auth)
-- ---------------------------
CREATE TABLE environment_settings (
  environment_id   uuid PRIMARY KEY REFERENCES environments(id) ON DELETE CASCADE,
  rpm_limit        integer NOT NULL DEFAULT 60,
  max_delay_ms     integer NOT NULL DEFAULT 10000,
  max_body_bytes   integer NOT NULL DEFAULT 524288, -- 512 KB
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),
  CHECK (rpm_limit >= 0),
  CHECK (max_delay_ms >= 0),
  CHECK (max_body_bytes >= 1024)
);

-- ---------------------------
-- spec_versions (raw + compiled snapshot)
-- ---------------------------
CREATE TABLE spec_versions (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  environment_id     uuid NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
  version            integer NOT NULL,              -- monotonically increasing per env
  raw_spec_text      text NOT NULL,                 -- uploaded YAML/JSON
  raw_spec_sha256    char(64) NOT NULL,             -- for dedupe/debug
  compiler_version   text NOT NULL DEFAULT 'v1',     -- bump when snapshot format changes
  compile_status     text NOT NULL,                 -- 'compiled' | 'failed'
  compile_errors     jsonb,                         -- structured errors if failed
  compiled_snapshot  jsonb,                         -- present when compiled
  created_at         timestamptz NOT NULL DEFAULT now(),
  UNIQUE (environment_id, version),
  CHECK (compile_status IN ('compiled', 'failed'))
);

CREATE INDEX idx_spec_versions_env_created ON spec_versions(environment_id, created_at DESC);
CREATE INDEX idx_spec_versions_env_status  ON spec_versions(environment_id, compile_status);

-- Helpful for querying JSON if needed later
CREATE INDEX idx_spec_versions_snapshot_gin ON spec_versions USING GIN (compiled_snapshot);

-- ---------------------------
-- deployments (what is active)
-- ---------------------------
CREATE TABLE deployments (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  environment_id   uuid NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
  spec_version_id  uuid NOT NULL REFERENCES spec_versions(id) ON DELETE RESTRICT,
  is_active        boolean NOT NULL DEFAULT false,
  created_at       timestamptz NOT NULL DEFAULT now(),
  activated_at     timestamptz
);

CREATE INDEX idx_deployments_env_created ON deployments(environment_id, created_at DESC);

-- Only one active deployment per environment:
CREATE UNIQUE INDEX uniq_active_deployment_per_env
ON deployments(environment_id)
WHERE is_active;

-- ---------------------------
-- api_keys (optional now, needed soon)
-- ---------------------------
CREATE TABLE api_keys (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  environment_id   uuid NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
  key_prefix       text NOT NULL,          -- e.g. first 8 chars for lookup
  key_hash         text NOT NULL,          -- store hash only (argon2/bcrypt)
  label            text,
  is_active        boolean NOT NULL DEFAULT true,
  created_at       timestamptz NOT NULL DEFAULT now(),
  last_used_at     timestamptz,
  UNIQUE (environment_id, key_prefix)
);

CREATE INDEX idx_api_keys_env_active ON api_keys(environment_id, is_active);