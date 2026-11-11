package pockets

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

const pocketsTable = "pockets"

// builder emits `?` placeholders; pkg/database rebinds them for the engine in use
var builder = sq.StatementBuilder.PlaceholderFormat(sq.Question)

var columns = []string{
	"id", "kind", "name", "purpose", "target_amount", "surplus_plan", "show_on_tv", "status", "created_by", "created_at", "updated_at",
}

type model struct {
	ID           int64     `db:"id"`
	Kind         string    `db:"kind"`
	Name         string    `db:"name"`
	Purpose      string    `db:"purpose"`
	TargetAmount *int64    `db:"target_amount"`
	SurplusPlan  *string   `db:"surplus_plan"`
	ShowOnTV     bool      `db:"show_on_tv"`
	Status       string    `db:"status"`
	CreatedBy    *int64    `db:"created_by"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (m model) toDomain() domain.Pocket {
	var plan *domain.SurplusPlan
	if m.SurplusPlan != nil {
		p := domain.SurplusPlan(*m.SurplusPlan)
		plan = &p
	}
	return domain.Pocket{
		ID:           m.ID,
		Kind:         domain.PocketKind(m.Kind),
		Name:         m.Name,
		Purpose:      m.Purpose,
		TargetAmount: m.TargetAmount,
		SurplusPlan:  plan,
		ShowOnTV:     m.ShowOnTV,
		Status:       domain.PocketStatus(m.Status),
		CreatedBy:    database.IDOrZero(m.CreatedBy),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func surplusPlan(plan *domain.SurplusPlan) *string {
	if plan == nil {
		return nil
	}
	s := string(*plan)
	return &s
}

type repo struct {
	client database.DatabaseClient
}

func New(client database.DatabaseClient) *repo {
	return &repo{client: client}
}

// FindAll returns every pocket, archived ones too; the four seeded built-in ones come first
func (r *repo) FindAll(ctx context.Context) ([]domain.Pocket, error) {
	query, args, err := builder.Select(columns...).From(pocketsTable).OrderBy("id").ToSql()
	if err != nil {
		return nil, errInternal(fmt.Errorf("error building find pockets query: %w", err))
	}

	var models []model
	if err := r.client.Select(ctx, &models, query, args...); err != nil {
		return nil, errInternal(fmt.Errorf("error finding pockets: %w", err))
	}

	pockets := make([]domain.Pocket, 0, len(models))
	for _, m := range models {
		pockets = append(pockets, m.toDomain())
	}
	return pockets, nil
}

// FindByID returns one pocket, or domain.ErrPocketNotFound
func (r *repo) FindByID(ctx context.Context, id int64) (domain.Pocket, error) {
	query, args, err := builder.Select(columns...).From(pocketsTable).Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return domain.Pocket{}, errInternal(fmt.Errorf("error building find pocket query: %w", err))
	}

	var m model
	if err := r.client.Get(ctx, &m, query, args...); err != nil {
		if errors.Is(err, pkgdb.ErrNotFound) {
			return domain.Pocket{}, errNotFound()
		}
		return domain.Pocket{}, errInternal(fmt.Errorf("error finding pocket: %w", err))
	}
	return m.toDomain(), nil
}

// Create inserts a pocket and returns it; a taken name is CodeAlreadyExists
func (r *repo) Create(ctx context.Context, p domain.Pocket) (domain.Pocket, error) {
	now := time.Now().UTC()
	query, args, err := builder.
		Insert(pocketsTable).
		Columns("kind", "name", "purpose", "target_amount", "surplus_plan", "show_on_tv", "status", "created_by", "created_at", "updated_at").
		Values(string(p.Kind), p.Name, p.Purpose, p.TargetAmount, surplusPlan(p.SurplusPlan), p.ShowOnTV, string(p.Status), database.NullableID(p.CreatedBy), now, now).
		Suffix("RETURNING id"). // LastInsertId is not supported on PostgreSQL
		ToSql()
	if err != nil {
		return domain.Pocket{}, errInternal(fmt.Errorf("error building create pocket query: %w", err))
	}

	var id int64
	if err := r.client.Get(ctx, &id, query, args...); err != nil {
		if errors.Is(err, pkgdb.ErrUniqueViolation) {
			return domain.Pocket{}, errNameTaken()
		}
		return domain.Pocket{}, errInternal(fmt.Errorf("error creating pocket: %w", err))
	}
	return r.FindByID(ctx, id)
}

// Update saves everything but the kind, or returns domain.ErrPocketNotFound
func (r *repo) Update(ctx context.Context, p domain.Pocket) (domain.Pocket, error) {
	query, args, err := builder.
		Update(pocketsTable).
		Set("name", p.Name).
		Set("purpose", p.Purpose).
		Set("target_amount", p.TargetAmount).
		Set("surplus_plan", surplusPlan(p.SurplusPlan)).
		Set("show_on_tv", p.ShowOnTV).
		Set("status", string(p.Status)).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"id": p.ID}).
		ToSql()
	if err != nil {
		return domain.Pocket{}, errInternal(fmt.Errorf("error building update pocket query: %w", err))
	}

	res, err := r.client.Exec(ctx, query, args...)
	if err != nil {
		if errors.Is(err, pkgdb.ErrUniqueViolation) {
			return domain.Pocket{}, errNameTaken()
		}
		return domain.Pocket{}, errInternal(fmt.Errorf("error updating pocket: %w", err))
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return domain.Pocket{}, errNotFound()
	}
	return r.FindByID(ctx, p.ID)
}

func errNotFound() *apperror.AppError {
	return apperror.New(apperror.CodeNotFound, domain.ErrPocketNotFound, "kas tidak ditemukan")
}

func errNameTaken() *apperror.AppError {
	return apperror.New(apperror.CodeAlreadyExists, errors.New("pocket name taken"), "nama kas sudah dipakai")
}

func errInternal(err error) *apperror.AppError {
	return apperror.New(apperror.CodeInternal, err, "gagal mengakses data kas")
}
