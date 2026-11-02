package screens

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

const screensTable = "screens"

// builder emits `?` placeholders; pkg/database rebinds them for the engine in use
var builder = sq.StatementBuilder.PlaceholderFormat(sq.Question)

type repo struct {
	client database.DatabaseClient
}

func New(client database.DatabaseClient) *repo {
	return &repo{client: client}
}

// FindAll returns every screen, blocked ones too, oldest first
func (r *repo) FindAll(ctx context.Context) ([]domain.Screen, error) {
	query, args, err := builder.
		Select(screenColumns...).
		From(screensTable).
		OrderBy("first_seen_at", "id").
		ToSql()
	if err != nil {
		return nil, errInternal(fmt.Errorf("error building find screens query: %w", err))
	}

	var models []screenModel
	if err := r.client.Select(ctx, &models, query, args...); err != nil {
		return nil, errInternal(fmt.Errorf("error finding screens: %w", err))
	}

	screens := make([]domain.Screen, 0, len(models))
	for _, m := range models {
		screens = append(screens, m.toDomain())
	}
	return screens, nil
}

// FindByID returns one screen, or domain.ErrScreenNotFound
func (r *repo) FindByID(ctx context.Context, id int64) (domain.Screen, error) {
	return r.findOne(ctx, sq.Eq{"id": id})
}

// FindByDeviceID returns the screen of a device, or domain.ErrScreenNotFound
func (r *repo) FindByDeviceID(ctx context.Context, deviceID string) (domain.Screen, error) {
	return r.findOne(ctx, sq.Eq{"device_id": deviceID})
}

func (r *repo) findOne(ctx context.Context, where sq.Eq) (domain.Screen, error) {
	query, args, err := builder.
		Select(screenColumns...).
		From(screensTable).
		Where(where).
		ToSql()
	if err != nil {
		return domain.Screen{}, errInternal(fmt.Errorf("error building find screen query: %w", err))
	}

	var m screenModel
	if err := r.client.Get(ctx, &m, query, args...); err != nil {
		if errors.Is(err, pkgdb.ErrNotFound) {
			return domain.Screen{}, errNotFound()
		}
		return domain.Screen{}, errInternal(fmt.Errorf("error finding screen: %w", err))
	}
	return m.toDomain(), nil
}

// Register inserts the device unless it already has a row (which it leaves as is), and returns
// the device's row
func (r *repo) Register(ctx context.Context, screen domain.Screen) (domain.Screen, error) {
	now := time.Now().UTC()

	query, args, err := builder.
		Insert(screensTable).
		Columns("device_id", "name_hint", "user_agent", "forced_mode", "first_seen_at", "updated_at").
		Values(screen.DeviceID, screen.NameHint, screen.UserAgent, string(domain.ScreenModeAuto), now, now).
		// same syntax on PostgreSQL and SQLite
		Suffix("ON CONFLICT (device_id) DO NOTHING").
		ToSql()
	if err != nil {
		return domain.Screen{}, errInternal(fmt.Errorf("error building register screen query: %w", err))
	}

	if _, err := r.client.Exec(ctx, query, args...); err != nil {
		return domain.Screen{}, errInternal(fmt.Errorf("error registering screen: %w", err))
	}
	return r.FindByDeviceID(ctx, screen.DeviceID)
}

// Update saves the name, forced mode and until, and blocked_at, or returns domain.ErrScreenNotFound
func (r *repo) Update(ctx context.Context, screen domain.Screen) (domain.Screen, error) {
	query, args, err := builder.
		Update(screensTable).
		Set("name", screen.Name).
		Set("forced_mode", string(screen.ForcedMode)).
		Set("forced_until", screen.ForcedUntil).
		Set("blocked_at", screen.BlockedAt).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"id": screen.ID}).
		ToSql()
	if err != nil {
		return domain.Screen{}, errInternal(fmt.Errorf("error building update screen query: %w", err))
	}

	res, err := r.client.Exec(ctx, query, args...)
	if err != nil {
		return domain.Screen{}, errInternal(fmt.Errorf("error updating screen: %w", err))
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return domain.Screen{}, errNotFound()
	}
	return r.FindByID(ctx, screen.ID)
}
