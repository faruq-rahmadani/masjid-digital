-- +goose Up
-- who created the user; NULL for users nobody created, e.g. the first super admin
ALTER TABLE users ADD COLUMN created_by BIGINT REFERENCES users(id);

-- +goose Down
ALTER TABLE users DROP COLUMN created_by;
