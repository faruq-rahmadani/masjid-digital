package fridayduties

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database"
)

const dutiesTable = "friday_duties"

// builder emits `?` placeholders; pkg/database rebinds them for the engine in use
var builder = sq.StatementBuilder.PlaceholderFormat(sq.Question)

type FridayDutyDBRepository struct {
	client database.DatabaseClient
}

func NewFridayDutyDBRepository(client database.DatabaseClient) *FridayDutyDBRepository {
	return &FridayDutyDBRepository{client: client}
}

// Save inserts the petugas of a Friday, or replaces them, and returns the stored row
func (r *FridayDutyDBRepository) Save(ctx context.Context, duty domain.FridayDuty) (domain.FridayDuty, error) {
	now := time.Now().UTC()

	query, args, err := builder.
		Insert(dutiesTable).
		Columns("date", "khatib", "imam", "muadzin", "theme", "updated_by", "created_at", "updated_at").
		Values(duty.Date, duty.Khatib, duty.Imam, duty.Muadzin, duty.Theme, nullableID(duty.UpdatedBy), now, now).
		// created_at stays from the first save; same syntax on PostgreSQL and SQLite
		Suffix("ON CONFLICT (date) DO UPDATE SET khatib = excluded.khatib, imam = excluded.imam, muadzin = excluded.muadzin, theme = excluded.theme, updated_by = excluded.updated_by, updated_at = excluded.updated_at").
		ToSql()
	if err != nil {
		return domain.FridayDuty{}, errInternal(fmt.Errorf("error building save friday duty query: %w", err))
	}

	if _, err := r.client.Exec(ctx, query, args...); err != nil {
		return domain.FridayDuty{}, errInternal(fmt.Errorf("error saving friday duty: %w", err))
	}

	duties, err := r.FindBetween(ctx, duty.Date, duty.Date)
	if err != nil {
		return domain.FridayDuty{}, err
	}
	if len(duties) == 0 {
		return domain.FridayDuty{}, errInternal(fmt.Errorf("friday duty %s missing after save", duty.Date))
	}
	return duties[0], nil
}

// FindBetween returns the petugas Jumat dated from..to inclusive (YYYY-MM-DD), by date
func (r *FridayDutyDBRepository) FindBetween(ctx context.Context, from, to string) ([]domain.FridayDuty, error) {
	query, args, err := builder.
		Select(dutyColumns...).
		From(dutiesTable).
		// dates are zero-padded text, so comparing them as strings orders them by date
		Where(sq.And{sq.GtOrEq{"date": from}, sq.LtOrEq{"date": to}}).
		OrderBy("date").
		ToSql()
	if err != nil {
		return nil, errInternal(fmt.Errorf("error building find friday duties query: %w", err))
	}

	var models []dutyModel
	if err := r.client.Select(ctx, &models, query, args...); err != nil {
		return nil, errInternal(fmt.Errorf("error finding friday duties: %w", err))
	}

	duties := make([]domain.FridayDuty, 0, len(models))
	for _, m := range models {
		duties = append(duties, m.toDomain())
	}
	return duties, nil
}

// Delete removes the petugas of a Friday, or returns domain.ErrFridayDutyNotFound
func (r *FridayDutyDBRepository) Delete(ctx context.Context, date string) error {
	query, args, err := builder.
		Delete(dutiesTable).
		Where(sq.Eq{"date": date}).
		ToSql()
	if err != nil {
		return errInternal(fmt.Errorf("error building delete friday duty query: %w", err))
	}

	res, err := r.client.Exec(ctx, query, args...)
	if err != nil {
		return errInternal(fmt.Errorf("error deleting friday duty: %w", err))
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return errNotFound()
	}
	return nil
}

// nullableID stores 0 (nobody) as NULL
func nullableID(id int64) *int64 {
	if id == 0 {
		return nil
	}
	return &id
}
