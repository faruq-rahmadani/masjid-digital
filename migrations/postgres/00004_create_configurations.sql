-- +goose Up
-- one row per group of settings (domain.ConfigurationKey); each deployment serves one mosque
CREATE TABLE configurations (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL,                -- JSON, its shape depends on key (see internal/repository/database/mosqueprofiles)
    updated_by BIGINT REFERENCES users(id), -- NULL when nobody saved it, e.g. seeded
    updated_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE configurations;
