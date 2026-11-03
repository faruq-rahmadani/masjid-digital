package activitylogs

import (
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
)

type activityLogModel struct {
	ID        int64     `db:"id"`
	ActorID   *int64    `db:"actor_id"`
	ActorName *string   `db:"actor_name"`
	Action    string    `db:"action"`
	EntityID  *int64    `db:"entity_id"`
	Summary   string    `db:"summary"`
	CreatedAt time.Time `db:"created_at"`
}

// activityLogColumns selects a log with its actor's name (LEFT JOIN users u)
var activityLogColumns = []string{
	"al.id", "al.actor_id", "u.name AS actor_name", "al.action", "al.entity_id", "al.summary", "al.created_at",
}

func (m activityLogModel) toDomain() domain.ActivityLog {
	var actorID, entityID int64
	if m.ActorID != nil {
		actorID = *m.ActorID
	}
	if m.EntityID != nil {
		entityID = *m.EntityID
	}
	var actorName string
	if m.ActorName != nil {
		actorName = *m.ActorName
	}
	return domain.ActivityLog{
		ID:        m.ID,
		ActorID:   actorID,
		ActorName: actorName,
		Action:    domain.ActivityAction(m.Action),
		EntityID:  entityID,
		Summary:   m.Summary,
		CreatedAt: m.CreatedAt.UTC(),
	}
}

func nullableID(id int64) *int64 {
	if id == 0 {
		return nil
	}
	return &id
}
