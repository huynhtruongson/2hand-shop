CREATE TABLE users (
    id          TEXT PRIMARY KEY,
    auth_provider VARCHAR(100)        NOT NULL,
    auth_provider_id TEXT            NOT NULL UNIQUE,
    email       VARCHAR(100)        NOT NULL UNIQUE,
    name        VARCHAR(200)        NOT NULL,
    gender      VARCHAR(10)         NOT NULL CHECK (gender IN ('male', 'female')),
    verified_at TIMESTAMPTZ,
    role        VARCHAR(100)     NOT NULL DEFAULT 'client',
    created_at  TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);