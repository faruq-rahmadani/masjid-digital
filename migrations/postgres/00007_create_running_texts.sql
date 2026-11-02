-- +goose Up
-- Texts scrolling in the TV's INFO bar. The scroll speed is one setting for all of them, in
-- configurations under running_text_setting.
CREATE TABLE running_texts (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    text       TEXT NOT NULL,                -- 1-200 characters (checked in domain)
    active     BOOLEAN NOT NULL DEFAULT TRUE,
    position   INT NOT NULL,                 -- display order, ascending
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE running_texts;
