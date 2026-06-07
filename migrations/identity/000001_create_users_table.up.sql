CREATE TABLE users (
    id              TEXT PRIMARY KEY,
    auth_provider   VARCHAR(100)        NOT NULL,
    auth_provider_id TEXT            NOT NULL UNIQUE,
    email               VARCHAR(100)        NOT NULL UNIQUE,
    first_name        VARCHAR(200)        NOT NULL,
    last_name        VARCHAR(200)        NOT NULL,
    gender      VARCHAR(10)         NOT NULL CHECK (gender IN ('male', 'female')),
    email_verified BOOLEAN             NOT NULL DEFAULT FALSE,
    role        VARCHAR(100)     NOT NULL DEFAULT 'client',
    created_at  TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);