package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// serverTimeResponse is the server's clock as Unix milliseconds
type serverTimeResponse struct {
	Now int64 `json:"now"`
}

// GetServerTime handles GET /time. The TV display corrects its own clock with
// it, since a TV's clock is often unset or drifting and prayer times need the
// minute right.
func (ctrl *Controller) GetServerTime(c echo.Context) error {
	// a cached answer is a stale clock
	c.Response().Header().Set(echo.HeaderCacheControl, "no-store")

	return c.JSON(http.StatusOK, dataResponse{Data: serverTimeResponse{Now: time.Now().UnixMilli()}})
}
