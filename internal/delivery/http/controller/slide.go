package controller

//go:generate go tool mockgen -source=slide.go -destination=../../../../mock/controller_slide.go -package=mock

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

type SlideUsecaseItf interface {
	List(ctx context.Context) ([]domain.Slide, error)
	Create(ctx context.Context, req domain.CreateSlideRequest) (domain.Slide, error)
	Update(ctx context.Context, req domain.UpdateSlideRequest) (domain.Slide, error)
	Delete(ctx context.Context, id int64) error
	Reorder(ctx context.Context, req domain.ReorderSlidesRequest) ([]domain.Slide, error)
	GetImage(ctx context.Context, id int64) (domain.SlideImage, error)
	ListDisplay(ctx context.Context) ([]domain.DisplaySlide, error)
}

const (
	// slideImageFormField is the multipart field holding the uploaded image
	slideImageFormField = "file"
	// slideImageCacheControl lets browsers and proxies keep an image for good: a slide's image
	// never changes and slide IDs are never reused
	slideImageCacheControl = "public, max-age=31536000, immutable"
)

func (ctrl *Controller) ListSlides(c echo.Context) error {
	slides, err := ctrl.slide.List(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: slides})
}

// CreateSlide handles POST /slides: a multipart/form-data upload with the image in the "file"
// field and optional "title", "caption", "duration_seconds" and "show_caption" (true when left
// out). One image per request.
func (ctrl *Controller) CreateSlide(c echo.Context) error {
	session, err := middleware.Session(c)
	if err != nil {
		return err
	}

	// Bind leaves out what the client didn't send, so the caption shows unless told otherwise
	req := domain.CreateSlideRequest{ShowCaption: true}
	if err := c.Bind(&req); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}

	req.CreatedBy = session.UserID
	if req.FileName, req.Data, err = readSlideImage(c); err != nil {
		return err
	}

	slide, err := ctrl.slide.Create(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, dataResponse{Data: slide})
}

func (ctrl *Controller) UpdateSlide(c echo.Context) error {
	id, err := paramID(c)
	if err != nil {
		return err
	}

	var req domain.UpdateSlideRequest
	if err := c.Bind(&req); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}

	req.ID = id

	slide, err := ctrl.slide.Update(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: slide})
}

func (ctrl *Controller) DeleteSlide(c echo.Context) error {
	id, err := paramID(c)
	if err != nil {
		return err
	}

	if err := ctrl.slide.Delete(c.Request().Context(), id); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, messageResponse{Message: "slide dihapus"})
}

func (ctrl *Controller) ReorderSlides(c echo.Context) error {
	var req domain.ReorderSlidesRequest
	if err := c.Bind(&req); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}

	slides, err := ctrl.slide.Reorder(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: slides})
}

// GetSlideImage handles GET /slides/:id/image, serving the image bytes so an <img> can point at it
func (ctrl *Controller) GetSlideImage(c echo.Context) error {
	id, err := paramID(c)
	if err != nil {
		return err
	}

	image, err := ctrl.slide.GetImage(c.Request().Context(), id)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderCacheControl, slideImageCacheControl)
	return c.Blob(http.StatusOK, image.ContentType, image.Data)
}

// GetSlideDisplay handles the TV's GET /slides: the active slides in order
func (ctrl *Controller) GetSlideDisplay(c echo.Context) error {
	slides, err := ctrl.slide.ListDisplay(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dataResponse{Data: slides})
}

// readSlideImage reads the uploaded image and its file name, at most one byte over the size
// limit so the domain can still tell the file is too large
func readSlideImage(c echo.Context) (string, []byte, error) {
	header, err := c.FormFile(slideImageFormField)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) || errors.Is(err, http.ErrNotMultipart) {
			return "", nil, apperror.New(apperror.CodeInvalidInput, err, "file gambar wajib diunggah")
		}
		return "", nil, apperror.New(apperror.CodeInvalidInput, err, "format request tidak valid")
	}

	file, err := header.Open()
	if err != nil {
		return "", nil, apperror.New(apperror.CodeInternal, fmt.Errorf("error opening uploaded slide image: %w", err), "gagal membaca file gambar")
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, domain.MaxSlideImageSize+1))
	if err != nil {
		return "", nil, apperror.New(apperror.CodeInternal, fmt.Errorf("error reading uploaded slide image: %w", err), "gagal membaca file gambar")
	}

	return header.Filename, data, nil
}
