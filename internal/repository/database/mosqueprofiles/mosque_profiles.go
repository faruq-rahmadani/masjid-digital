package mosqueprofiles

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

// configurationsTable holds one row per setting, keyed by settingKey, with the
// setting as JSON. Other settings (iqamah, TV display, ...) keep their own keys there.
const configurationsTable = "configurations"

// settingKey names a row of the configurations table
type settingKey string

const (
	keyProfile settingKey = "mosque_profile"
	// the logo has its own row so reading the profile does not carry the image
	keyLogo settingKey = "mosque_logo"
)

// builder emits `?` placeholders; pkg/database rebinds them for the engine in use
var builder = sq.StatementBuilder.PlaceholderFormat(sq.Question)

type MosqueProfileDBRepository struct {
	client database.DatabaseClient
}

func NewMosqueProfileDBRepository(client database.DatabaseClient) *MosqueProfileDBRepository {
	return &MosqueProfileDBRepository{client: client}
}

// Get returns the mosque's profile, or domain.ErrMosqueProfileNotFound when never saved
func (r *MosqueProfileDBRepository) Get(ctx context.Context) (domain.MosqueProfile, error) {
	var v profileValue
	if err := r.get(ctx, keyProfile, &v); err != nil {
		if errors.Is(err, pkgdb.ErrNotFound) {
			return domain.MosqueProfile{}, errProfileNotFound()
		}
		return domain.MosqueProfile{}, errInternal(err)
	}

	return v.toDomain(), nil
}

// Update inserts the profile, or replaces the saved one, by updatedBy
func (r *MosqueProfileDBRepository) Update(ctx context.Context, profile domain.MosqueProfile, updatedBy int64) error {
	if err := r.upsert(ctx, keyProfile, newProfileValue(profile), updatedBy); err != nil {
		return errInternal(err)
	}
	return nil
}

// GetLogo returns the mosque's logo, or domain.ErrMosqueLogoNotFound when none was uploaded
func (r *MosqueProfileDBRepository) GetLogo(ctx context.Context) (domain.MosqueLogo, error) {
	var v logoValue
	if err := r.get(ctx, keyLogo, &v); err != nil {
		if errors.Is(err, pkgdb.ErrNotFound) {
			return domain.MosqueLogo{}, errLogoNotFound()
		}
		return domain.MosqueLogo{}, errInternal(err)
	}

	return v.toDomain(), nil
}

// UpdateLogo inserts the logo, or replaces the saved one, by updatedBy
func (r *MosqueProfileDBRepository) UpdateLogo(ctx context.Context, logo domain.MosqueLogo, updatedBy int64) error {
	if err := r.upsert(ctx, keyLogo, newLogoValue(logo), updatedBy); err != nil {
		return errInternal(err)
	}
	return nil
}

// DeleteLogo removes the logo. Deleting a missing logo is not an error.
func (r *MosqueProfileDBRepository) DeleteLogo(ctx context.Context) error {
	query, args, err := builder.
		Delete(configurationsTable).
		Where(sq.Eq{"key": string(keyLogo)}).
		ToSql()
	if err != nil {
		return errInternal(fmt.Errorf("error building delete logo query: %w", err))
	}

	if _, err := r.client.Exec(ctx, query, args...); err != nil {
		return errInternal(fmt.Errorf("error deleting logo: %w", err))
	}

	return nil
}

// get decodes the JSON stored under key into dest. It returns pkgdb.ErrNotFound
// when nothing is stored there yet.
func (r *MosqueProfileDBRepository) get(ctx context.Context, key settingKey, dest any) error {
	query, args, err := builder.
		Select("value").
		From(configurationsTable).
		Where(sq.Eq{"key": string(key)}).
		ToSql()
	if err != nil {
		return fmt.Errorf("error building get %s query: %w", key, err)
	}

	var value string
	if err := r.client.Get(ctx, &value, query, args...); err != nil {
		return fmt.Errorf("error getting %s: %w", key, err)
	}

	if err := json.Unmarshal([]byte(value), dest); err != nil {
		return fmt.Errorf("error decoding %s: %w", key, err)
	}

	return nil
}

// upsert stores value as JSON under key, inserting the row or replacing the
// one there. updatedBy 0 (nobody) is stored as NULL.
func (r *MosqueProfileDBRepository) upsert(ctx context.Context, key settingKey, value any, updatedBy int64) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("error encoding %s: %w", key, err)
	}

	var by *int64
	if updatedBy != 0 {
		by = &updatedBy
	}

	query, args, err := builder.
		Insert(configurationsTable).
		Columns("key", "value", "updated_by", "updated_at").
		Values(string(key), string(encoded), by, time.Now().UTC()).
		// same syntax on PostgreSQL and SQLite
		Suffix("ON CONFLICT (key) DO UPDATE SET value = excluded.value, updated_by = excluded.updated_by, updated_at = excluded.updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("error building upsert %s query: %w", key, err)
	}

	if _, err := r.client.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("error saving %s: %w", key, err)
	}

	return nil
}
