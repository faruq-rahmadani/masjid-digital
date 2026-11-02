package controller

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
	"go.uber.org/mock/gomock"
)

func newSlideController(t *testing.T) (*Controller, *mock.MockSlideUsecaseItf) {
	t.Helper()
	slides := mock.NewMockSlideUsecaseItf(gomock.NewController(t))
	return NewController(nil, nil, nil, nil, nil, nil, nil, slides, nil), slides
}

// serveSlideUpload posts a multipart form with the image as "file" (left out when data is nil)
// and the given text fields
func serveSlideUpload(handler echo.HandlerFunc, data []byte, fields map[string]string) (*httptest.ResponseRecorder, error) {
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	if data != nil {
		part, _ := form.CreateFormFile("file", "Poster Kajian.png")
		part.Write(data)
	}
	for k, v := range fields {
		form.WriteField(k, v)
	}
	form.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/slides", &body)
	req.Header.Set(echo.HeaderContentType, form.FormDataContentType())
	req.Header.Set(echo.HeaderAuthorization, "Bearer valid-token")

	rec := httptest.NewRecorder()
	return rec, handler(echo.New().NewContext(req, rec))
}

func TestCreateSlideTakesCreatorFromSession(t *testing.T) {
	data := []byte("\x89PNG")
	ctrl, slides := newSlideController(t)
	slides.EXPECT().Create(gomock.Any(), domain.CreateSlideRequest{
		CreatedBy:       loggedInAs.UserID,
		Title:           "Kajian Ahad",
		Caption:         "08.00 WIB",
		DurationSeconds: 15,
		ShowCaption:     true, // left out: shown
		FileName:        "Poster Kajian.png",
		Data:            data,
	}).Return(domain.Slide{ID: 5, Title: "Kajian Ahad", DurationSeconds: 15, Active: true}, nil)

	rec, err := serveSlideUpload(authenticated(t, ctrl.CreateSlide), data, map[string]string{"title": "Kajian Ahad", "caption": "08.00 WIB", "duration_seconds": "15", "created_by": "99"})
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rec.Code)
	}
	if got := rec.Body.String(); !strings.Contains(got, `"id":5`) || !strings.Contains(got, `"duration_seconds":15`) {
		t.Errorf("body = %s", got)
	}
}

func TestCreateSlideWithoutCaptionOnTV(t *testing.T) {
	data := []byte("\x89PNG")
	ctrl, slides := newSlideController(t)
	slides.EXPECT().Create(gomock.Any(), gomock.Cond(func(req domain.CreateSlideRequest) bool { return !req.ShowCaption })).
		Return(domain.Slide{ID: 6}, nil)

	if _, err := serveSlideUpload(authenticated(t, ctrl.CreateSlide), data, map[string]string{"show_caption": "false"}); err != nil {
		t.Fatalf("handler: %v", err)
	}
}

func TestCreateSlideBadDuration(t *testing.T) {
	ctrl, _ := newSlideController(t)

	_, err := serveSlideUpload(authenticated(t, ctrl.CreateSlide), []byte("\x89PNG"), map[string]string{"duration_seconds": "lama"})
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeInvalidInput {
		t.Errorf("err = %v, want INVALID_INPUT", err)
	}
}

func TestCreateSlideWithoutFile(t *testing.T) {
	ctrl, _ := newSlideController(t)

	_, err := serveSlideUpload(authenticated(t, ctrl.CreateSlide), nil, map[string]string{"title": "Kajian"})
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeInvalidInput || appErr.MessageError != "file gambar wajib diunggah" {
		t.Errorf("err = %v, want INVALID_INPUT file gambar wajib diunggah", err)
	}
}

func TestUpdateSlideTakesIDFromPath(t *testing.T) {
	ctrl, slides := newSlideController(t)
	slides.EXPECT().Update(gomock.Any(), domain.UpdateSlideRequest{ID: 2, Title: "Infaq", Caption: "BSI", DurationSeconds: 15, Active: true}).
		Return(domain.Slide{ID: 2, Title: "Infaq"}, nil)

	rec, err := serve(authenticated(t, ctrl.UpdateSlide), http.MethodPut, "/api/v1/slides/2", `{"id":9,"title":"Infaq","caption":"BSI","duration_seconds":15,"active":true}`, "Bearer valid-token", "id", "2")
	if err != nil || rec.Code != http.StatusOK {
		t.Fatalf("handler: %d, %v", rec.Code, err)
	}
}

func TestGetSlideImageServesBytesWithLongCache(t *testing.T) {
	data := []byte("\x89PNG\r\n")
	ctrl, slides := newSlideController(t)
	slides.EXPECT().GetImage(gomock.Any(), int64(3)).Return(domain.SlideImage{ContentType: "image/png", Data: data}, nil)

	rec, err := serve(ctrl.GetSlideImage, http.MethodGet, "/api/public/v1/slides/3/image", "", "", "id", "3")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := rec.Header().Get(echo.HeaderContentType); got != "image/png" {
		t.Errorf("content type = %q, want image/png", got)
	}
	if got := rec.Header().Get(echo.HeaderCacheControl); !strings.Contains(got, "immutable") {
		t.Errorf("cache control = %q, want immutable", got)
	}
	if !bytes.Equal(rec.Body.Bytes(), data) {
		t.Errorf("body = %v, want %v", rec.Body.Bytes(), data)
	}
}

func TestGetSlideImageNotFound(t *testing.T) {
	ctrl, slides := newSlideController(t)
	notFound := apperror.New(apperror.CodeNotFound, domain.ErrSlideNotFound, "slide tidak ditemukan")
	slides.EXPECT().GetImage(gomock.Any(), int64(8)).Return(domain.SlideImage{}, notFound)

	_, err := serve(ctrl.GetSlideImage, http.MethodGet, "/api/public/v1/slides/8/image", "", "", "id", "8")
	if !errors.Is(err, domain.ErrSlideNotFound) {
		t.Errorf("err = %v, want ErrSlideNotFound", err)
	}
}

func TestGetSlideDisplay(t *testing.T) {
	ctrl, slides := newSlideController(t)
	slides.EXPECT().ListDisplay(gomock.Any()).Return([]domain.DisplaySlide{{ID: 1, Title: "Kajian", Caption: "Ahad", DurationSeconds: 10, ShowCaption: true}}, nil)

	rec, err := serve(ctrl.GetSlideDisplay, http.MethodGet, "/api/public/v1/slides", "", "")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"data":[{"id":1,"title":"Kajian","caption":"Ahad","duration_seconds":10,"show_caption":true}]}` {
		t.Errorf("body = %s", got)
	}
}
