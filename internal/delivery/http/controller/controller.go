package controller

import (
	"fmt"
	"strconv"

	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
)

type Controller struct {
	user          UserUsecaseItf
	auth          AuthUsecaseItf
	mosqueProfile MosqueProfileUsecaseItf
	prayerTime    PrayerTimeUsecaseItf
	runningText   RunningTextUsecaseItf
}

func NewController(user UserUsecaseItf, auth AuthUsecaseItf, mosqueProfile MosqueProfileUsecaseItf, prayerTime PrayerTimeUsecaseItf, runningText RunningTextUsecaseItf) *Controller {
	return &Controller{user: user, auth: auth, mosqueProfile: mosqueProfile, prayerTime: prayerTime, runningText: runningText}
}

// dataResponse wraps a successful payload as {"data": ...}
type dataResponse struct {
	Data any `json:"data"`
}

type messageResponse struct {
	Message string `json:"message"`
}

// paramID reads the :id path parameter as a positive ID
func paramID(c echo.Context) (int64, error) {
	raw := c.Param("id")

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, apperror.New(apperror.CodeInvalidInput, fmt.Errorf("invalid id %q", raw), "id tidak valid")
	}

	return id, nil
}
