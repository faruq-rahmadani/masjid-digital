-- +goose Up
-- Texts scrolling in the TV's INFO bar. The scroll speed is one setting for all of them, in
-- configurations under running_text_setting.
CREATE TABLE running_texts (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    text       TEXT NOT NULL,                -- 1-200 characters (checked in domain)
    active     BOOLEAN NOT NULL DEFAULT TRUE,
    position   INTEGER NOT NULL,             -- display order, ascending
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE running_texts;
