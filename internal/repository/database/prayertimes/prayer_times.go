package prayertimes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database"
	pkgdb "github.com/faruq-rahmadani/masjid-digital/pkg/database"
)

const (
	// configurationsTable holds one row per group of settings as JSON (see migration 00004)
	configurationsTable = "configurations"
	settingsKey         = "prayer_settings"
	overridesTable      = "prayer_time_overrides"
)

// builder emits `?` placeholders; pkg/database rebinds them for the engine in use
var builder = sq.StatementBuilder.PlaceholderFormat(sq.Question)

type PrayerTimeDBRepository struct {
	client database.DatabaseClient
}

func NewPrayerTimeDBRepository(client database.DatabaseClient) *PrayerTimeDBRepository {
	return &PrayerTimeDBRepository{client: client}
}

// GetSettings returns the saved settings, or domain.ErrPrayerSettingsNotFound when never saved
func (r *PrayerTimeDBRepository) GetSettings(ctx context.Context) (domain.PrayerSettings, error) {
	query, args, err := builder.
		Select("value").
		From(configurationsTable).
		Where(sq.Eq{"key": settingsKey}).
		ToSql()
	if err != nil {
		return domain.PrayerSettings{}, errInternal(fmt.Errorf("error building get prayer settings query: %w", err))
	}

	var value string
	if err := r.client.Get(ctx, &value, query, args...); err != nil {
		if errors.Is(err, pkgdb.ErrNotFound) {
			return domain.PrayerSettings{}, errSettingsNotFound()
		}
		return domain.PrayerSettings{}, errInternal(fmt.Errorf("error getting prayer settings: %w", err))
	}

	var v settingsValue
	if err := json.Unmarshal([]byte(value), &v); err != nil {
		return domain.PrayerSettings{}, errInternal(fmt.Errorf("error decoding prayer settings: %w", err))
	}
	return v.toDomain(), nil
}

// UpdateSettings inserts the settings, or replaces the saved ones, by updatedBy (0 is stored as NULL)
func (r *PrayerTimeDBRepository) UpdateSettings(ctx context.Context, settings domain.PrayerSettings, updatedBy int64) error {
	encoded, err := json.Marshal(newSettingsValue(settings))
	if err != nil {
		return errInternal(fmt.Errorf("error encoding prayer settings: %w", err))
	}

	query, args, err := builder.
		Insert(configurationsTable).
		Columns("key", "value", "updated_by", "updated_at").
		Values(settingsKey, string(encoded), nullableID(updatedBy), time.Now().UTC()).
		// same syntax on PostgreSQL and SQLite
		Suffix("ON CONFLICT (key) DO UPDATE SET value = excluded.value, updated_by = excluded.updated_by, updated_at = excluded.updated_at").
		ToSql()
	if err != nil {
		return errInternal(fmt.Errorf("error building upsert prayer settings query: %w", err))
	}

	if _, err := r.client.Exec(ctx, query, args...); err != nil {
		return errInternal(fmt.Errorf("error saving prayer settings: %w", err))
	}
	return nil
}

// CreateOverride inserts a date override and returns it with its ID and timestamps. It returns
// domain.ErrPrayerTimeOverrideExists when that prayer already has one on that date.
func (r *PrayerTimeDBRepository) CreateOverride(ctx context.Context, o domain.PrayerTimeOverride) (domain.PrayerTimeOverride, error) {
	now := time.Now().UTC()

	query, args, err := builder.
		Insert(overridesTable).
		Columns("date", "prayer", "time", "remarks", "created_by", "created_at", "updated_at").
		Values(o.Date, string(o.Prayer), o.Time, o.Remarks, nullableID(o.CreatedBy), now, now).
		Suffix("RETURNING id"). // LastInsertId is not supported on PostgreSQL
		ToSql()
	if err != nil {
		return domain.PrayerTimeOverride{}, errInternal(fmt.Errorf("error building create override query: %w", err))
	}

	if err := r.client.Get(ctx, &o.ID, query, args...); err != nil {
		if errors.Is(err, pkgdb.ErrUniqueViolation) {
			return domain.PrayerTimeOverride{}, errOverrideExists()
		}
		return domain.PrayerTimeOverride{}, errInternal(fmt.Errorf("error creating override: %w", err))
	}

	o.CreatedAt, o.UpdatedAt = now, now
	return o, nil
}

// FindOverridesBetween returns the overrides dated from..to inclusive (YYYY-MM-DD), by date then
// prayer
func (r *PrayerTimeDBRepository) FindOverridesBetween(ctx context.Context, from, to string) ([]domain.PrayerTimeOverride, error) {
	query, args, err := builder.
		Select(overrideColumns...).
		From(overridesTable).
		// dates are zero-padded text, so comparing them as strings orders them by date
		Where(sq.And{sq.GtOrEq{"date": from}, sq.LtOrEq{"date": to}}).
		OrderBy("date", "prayer").
		ToSql()
	if err != nil {
		return nil, errInternal(fmt.Errorf("error building find overrides query: %w", err))
	}

	var models []overrideModel
	if err := r.client.Select(ctx, &models, query, args...); err != nil {
		return nil, errInternal(fmt.Errorf("error finding overrides: %w", err))
	}

	overrides := make([]domain.PrayerTimeOverride, 0, len(models))
	for _, m := range models {
		overrides = append(overrides, m.toDomain())
	}
	return overrides, nil
}

// DeleteOverride removes a date override, or returns domain.ErrPrayerTimeOverrideNotFound
func (r *PrayerTimeDBRepository) DeleteOverride(ctx context.Context, id int64) error {
	query, args, err := builder.
		Delete(overridesTable).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return errInternal(fmt.Errorf("error building delete override query: %w", err))
	}

	res, err := r.client.Exec(ctx, query, args...)
	if err != nil {
		return errInternal(fmt.Errorf("error deleting override: %w", err))
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return errOverrideNotFound()
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
