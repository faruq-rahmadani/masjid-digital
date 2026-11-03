package controller

//go:generate go tool mockgen -source=activity_log.go -destination=../../../../mock/controller_activity_log.go -package=mock

import (
	"context"
	"net/http"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/labstack/echo/v4"
)

type ActivityLogUsecaseItf interface {
	ListRecent(ctx context.Context, req domain.ListActivityLogsRequest) ([]domain.ActivityLog, error)
}

// ListActivityLogs handles GET /activity-logs?limit=N: the latest changes, newest first
func (ctrl *Controller) ListActivityLogs(c echo.Context) error {
	logs, err := ctrl.activityLog.ListRecent(c.Request().Context(), domain.ListActivityLogsRequest{
		Limit: c.QueryParam("limit"),
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: logs})
}
