package runningtexts

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
	runningTextsTable = "running_texts"
	// configurationsTable holds one row per group of settings as JSON (see migration 00004)
	configurationsTable = "configurations"
	settingKey          = "running_text_setting"
)

// builder emits `?` placeholders; pkg/database rebinds them for the engine in use
var builder = sq.StatementBuilder.PlaceholderFormat(sq.Question)

type repo struct {
	client database.DatabaseClient
}

func New(client database.DatabaseClient) *repo {
	return &repo{client: client}
}

// selectQuery selects running texts with their creator's name
func selectQuery() sq.SelectBuilder {
	return builder.
		Select(runningTextColumns...).
		From(runningTextsTable + " rt").
		LeftJoin("users u ON u.id = rt.created_by")
}

// FindAll returns every running text in display order
func (r *repo) FindAll(ctx context.Context) ([]domain.RunningText, error) {
	query, args, err := selectQuery().
		OrderBy("rt.position", "rt.id").
		ToSql()
	if err != nil {
		return nil, errInternal(fmt.Errorf("error building find running texts query: %w", err))
	}

	var models []runningTextModel
	if err := r.client.Select(ctx, &models, query, args...); err != nil {
		return nil, errInternal(fmt.Errorf("error finding running texts: %w", err))
	}

	texts := make([]domain.RunningText, 0, len(models))
	for _, m := range models {
		texts = append(texts, m.toDomain())
	}
	return texts, nil
}

// Create inserts an active running text after the last one and returns it
func (r *repo) Create(ctx context.Context, text domain.RunningText) (domain.RunningText, error) {
	now := time.Now().UTC()

	query, args, err := builder.
		Insert(runningTextsTable).
		Columns("text", "active", "position", "created_by", "created_at", "updated_at").
		Values(text.Text, true, sq.Expr("(SELECT COALESCE(MAX(position) + 1, 0) FROM "+runningTextsTable+")"), nullableID(text.CreatedBy), now, now).
		Suffix("RETURNING id"). // LastInsertId is not supported on PostgreSQL
		ToSql()
	if err != nil {
		return domain.RunningText{}, errInternal(fmt.Errorf("error building create running text query: %w", err))
	}

	var id int64
	if err := r.client.Get(ctx, &id, query, args...); err != nil {
		return domain.RunningText{}, errInternal(fmt.Errorf("error creating running text: %w", err))
	}
	return r.findByID(ctx, id)
}

// Update saves the text and whether it is active, or returns domain.ErrRunningTextNotFound.
// Turning a text off stamps deactivated_at; turning it on clears it.
func (r *repo) Update(ctx context.Context, text domain.RunningText) (domain.RunningText, error) {
	now := time.Now().UTC()

	query, args, err := builder.
		Update(runningTextsTable).
		Set("text", text.Text).
		Set("active", text.Active).
		// keep the first stamp when an inactive text is only edited
		Set("deactivated_at", sq.Expr("CASE WHEN ? THEN NULL ELSE COALESCE(deactivated_at, ?) END", text.Active, now)).
		Set("updated_at", now).
		Where(sq.Eq{"id": text.ID}).
		ToSql()
	if err != nil {
		return domain.RunningText{}, errInternal(fmt.Errorf("error building update running text query: %w", err))
	}

	res, err := r.client.Exec(ctx, query, args...)
	if err != nil {
		return domain.RunningText{}, errInternal(fmt.Errorf("error updating running text: %w", err))
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return domain.RunningText{}, errNotFound()
	}
	return r.findByID(ctx, text.ID)
}

// findByID returns one running text with its creator's name, or domain.ErrRunningTextNotFound
func (r *repo) findByID(ctx context.Context, id int64) (domain.RunningText, error) {
	query, args, err := selectQuery().
		Where(sq.Eq{"rt.id": id}).
		ToSql()
	if err != nil {
		return domain.RunningText{}, errInternal(fmt.Errorf("error building find running text query: %w", err))
	}

	var m runningTextModel
	if err := r.client.Get(ctx, &m, query, args...); err != nil {
		if errors.Is(err, pkgdb.ErrNotFound) {
			return domain.RunningText{}, errNotFound()
		}
		return domain.RunningText{}, errInternal(fmt.Errorf("error finding running text: %w", err))
	}
	return m.toDomain(), nil
}

// Delete removes a running text, or returns domain.ErrRunningTextNotFound
func (r *repo) Delete(ctx context.Context, id int64) error {
	query, args, err := builder.
		Delete(runningTextsTable).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return errInternal(fmt.Errorf("error building delete running text query: %w", err))
	}

	res, err := r.client.Exec(ctx, query, args...)
	if err != nil {
		return errInternal(fmt.Errorf("error deleting running text: %w", err))
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return errNotFound()
	}
	return nil
}

// Reorder sets the position of each ID to its index in ids, in one statement so the order is
// never half saved
func (r *repo) Reorder(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	position := sq.Case("id")
	for i, id := range ids {
		// the cast types the parameter, which PostgreSQL would otherwise read as text
		position = position.When(sq.Expr("?", id), sq.Expr("CAST(? AS INTEGER)", i))
	}

	query, args, err := builder.
		Update(runningTextsTable).
		Set("position", position).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"id": ids}).
		ToSql()
	if err != nil {
		return errInternal(fmt.Errorf("error building reorder running texts query: %w", err))
	}

	if _, err := r.client.Exec(ctx, query, args...); err != nil {
		return errInternal(fmt.Errorf("error reordering running texts: %w", err))
	}
	return nil
}

// GetSetting returns the saved setting, or domain.ErrRunningTextSettingNotFound when never saved
func (r *repo) GetSetting(ctx context.Context) (domain.RunningTextSetting, error) {
	query, args, err := builder.
		Select("value").
		From(configurationsTable).
		Where(sq.Eq{"key": settingKey}).
		ToSql()
	if err != nil {
		return domain.RunningTextSetting{}, errInternal(fmt.Errorf("error building get running text setting query: %w", err))
	}

	var value string
	if err := r.client.Get(ctx, &value, query, args...); err != nil {
		if errors.Is(err, pkgdb.ErrNotFound) {
			return domain.RunningTextSetting{}, errSettingNotFound()
		}
		return domain.RunningTextSetting{}, errInternal(fmt.Errorf("error getting running text setting: %w", err))
	}

	var v settingValue
	if err := json.Unmarshal([]byte(value), &v); err != nil {
		return domain.RunningTextSetting{}, errInternal(fmt.Errorf("error decoding running text setting: %w", err))
	}
	return domain.RunningTextSetting{Speed: domain.ScrollSpeed(v.Speed)}, nil
}

// UpdateSetting inserts the setting, or replaces the saved one, by updatedBy (0 is stored as NULL)
func (r *repo) UpdateSetting(ctx context.Context, setting domain.RunningTextSetting, updatedBy int64) error {
	encoded, err := json.Marshal(settingValue{Speed: string(setting.Speed)})
	if err != nil {
		return errInternal(fmt.Errorf("error encoding running text setting: %w", err))
	}

	query, args, err := builder.
		Insert(configurationsTable).
		Columns("key", "value", "updated_by", "updated_at").
		Values(settingKey, string(encoded), nullableID(updatedBy), time.Now().UTC()).
		// same syntax on PostgreSQL and SQLite
		Suffix("ON CONFLICT (key) DO UPDATE SET value = excluded.value, updated_by = excluded.updated_by, updated_at = excluded.updated_at").
		ToSql()
	if err != nil {
		return errInternal(fmt.Errorf("error building upsert running text setting query: %w", err))
	}

	if _, err := r.client.Exec(ctx, query, args...); err != nil {
		return errInternal(fmt.Errorf("error saving running text setting: %w", err))
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
