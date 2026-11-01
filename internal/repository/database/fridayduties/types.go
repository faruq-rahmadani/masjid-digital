package fridayduties

import (
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
)

// dutyModel is a row of friday_duties
type dutyModel struct {
	ID        int64     `db:"id"`
	Date      string    `db:"date"`
	Khatib    string    `db:"khatib"`
	Imam      string    `db:"imam"`
	Muadzin   string    `db:"muadzin"`
	Theme     string    `db:"theme"`
	UpdatedBy *int64    `db:"updated_by"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

var dutyColumns = []string{"id", "date", "khatib", "imam", "muadzin", "theme", "updated_by", "created_at", "updated_at"}

func (m dutyModel) toDomain() domain.FridayDuty {
	var updatedBy int64
	if m.UpdatedBy != nil {
		updatedBy = *m.UpdatedBy
	}
	return domain.FridayDuty{
		Date:      m.Date,
		Khatib:    m.Khatib,
		Imam:      m.Imam,
		Muadzin:   m.Muadzin,
		Theme:     m.Theme,
		UpdatedBy: updatedBy,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
