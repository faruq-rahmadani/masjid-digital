package controller

//go:generate go tool mockgen -source=inquiry.go -destination=../../../../mock/controller_inquiry.go -package=mock

import (
	"context"
	"net/http"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
)

type InquiryUsecaseItf interface {
	Submit(ctx context.Context, req domain.InquiryRequest) error
}

// SubmitInquiry handles POST /inquiries, the public site's inquiry form. The router
// rate limits it, since anyone can call it.
func (ctrl *Controller) SubmitInquiry(c echo.Context) error {
	var req domain.InquiryRequest
	if err := c.Bind(&req); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}

	if err := ctrl.inquiry.Submit(c.Request().Context(), req); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, messageResponse{Message: "terima kasih, pesan Anda sudah kami terima"})
}
