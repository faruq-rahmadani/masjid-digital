-- +goose Up
-- When a running text was last turned off: the admin list shows it as "terakhir tampil". NULL while
-- active (it is on the TV now) and for texts never turned off.
ALTER TABLE running_texts ADD COLUMN deactivated_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE running_texts DROP COLUMN deactivated_at;
