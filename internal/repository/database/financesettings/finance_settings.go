// Package financesettings stores the Keuangan settings as one JSON row in configurations
package financesettings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	pkgdb "github.com/faruq-rahmadani/masjid-digital/pkg/database"
)

const (
	configurationsTable = "configurations"
	tvSettingKey        = "finance_tv_setting"
)

var builder = sq.StatementBuilder.PlaceholderFormat(sq.Question)

type repo struct {
	client database.DatabaseClient
}

func New(client database.DatabaseClient) *repo {
	return &repo{client: client}
}

// GetTVSetting returns domain.ErrFinanceTVSettingNotFound until it is first saved
func (r *repo) GetTVSetting(ctx context.Context) (domain.FinanceTVSetting, error) {
	query, args, err := builder.Select("value").From(configurationsTable).Where(sq.Eq{"key": tvSettingKey}).ToSql()
	if err != nil {
		return domain.FinanceTVSetting{}, errInternal(fmt.Errorf("error building get finance tv setting query: %w", err))
	}

	var value string
	if err := r.client.Get(ctx, &value, query, args...); err != nil {
		if errors.Is(err, pkgdb.ErrNotFound) {
			return domain.FinanceTVSetting{}, apperror.New(apperror.CodeNotFound, domain.ErrFinanceTVSettingNotFound, "pengaturan kas di TV belum disimpan")
		}
		return domain.FinanceTVSetting{}, errInternal(fmt.Errorf("error getting finance tv setting: %w", err))
	}

	var setting domain.FinanceTVSetting
	if err := json.Unmarshal([]byte(value), &setting); err != nil {
		return domain.FinanceTVSetting{}, errInternal(fmt.Errorf("error decoding finance tv setting: %w", err))
	}
	return setting, nil
}

// UpdateTVSetting inserts or replaces the setting
func (r *repo) UpdateTVSetting(ctx context.Context, setting domain.FinanceTVSetting, updatedBy int64) error {
	encoded, err := json.Marshal(setting)
	if err != nil {
		return errInternal(fmt.Errorf("error encoding finance tv setting: %w", err))
	}
	query, args, err := builder.
		Insert(configurationsTable).
		Columns("key", "value", "updated_by", "updated_at").
		Values(tvSettingKey, string(encoded), database.NullableID(updatedBy), time.Now().UTC()).
		Suffix("ON CONFLICT (key) DO UPDATE SET value = excluded.value, updated_by = excluded.updated_by, updated_at = excluded.updated_at").
		ToSql()
	if err != nil {
		return errInternal(fmt.Errorf("error building upsert finance tv setting query: %w", err))
	}
	if _, err := r.client.Exec(ctx, query, args...); err != nil {
		return errInternal(fmt.Errorf("error saving finance tv setting: %w", err))
	}
	return nil
}

func errInternal(err error) *apperror.AppError {
	return apperror.New(apperror.CodeInternal, err, "gagal mengakses pengaturan keuangan")
}
