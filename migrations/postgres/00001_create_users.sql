-- +goose Up
CREATE TABLE users (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name          TEXT NOT NULL,
    phone         TEXT NOT NULL UNIQUE,      -- login identifier, E.164 (+6281234567890), normalized in Go
    email         TEXT UNIQUE,               -- optional, stored lowercase
    password_hash TEXT NOT NULL,             -- argon2id
    role          TEXT NOT NULL DEFAULT 'user',   -- platform role, allowed values checked in the backend (domain.UserRole)
    status        TEXT NOT NULL DEFAULT 'active', -- allowed values checked in the backend (domain.UserStatus)
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE users;
