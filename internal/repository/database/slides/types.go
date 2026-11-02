package slides

import (
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
)

type slideModel struct {
	ID              int64     `db:"id"`
	Title           string    `db:"title"`
	Caption         string    `db:"caption"`
	DurationSeconds int       `db:"duration_seconds"`
	ShowCaption     bool      `db:"show_caption"`
	Active          bool      `db:"active"`
	Position        int       `db:"position"`
	CreatedBy       *int64    `db:"created_by"`
	CreatedByName   *string   `db:"created_by_name"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

type imageModel struct {
	ContentType string `db:"image_type"`
	Data        []byte `db:"image_data"`
}

// slideColumns selects a slide, without its image, with its creator's name (LEFT JOIN users u)
var slideColumns = []string{
	"s.id", "s.title", "s.caption", "s.duration_seconds", "s.show_caption", "s.active", "s.position", "s.created_by",
	"u.name AS created_by_name", "s.created_at", "s.updated_at",
}

func (m slideModel) toDomain() domain.Slide {
	var createdBy int64
	if m.CreatedBy != nil {
		createdBy = *m.CreatedBy
	}
	var createdByName string
	if m.CreatedByName != nil {
		createdByName = *m.CreatedByName
	}
	return domain.Slide{
		ID:              m.ID,
		Title:           m.Title,
		Caption:         m.Caption,
		DurationSeconds: m.DurationSeconds,
		ShowCaption:     m.ShowCaption,
		Active:          m.Active,
		Position:        m.Position,
		CreatedBy:       createdBy,
		CreatedByName:   createdByName,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}
