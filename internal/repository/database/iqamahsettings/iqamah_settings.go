package iqamahsettings

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
	settingsKey         = "iqamah_settings"
)

// builder emits `?` placeholders; pkg/database rebinds them for the engine in use
var builder = sq.StatementBuilder.PlaceholderFormat(sq.Question)

type IqamahDBRepository struct {
	client database.DatabaseClient
}

func NewIqamahDBRepository(client database.DatabaseClient) *IqamahDBRepository {
	return &IqamahDBRepository{client: client}
}

// GetSettings returns the saved settings, or domain.ErrIqamahSettingsNotFound when never saved
func (r *IqamahDBRepository) GetSettings(ctx context.Context) (domain.IqamahSettings, error) {
	query, args, err := builder.
		Select("value").
		From(configurationsTable).
		Where(sq.Eq{"key": settingsKey}).
		ToSql()
	if err != nil {
		return domain.IqamahSettings{}, errInternal(fmt.Errorf("error building get iqamah settings query: %w", err))
	}

	var value string
	if err := r.client.Get(ctx, &value, query, args...); err != nil {
		if errors.Is(err, pkgdb.ErrNotFound) {
			return domain.IqamahSettings{}, errSettingsNotFound()
		}
		return domain.IqamahSettings{}, errInternal(fmt.Errorf("error getting iqamah settings: %w", err))
	}

	var v settingsValue
	if err := json.Unmarshal([]byte(value), &v); err != nil {
		return domain.IqamahSettings{}, errInternal(fmt.Errorf("error decoding iqamah settings: %w", err))
	}
	return v.toDomain(), nil
}

// UpdateSettings inserts the settings, or replaces the saved ones, by updatedBy (0 is stored as NULL)
func (r *IqamahDBRepository) UpdateSettings(ctx context.Context, settings domain.IqamahSettings, updatedBy int64) error {
	encoded, err := json.Marshal(newSettingsValue(settings))
	if err != nil {
		return errInternal(fmt.Errorf("error encoding iqamah settings: %w", err))
	}

	query, args, err := builder.
		Insert(configurationsTable).
		Columns("key", "value", "updated_by", "updated_at").
		Values(settingsKey, string(encoded), nullableID(updatedBy), time.Now().UTC()).
		// same syntax on PostgreSQL and SQLite
		Suffix("ON CONFLICT (key) DO UPDATE SET value = excluded.value, updated_by = excluded.updated_by, updated_at = excluded.updated_at").
		ToSql()
	if err != nil {
		return errInternal(fmt.Errorf("error building upsert iqamah settings query: %w", err))
	}

	if _, err := r.client.Exec(ctx, query, args...); err != nil {
		return errInternal(fmt.Errorf("error saving iqamah settings: %w", err))
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
