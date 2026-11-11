package accounts

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	pkgdb "github.com/faruq-rahmadani/masjid-digital/pkg/database"
)

const accountsTable = "accounts"

// builder emits `?` placeholders; pkg/database rebinds them for the engine in use
var builder = sq.StatementBuilder.PlaceholderFormat(sq.Question)

var columns = []string{"id", "name", "kind", "number", "archived_at", "created_by", "created_at", "updated_at"}

type model struct {
	ID         int64      `db:"id"`
	Name       string     `db:"name"`
	Kind       string     `db:"kind"`
	Number     string     `db:"number"`
	ArchivedAt *time.Time `db:"archived_at"`
	CreatedBy  *int64     `db:"created_by"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
}

func (m model) toDomain() domain.Account {
	return domain.Account{
		ID:         m.ID,
		Name:       m.Name,
		Kind:       domain.AccountKind(m.Kind),
		Number:     m.Number,
		ArchivedAt: m.ArchivedAt,
		CreatedBy:  database.IDOrZero(m.CreatedBy),
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

type repo struct {
	client database.DatabaseClient
}

func New(client database.DatabaseClient) *repo {
	return &repo{client: client}
}

// FindAll returns every account, archived ones too, Tunai (the first, seeded) first
func (r *repo) FindAll(ctx context.Context) ([]domain.Account, error) {
	query, args, err := builder.Select(columns...).From(accountsTable).OrderBy("id").ToSql()
	if err != nil {
		return nil, errInternal(fmt.Errorf("error building find accounts query: %w", err))
	}

	var models []model
	if err := r.client.Select(ctx, &models, query, args...); err != nil {
		return nil, errInternal(fmt.Errorf("error finding accounts: %w", err))
	}

	accounts := make([]domain.Account, 0, len(models))
	for _, m := range models {
		accounts = append(accounts, m.toDomain())
	}
	return accounts, nil
}

// FindByID returns one account, or domain.ErrAccountNotFound
func (r *repo) FindByID(ctx context.Context, id int64) (domain.Account, error) {
	query, args, err := builder.Select(columns...).From(accountsTable).Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return domain.Account{}, errInternal(fmt.Errorf("error building find account query: %w", err))
	}

	var m model
	if err := r.client.Get(ctx, &m, query, args...); err != nil {
		if errors.Is(err, pkgdb.ErrNotFound) {
			return domain.Account{}, errNotFound()
		}
		return domain.Account{}, errInternal(fmt.Errorf("error finding account: %w", err))
	}
	return m.toDomain(), nil
}

// Create inserts an account and returns it; a taken name is CodeAlreadyExists
func (r *repo) Create(ctx context.Context, account domain.Account) (domain.Account, error) {
	now := time.Now().UTC()
	query, args, err := builder.
		Insert(accountsTable).
		Columns("name", "kind", "number", "created_by", "created_at", "updated_at").
		Values(account.Name, string(account.Kind), account.Number, database.NullableID(account.CreatedBy), now, now).
		Suffix("RETURNING id"). // LastInsertId is not supported on PostgreSQL
		ToSql()
	if err != nil {
		return domain.Account{}, errInternal(fmt.Errorf("error building create account query: %w", err))
	}

	var id int64
	if err := r.client.Get(ctx, &id, query, args...); err != nil {
		if errors.Is(err, pkgdb.ErrUniqueViolation) {
			return domain.Account{}, errNameTaken()
		}
		return domain.Account{}, errInternal(fmt.Errorf("error creating account: %w", err))
	}
	return r.FindByID(ctx, id)
}

// Update saves the name, number and archived_at, or returns domain.ErrAccountNotFound
func (r *repo) Update(ctx context.Context, account domain.Account) (domain.Account, error) {
	query, args, err := builder.
		Update(accountsTable).
		Set("name", account.Name).
		Set("number", account.Number).
		Set("archived_at", account.ArchivedAt).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"id": account.ID}).
		ToSql()
	if err != nil {
		return domain.Account{}, errInternal(fmt.Errorf("error building update account query: %w", err))
	}

	res, err := r.client.Exec(ctx, query, args...)
	if err != nil {
		if errors.Is(err, pkgdb.ErrUniqueViolation) {
			return domain.Account{}, errNameTaken()
		}
		return domain.Account{}, errInternal(fmt.Errorf("error updating account: %w", err))
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return domain.Account{}, errNotFound()
	}
	return r.FindByID(ctx, account.ID)
}

func errNotFound() *apperror.AppError {
	return apperror.New(apperror.CodeNotFound, domain.ErrAccountNotFound, "akun tidak ditemukan")
}

func errNameTaken() *apperror.AppError {
	return apperror.New(apperror.CodeAlreadyExists, errors.New("account name taken"), "nama akun sudah dipakai")
}

func errInternal(err error) *apperror.AppError {
	return apperror.New(apperror.CodeInternal, err, "gagal mengakses data akun")
}
