package activitylog

import (
	"context"
	"log"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase"
)

type UseCase struct {
	logs usecase.ActivityLogRepositoryItf
}

func NewActivityLogUsecase(logs usecase.ActivityLogRepositoryItf) *UseCase {
	return &UseCase{logs: logs}
}

// Save stores a published log. It is the activity log bus's handler: best effort, so a failed
// insert is only logged.
func (u *UseCase) Save(ctx context.Context, activityLog domain.ActivityLog) {
	if err := u.logs.Create(ctx, activityLog); err != nil {
		log.Printf("error saving activity log %s: %v", activityLog.Action, err)
	}
}

// ListRecent returns the latest logs, newest first
func (u *UseCase) ListRecent(ctx context.Context, req domain.ListActivityLogsRequest) ([]domain.ActivityLog, error) {
	limit, err := req.ValidateAndParse()
	if err != nil {
		return nil, err
	}
	return u.logs.FindRecent(ctx, limit)
}
