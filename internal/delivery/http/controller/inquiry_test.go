package controller

import (
	"errors"
	"net/http"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"go.uber.org/mock/gomock"
)

func TestSubmitInquiry(t *testing.T) {
	inquiries := mock.NewMockInquiryUsecaseItf(gomock.NewController(t))
	inquiries.EXPECT().Submit(gomock.Any(), domain.InquiryRequest{
		FullName: "Ahmad Fauzi", Phone: "081234567890", Domicile: "Banjarbaru",
		Reason: "Ingin pasang di masjid", Source: "Media sosial", Consent: true,
	}).Return(nil)

	body := `{"full_name":"Ahmad Fauzi","phone":"081234567890","domicile":"Banjarbaru","reason":"Ingin pasang di masjid","source":"Media sosial","consent":true}`
	rec, err := serve(NewController(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, inquiries).SubmitInquiry, http.MethodPost, "/api/public/v1/inquiries", body, "")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rec.Code)
	}
}

func TestSubmitInquiryErrors(t *testing.T) {
	t.Run("malformed JSON", func(t *testing.T) {
		inquiries := mock.NewMockInquiryUsecaseItf(gomock.NewController(t))
		_, err := serve(NewController(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, inquiries).SubmitInquiry, http.MethodPost, "/api/public/v1/inquiries", `{"full_name":`, "")
		if appErr, ok := errors.AsType[*apperror.AppError](err); !ok || appErr.Code != apperror.CodeInvalidInput {
			t.Errorf("error = %v, want INVALID_INPUT", err)
		}
	})

	t.Run("usecase error passes through", func(t *testing.T) {
		inquiries := mock.NewMockInquiryUsecaseItf(gomock.NewController(t))
		want := apperror.New(apperror.CodeUnavailable, errors.New("down"), "pesan belum terkirim")
		inquiries.EXPECT().Submit(gomock.Any(), gomock.Any()).Return(want)
		_, err := serve(NewController(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, inquiries).SubmitInquiry, http.MethodPost, "/api/public/v1/inquiries", `{}`, "")
		if !errors.Is(err, want) {
			t.Errorf("error = %v, want %v", err, want)
		}
	})
}
