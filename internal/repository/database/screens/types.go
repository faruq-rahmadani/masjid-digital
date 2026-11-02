package screens

import (
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
)

type screenModel struct {
	ID          int64      `db:"id"`
	DeviceID    string     `db:"device_id"`
	Name        string     `db:"name"`
	NameHint    string     `db:"name_hint"`
	UserAgent   string     `db:"user_agent"`
	ForcedMode  string     `db:"forced_mode"`
	ForcedUntil *time.Time `db:"forced_until"`
	BlockedAt   *time.Time `db:"blocked_at"`
	FirstSeenAt time.Time  `db:"first_seen_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

var screenColumns = []string{
	"id", "device_id", "name", "name_hint", "user_agent", "forced_mode", "forced_until", "blocked_at", "first_seen_at", "updated_at",
}

func (m screenModel) toDomain() domain.Screen {
	return domain.Screen{
		ID:          m.ID,
		DeviceID:    m.DeviceID,
		Name:        m.Name,
		NameHint:    m.NameHint,
		UserAgent:   m.UserAgent,
		ForcedMode:  domain.ScreenMode(m.ForcedMode),
		ForcedUntil: m.ForcedUntil,
		BlockedAt:   m.BlockedAt,
		FirstSeenAt: m.FirstSeenAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
