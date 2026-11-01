package runningtexts

import (
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
)

type settingValue struct {
	Speed string `json:"speed"`
}

type runningTextModel struct {
	ID            int64      `db:"id"`
	Text          string     `db:"text"`
	Active        bool       `db:"active"`
	Position      int        `db:"position"`
	CreatedBy     *int64     `db:"created_by"`
	CreatedByName *string    `db:"created_by_name"`
	DeactivatedAt *time.Time `db:"deactivated_at"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
}

// runningTextColumns selects a running text with its creator's name (LEFT JOIN users u)
var runningTextColumns = []string{
	"rt.id", "rt.text", "rt.active", "rt.position", "rt.created_by", "u.name AS created_by_name",
	"rt.deactivated_at", "rt.created_at", "rt.updated_at",
}

func (m runningTextModel) toDomain() domain.RunningText {
	var createdBy int64
	if m.CreatedBy != nil {
		createdBy = *m.CreatedBy
	}
	var createdByName string
	if m.CreatedByName != nil {
		createdByName = *m.CreatedByName
	}
	return domain.RunningText{
		ID:            m.ID,
		Text:          m.Text,
		Active:        m.Active,
		Position:      m.Position,
		CreatedBy:     createdBy,
		CreatedByName: createdByName,
		DeactivatedAt: m.DeactivatedAt,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}
