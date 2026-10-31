package controller

//go:generate go tool mockgen -source=mosque_profile.go -destination=../../../../mock/controller_mosque_profile.go -package=mock

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/faruq-rahmadani/masjid-digital/internal/delivery/http/middleware"
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
)

type MosqueProfileUsecaseItf interface {
	Get(ctx context.Context) (domain.MosqueProfile, error)
	Update(ctx context.Context, req domain.UpdateMosqueProfileRequest) (domain.MosqueProfile, error)
	GetLogo(ctx context.Context) (domain.MosqueLogo, error)
	UpdateLogo(ctx context.Context, req domain.UpdateMosqueLogoRequest) error
	DeleteLogo(ctx context.Context) error
}

// mosqueLogoFormField is the multipart field holding the uploaded logo
const mosqueLogoFormField = "file"

// GetMosqueProfile handles GET /mosque-profile. Every field is
// empty until the profile is first saved.
func (ctrl *Controller) GetMosqueProfile(c echo.Context) error {
	profile, err := ctrl.mosqueProfile.Get(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: profile})
}

// UpdateMosqueProfile handles PUT /mosque-profile, replacing every field
func (ctrl *Controller) UpdateMosqueProfile(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}

	var req domain.UpdateMosqueProfileRequest
	if err := c.Bind(&req); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}

	// set after Bind so whatever the client sent is overwritten
	req.UpdatedBy = session.UserID

	profile, err := ctrl.mosqueProfile.Update(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: profile})
}

// GetMosqueLogo handles GET /mosque-profile/logo
func (ctrl *Controller) GetMosqueLogo(c echo.Context) error {
	logo, err := ctrl.mosqueProfile.GetLogo(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: struct {
		URL string `json:"url"`
	}{URL: logo.DataURI()}})
}

// UpdateMosqueLogo handles PUT /mosque-profile/logo: a multipart/form-data
// upload with the image in the "file" field
func (ctrl *Controller) UpdateMosqueLogo(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}

	data, err := readMosqueLogo(c)
	if err != nil {
		return err
	}

	if err := ctrl.mosqueProfile.UpdateLogo(c.Request().Context(), domain.UpdateMosqueLogoRequest{
		UpdatedBy: session.UserID,
		Data:      data,
	}); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, messageResponse{Message: "logo masjid berhasil disimpan"})
}

// DeleteMosqueLogo handles DELETE /mosque-profile/logo
func (ctrl *Controller) DeleteMosqueLogo(c echo.Context) error {
	if err := ctrl.mosqueProfile.DeleteLogo(c.Request().Context()); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, messageResponse{Message: "logo masjid berhasil dihapus"})
}

// readMosqueLogo reads the uploaded logo, at most one byte over the size limit
// so the domain can still tell the file is too large
func readMosqueLogo(c echo.Context) ([]byte, error) {
	header, err := c.FormFile(mosqueLogoFormField)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) || errors.Is(err, http.ErrNotMultipart) {
			return nil, apperror.New(apperror.CodeInvalidInput, err, "file logo wajib diunggah")
		}
		return nil, apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}

	file, err := header.Open()
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, fmt.Errorf("error opening uploaded logo: %w", err), "gagal membaca file logo")
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, domain.MaxMosqueLogoSize+1))
	if err != nil {
		return nil, apperror.New(apperror.CodeInternal, fmt.Errorf("error reading uploaded logo: %w", err), "gagal membaca file logo")
	}

	return data, nil
}
