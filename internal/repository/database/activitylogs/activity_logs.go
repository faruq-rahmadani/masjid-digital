package activitylogs

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/repository/database"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

const activityLogsTable = "activity_logs"

// builder emits `?` placeholders; pkg/database rebinds them for the engine in use
var builder = sq.StatementBuilder.PlaceholderFormat(sq.Question)

type repo struct {
	client database.DatabaseClient
}

func New(client database.DatabaseClient) *repo {
	return &repo{client: client}
}

// Create inserts one log
func (r *repo) Create(ctx context.Context, log domain.ActivityLog) error {
	query, args, err := builder.
		Insert(activityLogsTable).
		Columns("actor_id", "action", "entity_id", "summary", "created_at").
		Values(nullableID(log.ActorID), string(log.Action), nullableID(log.EntityID), log.Summary, log.CreatedAt.UTC()).
		ToSql()
	if err != nil {
		return errInternal(fmt.Errorf("error building create activity log query: %w", err))
	}

	if _, err := r.client.Exec(ctx, query, args...); err != nil {
		return errInternal(fmt.Errorf("error creating activity log: %w", err))
	}
	return nil
}

// FindRecent returns the latest limit logs, newest first, with their actor's name
func (r *repo) FindRecent(ctx context.Context, limit int) ([]domain.ActivityLog, error) {
	query, args, err := builder.
		Select(activityLogColumns...).
		From(activityLogsTable+" al").
		LeftJoin("users u ON u.id = al.actor_id").
		OrderBy("al.created_at DESC", "al.id DESC").
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, errInternal(fmt.Errorf("error building find activity logs query: %w", err))
	}

	var models []activityLogModel
	if err := r.client.Select(ctx, &models, query, args...); err != nil {
		return nil, errInternal(fmt.Errorf("error finding activity logs: %w", err))
	}

	logs := make([]domain.ActivityLog, 0, len(models))
	for _, m := range models {
		logs = append(logs, m.toDomain())
	}
	return logs, nil
}

func errInternal(err error) *apperror.AppError {
	return apperror.New(apperror.CodeInternal, err, "gagal mengakses log aktivitas")
}
