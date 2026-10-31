package users

import (
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
)

// userColumns lists every column of the users table, for selecting a full userModel
var userColumns = []string{
	"id",
	"name",
	"phone",
	"email",
	"password_hash",
	"role",
	"status",
	"last_login_at",
	"created_by",
	"created_at",
	"updated_at",
}

// userModel is a row of the users table. It stays inside this package;
// other layers work with domain.User.
type userModel struct {
	ID           int64      `db:"id"`
	Name         string     `db:"name"`
	Phone        string     `db:"phone"`
	Email        *string    `db:"email"` // NULL when unset, so UNIQUE ignores it
	PasswordHash string     `db:"password_hash"`
	Role         string     `db:"role"`
	Status       string     `db:"status"`
	LastLoginAt  *time.Time `db:"last_login_at"`
	CreatedBy    *int64     `db:"created_by"` // NULL when nobody created the user
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

// newUserModel converts a domain.User into a row, storing times in UTC
func newUserModel(user domain.User) userModel {
	m := userModel{
		ID:           user.ID,
		Name:         user.Name,
		Phone:        user.Phone,
		PasswordHash: user.PasswordHash,
		Role:         string(user.Role),
		Status:       string(user.Status),
		CreatedAt:    user.CreatedAt.UTC(),
		UpdatedAt:    user.UpdatedAt.UTC(),
	}

	if user.Email != "" {
		m.Email = &user.Email
	}

	if user.LastLoginAt != nil {
		lastLoginAt := user.LastLoginAt.UTC()
		m.LastLoginAt = &lastLoginAt
	}

	// 0 means nobody; store NULL since no user has ID 0 for the foreign key to point at
	if user.CreatedBy != 0 {
		m.CreatedBy = &user.CreatedBy
	}

	return m
}

// toDomain converts the row into a domain.User. Times come back in UTC whatever
// the engine returns: pgx reads TIMESTAMPTZ in the server's local zone, SQLite in UTC.
func (m userModel) toDomain() domain.User {
	user := domain.User{
		ID:           m.ID,
		Name:         m.Name,
		Phone:        m.Phone,
		PasswordHash: m.PasswordHash,
		Role:         domain.UserRole(m.Role),
		Status:       domain.UserStatus(m.Status),
		CreatedAt:    m.CreatedAt.UTC(),
		UpdatedAt:    m.UpdatedAt.UTC(),
	}

	if m.LastLoginAt != nil {
		lastLoginAt := m.LastLoginAt.UTC()
		user.LastLoginAt = &lastLoginAt
	}

	if m.Email != nil {
		user.Email = *m.Email
	}

	if m.CreatedBy != nil {
		user.CreatedBy = *m.CreatedBy
	}

	return user
}
