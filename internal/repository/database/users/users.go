package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database"
	pkgdb "github.com/faruq-rahmadani/masjid-digital/pkg/database"
)

const usersTable = "users"

// builder emits `?` placeholders; pkg/database rebinds them for the engine in use
var builder = sq.StatementBuilder.PlaceholderFormat(sq.Question)

type UserDBRepository struct {
	client database.DatabaseClient
}

func NewUserDBRepository(client database.DatabaseClient) *UserDBRepository {
	return &UserDBRepository{client: client}
}

// Create inserts a new user and returns its generated ID. The caller sets role
// and status; timestamps are set here. It returns domain.ErrUserAlreadyExists
// when the phone or email is taken.
func (r *UserDBRepository) Create(ctx context.Context, user domain.User) (int64, error) {
	m := newUserModel(user)
	now := time.Now().UTC()

	query, args, err := builder.
		Insert(usersTable).
		Columns(
			"name", "phone", "email", "password_hash", "role", "status", "created_by", "created_at", "updated_at",
		).
		Values(
			m.Name, m.Phone, m.Email, m.PasswordHash, m.Role, m.Status, m.CreatedBy, now, now,
		).
		Suffix("RETURNING id"). // LastInsertId is not supported on PostgreSQL
		ToSql()
	if err != nil {
		return 0, errInternal(fmt.Errorf("error building create user query: %w", err))
	}

	var id int64
	if err := r.client.Get(ctx, &id, query, args...); err != nil {
		if errors.Is(err, pkgdb.ErrUniqueViolation) {
			return 0, errUserAlreadyExists()
		}
		return 0, errInternal(fmt.Errorf("error creating user: %w", err))
	}

	return id, nil
}

// GetByPhone returns the user with the given E.164 phone, or domain.ErrUserNotFound
func (r *UserDBRepository) GetByPhone(ctx context.Context, phone string) (domain.User, error) {
	return r.getBy(ctx, "phone", phone)
}

// GetByEmail returns the user with the given lowercase email, or domain.ErrUserNotFound
func (r *UserDBRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	return r.getBy(ctx, "email", email)
}

// UpdateLastLoginAt records when the user last logged in
func (r *UserDBRepository) UpdateLastLoginAt(ctx context.Context, id int64, at time.Time) error {
	query, args, err := builder.
		Update(usersTable).
		Set("last_login_at", at.UTC()).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return errInternal(fmt.Errorf("error building update last login query: %w", err))
	}

	res, err := r.client.Exec(ctx, query, args...)
	if err != nil {
		return errInternal(fmt.Errorf("error updating last login: %w", err))
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return errInternal(fmt.Errorf("error reading updated rows: %w", err))
	}
	if affected == 0 {
		return errUserNotFound()
	}

	return nil
}

// getBy returns the single user whose column equals value
func (r *UserDBRepository) getBy(ctx context.Context, column string, value any) (domain.User, error) {
	query, args, err := builder.
		Select(userColumns...).
		From(usersTable).
		Where(sq.Eq{column: value}).
		ToSql()
	if err != nil {
		return domain.User{}, errInternal(fmt.Errorf("error building get user by %s query: %w", column, err))
	}

	var m userModel
	if err := r.client.Get(ctx, &m, query, args...); err != nil {
		if errors.Is(err, pkgdb.ErrNotFound) {
			return domain.User{}, errUserNotFound()
		}
		return domain.User{}, errInternal(fmt.Errorf("error getting user by %s: %w", column, err))
	}

	return m.toDomain(), nil
}

// FindAll returns every user, oldest first
func (r *UserDBRepository) FindAll(ctx context.Context) ([]domain.User, error) {
	query, args, err := builder.
		Select(userColumns...).
		From(usersTable).
		OrderBy("id").
		ToSql()
	if err != nil {
		return nil, errInternal(fmt.Errorf("error building find all users query: %w", err))
	}

	var models []userModel
	if err := r.client.Select(ctx, &models, query, args...); err != nil {
		return nil, errInternal(fmt.Errorf("error finding users: %w", err))
	}

	users := make([]domain.User, 0, len(models))
	for _, m := range models {
		users = append(users, m.toDomain())
	}

	return users, nil
}

// FindByID returns the user with the given ID, or domain.ErrUserNotFound
func (r *UserDBRepository) FindByID(ctx context.Context, id int64) (domain.User, error) {
	return r.getBy(ctx, "id", id)
}

// Update saves the editable fields of user (name, phone, email, role, status)
// and bumps updated_at. It returns domain.ErrUserNotFound when the user does
// not exist and domain.ErrUserAlreadyExists when the phone or email is taken.
func (r *UserDBRepository) Update(ctx context.Context, user domain.User) error {
	m := newUserModel(user)

	query, args, err := builder.
		Update(usersTable).
		Set("name", m.Name).
		Set("phone", m.Phone).
		Set("email", m.Email).
		Set("role", m.Role).
		Set("status", m.Status).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"id": m.ID}).
		ToSql()
	if err != nil {
		return errInternal(fmt.Errorf("error building update user query: %w", err))
	}

	res, err := r.client.Exec(ctx, query, args...)
	if err != nil {
		if errors.Is(err, pkgdb.ErrUniqueViolation) {
			return errUserAlreadyExists()
		}
		return errInternal(fmt.Errorf("error updating user: %w", err))
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return errInternal(fmt.Errorf("error reading updated rows: %w", err))
	}
	if affected == 0 {
		return errUserNotFound()
	}

	return nil
}

// UpdatePasswordHash replaces the user's password hash and bumps updated_at
func (r *UserDBRepository) UpdatePasswordHash(ctx context.Context, id int64, passwordHash string) error {
	query, args, err := builder.
		Update(usersTable).
		Set("password_hash", passwordHash).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return errInternal(fmt.Errorf("error building update password query: %w", err))
	}

	res, err := r.client.Exec(ctx, query, args...)
	if err != nil {
		return errInternal(fmt.Errorf("error updating password: %w", err))
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return errInternal(fmt.Errorf("error reading updated rows: %w", err))
	}
	if affected == 0 {
		return errUserNotFound()
	}

	return nil
}
