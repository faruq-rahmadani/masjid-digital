package categories

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

const categoriesTable = "categories"

// builder emits `?` placeholders; pkg/database rebinds them for the engine in use
var builder = sq.StatementBuilder.PlaceholderFormat(sq.Question)

var columns = []string{"id", "name", "type", "scope", "code", "archived_at", "created_by", "created_at", "updated_at"}

type model struct {
	ID         int64      `db:"id"`
	Name       string     `db:"name"`
	Type       string     `db:"type"`
	Scope      string     `db:"scope"`
	Code       *string    `db:"code"`
	ArchivedAt *time.Time `db:"archived_at"`
	CreatedBy  *int64     `db:"created_by"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
}

func (m model) toDomain() domain.Category {
	return domain.Category{
		ID:         m.ID,
		Name:       m.Name,
		Type:       domain.CategoryType(m.Type),
		Scope:      domain.CategoryScope(m.Scope),
		Code:       m.Code,
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

// FindAll returns every category, archived ones too, by scope, type and name
func (r *repo) FindAll(ctx context.Context) ([]domain.Category, error) {
	query, args, err := builder.Select(columns...).From(categoriesTable).OrderBy("scope", "type", "id").ToSql()
	if err != nil {
		return nil, errInternal(fmt.Errorf("error building find categories query: %w", err))
	}

	var models []model
	if err := r.client.Select(ctx, &models, query, args...); err != nil {
		return nil, errInternal(fmt.Errorf("error finding categories: %w", err))
	}

	categories := make([]domain.Category, 0, len(models))
	for _, m := range models {
		categories = append(categories, m.toDomain())
	}
	return categories, nil
}

// FindByID returns one category, or domain.ErrCategoryNotFound
func (r *repo) FindByID(ctx context.Context, id int64) (domain.Category, error) {
	return r.findOne(ctx, sq.Eq{"id": id})
}

// FindByCode returns the system category with the code, or domain.ErrCategoryNotFound
func (r *repo) FindByCode(ctx context.Context, code string) (domain.Category, error) {
	return r.findOne(ctx, sq.Eq{"code": code})
}

func (r *repo) findOne(ctx context.Context, where sq.Eq) (domain.Category, error) {
	query, args, err := builder.Select(columns...).From(categoriesTable).Where(where).ToSql()
	if err != nil {
		return domain.Category{}, errInternal(fmt.Errorf("error building find category query: %w", err))
	}

	var m model
	if err := r.client.Get(ctx, &m, query, args...); err != nil {
		if errors.Is(err, pkgdb.ErrNotFound) {
			return domain.Category{}, errNotFound()
		}
		return domain.Category{}, errInternal(fmt.Errorf("error finding category: %w", err))
	}
	return m.toDomain(), nil
}

// Create inserts a category and returns it; a name taken in its scope and type is CodeAlreadyExists
func (r *repo) Create(ctx context.Context, c domain.Category) (domain.Category, error) {
	now := time.Now().UTC()
	query, args, err := builder.
		Insert(categoriesTable).
		Columns("name", "type", "scope", "created_by", "created_at", "updated_at").
		Values(c.Name, string(c.Type), string(c.Scope), database.NullableID(c.CreatedBy), now, now).
		Suffix("RETURNING id"). // LastInsertId is not supported on PostgreSQL
		ToSql()
	if err != nil {
		return domain.Category{}, errInternal(fmt.Errorf("error building create category query: %w", err))
	}

	var id int64
	if err := r.client.Get(ctx, &id, query, args...); err != nil {
		if errors.Is(err, pkgdb.ErrUniqueViolation) {
			return domain.Category{}, errNameTaken()
		}
		return domain.Category{}, errInternal(fmt.Errorf("error creating category: %w", err))
	}
	return r.FindByID(ctx, id)
}

// Update saves the name and archived_at, or returns domain.ErrCategoryNotFound
func (r *repo) Update(ctx context.Context, c domain.Category) (domain.Category, error) {
	query, args, err := builder.
		Update(categoriesTable).
		Set("name", c.Name).
		Set("archived_at", c.ArchivedAt).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"id": c.ID}).
		ToSql()
	if err != nil {
		return domain.Category{}, errInternal(fmt.Errorf("error building update category query: %w", err))
	}

	res, err := r.client.Exec(ctx, query, args...)
	if err != nil {
		if errors.Is(err, pkgdb.ErrUniqueViolation) {
			return domain.Category{}, errNameTaken()
		}
		return domain.Category{}, errInternal(fmt.Errorf("error updating category: %w", err))
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return domain.Category{}, errNotFound()
	}
	return r.FindByID(ctx, c.ID)
}

// Delete removes a category, or returns domain.ErrCategoryNotFound
func (r *repo) Delete(ctx context.Context, id int64) error {
	query, args, err := builder.Delete(categoriesTable).Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return errInternal(fmt.Errorf("error building delete category query: %w", err))
	}

	res, err := r.client.Exec(ctx, query, args...)
	if err != nil {
		return errInternal(fmt.Errorf("error deleting category: %w", err))
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return errNotFound()
	}
	return nil
}

func errNotFound() *apperror.AppError {
	return apperror.New(apperror.CodeNotFound, domain.ErrCategoryNotFound, "kategori tidak ditemukan")
}

func errNameTaken() *apperror.AppError {
	return apperror.New(apperror.CodeAlreadyExists, errors.New("category name taken"), "kategori itu sudah ada")
}

func errInternal(err error) *apperror.AppError {
	return apperror.New(apperror.CodeInternal, err, "gagal mengakses data kategori")
}
