-- +goose Up
-- Petugas Jumat: who leads one Friday's prayer. Iqamah, mode sholat and alarm settings live in
-- configurations under iqamah_settings.
CREATE TABLE friday_duties (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    date       TEXT NOT NULL UNIQUE,         -- the Friday, local date of the mosque, YYYY-MM-DD
    khatib     TEXT NOT NULL DEFAULT '',
    imam       TEXT NOT NULL DEFAULT '',
    muadzin    TEXT NOT NULL DEFAULT '',
    theme      TEXT NOT NULL DEFAULT '',     -- tema khutbah
    updated_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE friday_duties;
