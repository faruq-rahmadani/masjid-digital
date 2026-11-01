-- +goose Up
-- Petugas Jumat: who leads one Friday's prayer. Iqamah, mode sholat and alarm settings live in
-- configurations under iqamah_settings.
CREATE TABLE friday_duties (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    date       TEXT NOT NULL UNIQUE,         -- the Friday, local date of the mosque, YYYY-MM-DD
    khatib     TEXT NOT NULL DEFAULT '',
    imam       TEXT NOT NULL DEFAULT '',
    muadzin    TEXT NOT NULL DEFAULT '',
    theme      TEXT NOT NULL DEFAULT '',     -- tema khutbah
    updated_by BIGINT REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE friday_duties;
