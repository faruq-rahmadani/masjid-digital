-- +goose Up
-- the "user" role is gone: roles are super_admin, takmir and bendahara (domain.UserRole)
UPDATE users SET role = 'takmir' WHERE role = 'user';
-- the role is now always chosen when creating a user
ALTER TABLE users ALTER COLUMN role DROP DEFAULT;

-- +goose Down
-- the data change is not undone: takmir rows cannot be told apart from ones that were "user"
ALTER TABLE users ALTER COLUMN role SET DEFAULT 'user';
