package slides

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

const slidesTable = "slides"

// builder emits `?` placeholders; pkg/database rebinds them for the engine in use
var builder = sq.StatementBuilder.PlaceholderFormat(sq.Question)

type repo struct {
	client database.DatabaseClient
}

func New(client database.DatabaseClient) *repo {
	return &repo{client: client}
}

// selectQuery selects slides, without their images, with their creator's name
func selectQuery() sq.SelectBuilder {
	return builder.
		Select(slideColumns...).
		From(slidesTable + " s").
		LeftJoin("users u ON u.id = s.created_by")
}

// FindAll returns every slide in display order
func (r *repo) FindAll(ctx context.Context) ([]domain.Slide, error) {
	query, args, err := selectQuery().
		OrderBy("s.position", "s.id").
		ToSql()
	if err != nil {
		return nil, errInternal(fmt.Errorf("error building find slides query: %w", err))
	}

	var models []slideModel
	if err := r.client.Select(ctx, &models, query, args...); err != nil {
		return nil, errInternal(fmt.Errorf("error finding slides: %w", err))
	}

	slides := make([]domain.Slide, 0, len(models))
	for _, m := range models {
		slides = append(slides, m.toDomain())
	}
	return slides, nil
}

// Create inserts an active slide with its image after the last one and returns it
func (r *repo) Create(ctx context.Context, slide domain.Slide, image domain.SlideImage) (domain.Slide, error) {
	now := time.Now().UTC()

	query, args, err := builder.
		Insert(slidesTable).
		Columns("title", "caption", "duration_seconds", "show_caption", "active", "position", "created_by", "created_at", "updated_at", "image_type", "image_data").
		Values(slide.Title, slide.Caption, slide.DurationSeconds, slide.ShowCaption, true, sq.Expr("(SELECT COALESCE(MAX(position) + 1, 0) FROM "+slidesTable+")"), nullableID(slide.CreatedBy), now, now, image.ContentType, image.Data).
		Suffix("RETURNING id"). // LastInsertId is not supported on PostgreSQL
		ToSql()
	if err != nil {
		return domain.Slide{}, errInternal(fmt.Errorf("error building create slide query: %w", err))
	}

	var id int64
	if err := r.client.Get(ctx, &id, query, args...); err != nil {
		return domain.Slide{}, errInternal(fmt.Errorf("error creating slide: %w", err))
	}
	return r.findByID(ctx, id)
}

// Update saves the title, caption, duration, whether the caption shows and whether the slide is
// active, or returns
// domain.ErrSlideNotFound
func (r *repo) Update(ctx context.Context, slide domain.Slide) (domain.Slide, error) {
	query, args, err := builder.
		Update(slidesTable).
		Set("title", slide.Title).
		Set("caption", slide.Caption).
		Set("duration_seconds", slide.DurationSeconds).
		Set("show_caption", slide.ShowCaption).
		Set("active", slide.Active).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"id": slide.ID}).
		ToSql()
	if err != nil {
		return domain.Slide{}, errInternal(fmt.Errorf("error building update slide query: %w", err))
	}

	res, err := r.client.Exec(ctx, query, args...)
	if err != nil {
		return domain.Slide{}, errInternal(fmt.Errorf("error updating slide: %w", err))
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return domain.Slide{}, errNotFound()
	}
	return r.findByID(ctx, slide.ID)
}

// findByID returns one slide with its creator's name, or domain.ErrSlideNotFound
func (r *repo) findByID(ctx context.Context, id int64) (domain.Slide, error) {
	query, args, err := selectQuery().
		Where(sq.Eq{"s.id": id}).
		ToSql()
	if err != nil {
		return domain.Slide{}, errInternal(fmt.Errorf("error building find slide query: %w", err))
	}

	var m slideModel
	if err := r.client.Get(ctx, &m, query, args...); err != nil {
		if errors.Is(err, pkgdb.ErrNotFound) {
			return domain.Slide{}, errNotFound()
		}
		return domain.Slide{}, errInternal(fmt.Errorf("error finding slide: %w", err))
	}
	return m.toDomain(), nil
}

// GetImage returns a slide's image, or domain.ErrSlideNotFound
func (r *repo) GetImage(ctx context.Context, id int64) (domain.SlideImage, error) {
	query, args, err := builder.
		Select("image_type", "image_data").
		From(slidesTable).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return domain.SlideImage{}, errInternal(fmt.Errorf("error building get slide image query: %w", err))
	}

	var m imageModel
	if err := r.client.Get(ctx, &m, query, args...); err != nil {
		if errors.Is(err, pkgdb.ErrNotFound) {
			return domain.SlideImage{}, errNotFound()
		}
		return domain.SlideImage{}, errInternal(fmt.Errorf("error getting slide image: %w", err))
	}
	return domain.SlideImage{ContentType: m.ContentType, Data: m.Data}, nil
}

// Delete removes a slide and its image, or returns domain.ErrSlideNotFound
func (r *repo) Delete(ctx context.Context, id int64) error {
	query, args, err := builder.
		Delete(slidesTable).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return errInternal(fmt.Errorf("error building delete slide query: %w", err))
	}

	res, err := r.client.Exec(ctx, query, args...)
	if err != nil {
		return errInternal(fmt.Errorf("error deleting slide: %w", err))
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
		Update(slidesTable).
		Set("position", position).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"id": ids}).
		ToSql()
	if err != nil {
		return errInternal(fmt.Errorf("error building reorder slides query: %w", err))
	}

	if _, err := r.client.Exec(ctx, query, args...); err != nil {
		return errInternal(fmt.Errorf("error reordering slides: %w", err))
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
