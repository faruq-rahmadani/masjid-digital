-- +goose Up
-- Images rotating in the TV's main panel. The image never changes (a new picture is a new slide),
-- so it is served with a long cache. image_data is last so listing slides skips it.
CREATE TABLE slides (
    id               INTEGER PRIMARY KEY AUTOINCREMENT, -- never reused, which the image cache relies on
    title            TEXT NOT NULL,                     -- 1-100 characters (checked in domain)
    caption          TEXT NOT NULL DEFAULT '',          -- 0-200 characters
    duration_seconds INTEGER NOT NULL,                  -- 5, 10, 15, 20 or 30
    active           BOOLEAN NOT NULL DEFAULT TRUE,
    position         INTEGER NOT NULL,                  -- display order, ascending
    created_by       INTEGER REFERENCES users(id),
    created_at       TIMESTAMP NOT NULL,
    updated_at       TIMESTAMP NOT NULL,
    image_type       TEXT NOT NULL,                     -- image/png, image/jpeg or image/webp
    image_data       BLOB NOT NULL                      -- at most 5 MB
);

-- +goose Down
DROP TABLE slides;
