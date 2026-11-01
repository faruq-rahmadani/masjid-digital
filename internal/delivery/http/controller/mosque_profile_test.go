package controller

import (
	"bytes"
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

var mosqueProfile = domain.MosqueProfile{
	Name:         "Masjid Al-Ikhlas",
	Address:      "Jl. Merdeka No. 1",
	ContactPhone: "+6281234567890",
	Latitude:     -6.2088,
	Longitude:    106.8456,
	Timezone:     domain.MosqueTimezoneWIB,
}

const mosqueProfileJSON = `{"name":"Masjid Al-Ikhlas","address":"Jl. Merdeka No. 1","contact_phone":"+6281234567890","latitude":-6.2088,"longitude":106.8456,"timezone":"Asia/Jakarta"}`

// serveUpload runs handler for a logged in multipart/form-data request carrying
// data as the file field, or no file at all when field is empty
func serveUpload(handler echo.HandlerFunc, field string, data []byte) (*httptest.ResponseRecorder, error) {
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	if field != "" {
		part, _ := form.CreateFormFile(field, "logo.png")
		part.Write(data)
	}
	form.Close()

	req := httptest.NewRequest(http.MethodPut, "/api/v1/mosque-profile/logo", &body)
	req.Header.Set(echo.HeaderContentType, form.FormDataContentType())
	req.Header.Set(echo.HeaderAuthorization, "Bearer valid-token")

	rec := httptest.NewRecorder()
	return rec, handler(echo.New().NewContext(req, rec))
}

func TestGetMosqueProfile(t *testing.T) {
	profiles := mock.NewMockMosqueProfileUsecaseItf(gomock.NewController(t))
	profiles.EXPECT().Get(gomock.Any()).Return(mosqueProfile, nil)

	rec, err := serve(authenticated(t, NewController(nil, nil, profiles, nil, nil).GetMosqueProfile), http.MethodGet, "/api/v1/mosque-profile", "", "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"data":`+mosqueProfileJSON+`}` {
		t.Errorf("body = %s", got)
	}
}

func TestGetMosqueProfileNeverSaved(t *testing.T) {
	profiles := mock.NewMockMosqueProfileUsecaseItf(gomock.NewController(t))
	profiles.EXPECT().Get(gomock.Any()).Return(domain.MosqueProfile{}, nil)

	rec, err := serve(authenticated(t, NewController(nil, nil, profiles, nil, nil).GetMosqueProfile), http.MethodGet, "/api/v1/mosque-profile", "", "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	want := `{"data":{"name":"","address":"","contact_phone":"","latitude":0,"longitude":0,"timezone":""}}`
	if got := strings.TrimSpace(rec.Body.String()); got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}

func TestUpdateMosqueProfile(t *testing.T) {
	profiles := mock.NewMockMosqueProfileUsecaseItf(gomock.NewController(t))
	profiles.EXPECT().Update(gomock.Any(), domain.UpdateMosqueProfileRequest{
		UpdatedBy:    loggedInAs.UserID,
		Name:         "Masjid Al-Ikhlas",
		Address:      "Jl. Merdeka No. 1",
		ContactPhone: "0812-3456-7890",
		Latitude:     -6.2088,
		Longitude:    106.8456,
		Timezone:     domain.MosqueTimezoneWIB,
	}).Return(mosqueProfile, nil)

	// UpdatedBy in the body must not fake who made the change
	rec, err := serve(authenticated(t, NewController(nil, nil, profiles, nil, nil).UpdateMosqueProfile), http.MethodPut, "/api/v1/mosque-profile",
		`{"UpdatedBy":1,"name":"Masjid Al-Ikhlas","address":"Jl. Merdeka No. 1","contact_phone":"0812-3456-7890","latitude":-6.2088,"longitude":106.8456,"timezone":"Asia/Jakarta"}`, "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"data":`+mosqueProfileJSON+`}` {
		t.Errorf("body = %s", got)
	}
}

func TestUpdateMosqueProfileInvalidJSON(t *testing.T) {
	// no EXPECT: a malformed body must not reach the usecase
	profiles := mock.NewMockMosqueProfileUsecaseItf(gomock.NewController(t))

	_, err := serve(authenticated(t, NewController(nil, nil, profiles, nil, nil).UpdateMosqueProfile), http.MethodPut, "/api/v1/mosque-profile", `{"name":`, "Bearer valid-token")
	assertCode(t, err, apperror.CodeInvalidInput)
}

func TestUpdateMosqueProfileUsecaseError(t *testing.T) {
	want := apperror.New(apperror.CodeInvalidInput, nil, "lokasi masjid wajib di isi")

	profiles := mock.NewMockMosqueProfileUsecaseItf(gomock.NewController(t))
	profiles.EXPECT().Update(gomock.Any(), gomock.Any()).Return(domain.MosqueProfile{}, want)

	_, err := serve(authenticated(t, NewController(nil, nil, profiles, nil, nil).UpdateMosqueProfile), http.MethodPut, "/api/v1/mosque-profile", `{"name":"Masjid"}`, "Bearer valid-token")
	if err != want {
		t.Errorf("err = %v, want the usecase error returned as is", err)
	}
}

func TestGetMosqueLogo(t *testing.T) {
	profiles := mock.NewMockMosqueProfileUsecaseItf(gomock.NewController(t))
	profiles.EXPECT().GetLogo(gomock.Any()).Return(domain.MosqueLogo{ContentType: "image/png", Data: []byte("\x89PNG\r\n\x1a\n")}, nil)

	rec, err := serve(authenticated(t, NewController(nil, nil, profiles, nil, nil).GetMosqueLogo), http.MethodGet, "/api/v1/mosque-profile/logo", "", "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"data":{"url":"data:image/png;base64,iVBORw0KGgo="}}` {
		t.Errorf("body = %s", got)
	}
}

func TestGetMosqueLogoNotFound(t *testing.T) {
	want := apperror.New(apperror.CodeNotFound, domain.ErrMosqueLogoNotFound, "logo masjid belum diunggah")

	profiles := mock.NewMockMosqueProfileUsecaseItf(gomock.NewController(t))
	profiles.EXPECT().GetLogo(gomock.Any()).Return(domain.MosqueLogo{}, want)

	_, err := serve(authenticated(t, NewController(nil, nil, profiles, nil, nil).GetMosqueLogo), http.MethodGet, "/api/v1/mosque-profile/logo", "", "Bearer valid-token")
	if err != want {
		t.Errorf("err = %v, want the usecase error returned as is", err)
	}
}

func TestUpdateMosqueLogo(t *testing.T) {
	data := []byte("\x89PNG\r\n\x1a\nrest of the image")

	profiles := mock.NewMockMosqueProfileUsecaseItf(gomock.NewController(t))
	profiles.EXPECT().UpdateLogo(gomock.Any(), domain.UpdateMosqueLogoRequest{UpdatedBy: loggedInAs.UserID, Data: data}).Return(nil)

	rec, err := serveUpload(authenticated(t, NewController(nil, nil, profiles, nil, nil).UpdateMosqueLogo), "file", data)
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != `{"message":"logo masjid berhasil disimpan"}` {
		t.Errorf("response = %d %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateMosqueLogoReadsOneByteOverTheLimit(t *testing.T) {
	profiles := mock.NewMockMosqueProfileUsecaseItf(gomock.NewController(t))
	// the usecase gets enough bytes to tell the file is too large, not the whole upload
	profiles.EXPECT().UpdateLogo(gomock.Any(), gomock.Cond(func(req domain.UpdateMosqueLogoRequest) bool {
		return len(req.Data) == domain.MaxMosqueLogoSize+1
	})).Return(nil)

	if _, err := serveUpload(authenticated(t, NewController(nil, nil, profiles, nil, nil).UpdateMosqueLogo), "file", make([]byte, domain.MaxMosqueLogoSize+100)); err != nil {
		t.Fatalf("handler: %v", err)
	}
}

func TestUpdateMosqueLogoRejects(t *testing.T) {
	t.Run("no file", func(t *testing.T) {
		// no EXPECT: the request must not reach the usecase
		profiles := mock.NewMockMosqueProfileUsecaseItf(gomock.NewController(t))

		_, err := serveUpload(authenticated(t, NewController(nil, nil, profiles, nil, nil).UpdateMosqueLogo), "", nil)
		assertCode(t, err, apperror.CodeInvalidInput)
	})

	t.Run("wrong field", func(t *testing.T) {
		profiles := mock.NewMockMosqueProfileUsecaseItf(gomock.NewController(t))

		_, err := serveUpload(authenticated(t, NewController(nil, nil, profiles, nil, nil).UpdateMosqueLogo), "logo", []byte("\x89PNG"))
		assertCode(t, err, apperror.CodeInvalidInput)
	})

	t.Run("not multipart", func(t *testing.T) {
		profiles := mock.NewMockMosqueProfileUsecaseItf(gomock.NewController(t))

		_, err := serve(authenticated(t, NewController(nil, nil, profiles, nil, nil).UpdateMosqueLogo), http.MethodPut, "/api/v1/mosque-profile/logo", `{"file":"x"}`, "Bearer valid-token")
		assertCode(t, err, apperror.CodeInvalidInput)
	})
}

func TestDeleteMosqueLogo(t *testing.T) {
	profiles := mock.NewMockMosqueProfileUsecaseItf(gomock.NewController(t))
	profiles.EXPECT().DeleteLogo(gomock.Any()).Return(nil)

	rec, err := serve(authenticated(t, NewController(nil, nil, profiles, nil, nil).DeleteMosqueLogo), http.MethodDelete, "/api/v1/mosque-profile/logo", "", "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != `{"message":"logo masjid berhasil dihapus"}` {
		t.Errorf("response = %d %s", rec.Code, rec.Body.String())
	}
}
