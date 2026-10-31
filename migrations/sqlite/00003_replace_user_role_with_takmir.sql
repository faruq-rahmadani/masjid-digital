-- +goose Up
-- the "user" role is gone: roles are super_admin, takmir and bendahara (domain.UserRole)
UPDATE users SET role = 'takmir' WHERE role = 'user';
-- SQLite cannot drop a column default without rebuilding the table, so the old
-- DEFAULT 'user' stays; it is never used because the app always inserts a role

-- +goose Down
-- the data change is not undone: takmir rows cannot be told apart from ones that were "user"
SELECT 1;
