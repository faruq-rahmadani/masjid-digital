-- +goose Up
-- Whether the TV shows the slide's title and caption under it; off for posters that already
-- carry their own text, which then fill the whole panel.
ALTER TABLE slides ADD COLUMN show_caption BOOLEAN NOT NULL DEFAULT TRUE;

-- +goose Down
ALTER TABLE slides DROP COLUMN show_caption;
