package mosqueprofile

import (
	"context"
	"errors"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"go.uber.org/mock/gomock"
)

// repository errors, shaped like the real repository's AppErrors
var (
	errProfileNotFound = apperror.New(apperror.CodeNotFound, domain.ErrMosqueProfileNotFound, "profil masjid belum diisi")
	errLogoNotFound    = apperror.New(apperror.CodeNotFound, domain.ErrMosqueLogoNotFound, "logo masjid belum diunggah")
	errDatabase        = apperror.New(apperror.CodeInternal, errors.New("db down"), "gagal mengakses data profil masjid")
)

// pngHeader is enough of a PNG for content sniffing
var pngHeader = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")

func newTestUsecase(t *testing.T) (*UseCase, *mock.MockMosqueProfileRepositoryItf) {
	t.Helper()

	repo := mock.NewMockMosqueProfileRepositoryItf(gomock.NewController(t))
	return NewMosqueProfileUsecase(repo), repo
}

func validProfileRequest() domain.UpdateMosqueProfileRequest {
	return domain.UpdateMosqueProfileRequest{
		UpdatedBy:    7,
		Name:         "  Masjid Al-Ikhlas ",
		Address:      " Jl. Merdeka No. 1 ",
		ContactPhone: "0812-3456-7890",
		Latitude:     -6.2088,
		Longitude:    106.8456,
		Timezone:     domain.MosqueTimezoneWIB,
	}
}

// wantProfile is validProfileRequest normalized
var wantProfile = domain.MosqueProfile{
	Name:         "Masjid Al-Ikhlas",
	Address:      "Jl. Merdeka No. 1",
	ContactPhone: "+6281234567890",
	Latitude:     -6.2088,
	Longitude:    106.8456,
	Timezone:     domain.MosqueTimezoneWIB,
}

func TestGet(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().Get(gomock.Any()).Return(wantProfile, nil)

	got, err := uc.Get(context.Background())
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != wantProfile {
		t.Errorf("profile = %+v, want %+v", got, wantProfile)
	}
}

func TestGetNeverSaved(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().Get(gomock.Any()).Return(domain.MosqueProfile{}, errProfileNotFound)

	got, err := uc.Get(context.Background())
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != (domain.MosqueProfile{}) {
		t.Errorf("profile = %+v, want empty", got)
	}
}

func TestGetRepositoryError(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().Get(gomock.Any()).Return(domain.MosqueProfile{}, errDatabase)

	if _, err := uc.Get(context.Background()); err != errDatabase {
		t.Errorf("err = %v, want the repository error returned as is", err)
	}
}

func TestUpdate(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().Update(gomock.Any(), wantProfile, int64(7)).Return(nil)

	got, err := uc.Update(context.Background(), validProfileRequest())
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if got != wantProfile {
		t.Errorf("profile = %+v, want %+v", got, wantProfile)
	}
}

func TestUpdateOptionalFields(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().Update(gomock.Any(), gomock.Cond(func(p domain.MosqueProfile) bool {
		return p.Address == "" && p.ContactPhone == ""
	}), gomock.Any()).Return(nil)

	req := validProfileRequest()
	req.Address = "  "
	req.ContactPhone = " "
	if _, err := uc.Update(context.Background(), req); err != nil {
		t.Fatalf("update: %v", err)
	}
}

func TestUpdateInvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*domain.UpdateMosqueProfileRequest)
	}{
		{name: "empty name", modify: func(r *domain.UpdateMosqueProfileRequest) { r.Name = " " }},
		{name: "invalid contact phone", modify: func(r *domain.UpdateMosqueProfileRequest) { r.ContactPhone = "123" }},
		{name: "missing location", modify: func(r *domain.UpdateMosqueProfileRequest) { r.Latitude, r.Longitude = 0, 0 }},
		{name: "latitude out of range", modify: func(r *domain.UpdateMosqueProfileRequest) { r.Latitude = 91 }},
		{name: "longitude out of range", modify: func(r *domain.UpdateMosqueProfileRequest) { r.Longitude = -181 }},
		{name: "empty timezone", modify: func(r *domain.UpdateMosqueProfileRequest) { r.Timezone = "" }},
		{name: "timezone outside Indonesia", modify: func(r *domain.UpdateMosqueProfileRequest) { r.Timezone = "Asia/Singapore" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// no EXPECT: invalid input must not reach the repository
			uc, _ := newTestUsecase(t)

			req := validProfileRequest()
			tt.modify(&req)

			_, err := uc.Update(context.Background(), req)
			assertCode(t, err, apperror.CodeInvalidInput)
		})
	}
}

func TestUpdateRepositoryError(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(errDatabase)

	if _, err := uc.Update(context.Background(), validProfileRequest()); err != errDatabase {
		t.Errorf("err = %v, want the repository error returned as is", err)
	}
}

func TestGetLogoNeverUploaded(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().GetLogo(gomock.Any()).Return(domain.MosqueLogo{}, errLogoNotFound)

	if _, err := uc.GetLogo(context.Background()); err != errLogoNotFound {
		t.Errorf("err = %v, want the repository NOT_FOUND returned as is", err)
	}
}

func TestUpdateLogo(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().UpdateLogo(gomock.Any(), domain.MosqueLogo{ContentType: "image/png", Data: pngHeader}, int64(7)).Return(nil)

	if err := uc.UpdateLogo(context.Background(), domain.UpdateMosqueLogoRequest{UpdatedBy: 7, Data: pngHeader}); err != nil {
		t.Fatalf("update: %v", err)
	}
}

func TestUpdateLogoAcceptedFormats(t *testing.T) {
	images := map[string]struct {
		data        []byte
		contentType string
	}{
		"png":  {pngHeader, "image/png"},
		"jpeg": {[]byte("\xff\xd8\xff\xe0\x00\x10JFIF"), "image/jpeg"},
		"webp": {[]byte("RIFF\x00\x00\x00\x00WEBPVP8 "), "image/webp"},
	}

	for name, img := range images {
		t.Run(name, func(t *testing.T) {
			uc, repo := newTestUsecase(t)
			repo.EXPECT().UpdateLogo(gomock.Any(), gomock.Cond(func(logo domain.MosqueLogo) bool {
				return logo.ContentType == img.contentType
			}), gomock.Any()).Return(nil)

			if err := uc.UpdateLogo(context.Background(), domain.UpdateMosqueLogoRequest{Data: img.data}); err != nil {
				t.Errorf("update: %v", err)
			}
		})
	}
}

func TestUpdateLogoInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{name: "empty", data: nil},
		{name: "too large", data: append(pngHeader, make([]byte, domain.MaxMosqueLogoSize)...)},
		{name: "svg", data: []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`)},
		{name: "gif", data: []byte("GIF89a")},
		{name: "text", data: []byte("bukan gambar")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// no EXPECT: an invalid file must not reach the repository
			uc, _ := newTestUsecase(t)

			err := uc.UpdateLogo(context.Background(), domain.UpdateMosqueLogoRequest{Data: tt.data})
			assertCode(t, err, apperror.CodeInvalidInput)
		})
	}
}

func TestDeleteLogo(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().DeleteLogo(gomock.Any()).Return(nil)

	if err := uc.DeleteLogo(context.Background()); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func assertCode(t *testing.T, err error, want apperror.Code) {
	t.Helper()

	appErr, ok := errors.AsType[*apperror.AppError](err)
	if !ok || appErr.Code != want {
		t.Errorf("err = %v, want %s", err, want)
	}
}
