package inquiry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"go.uber.org/mock/gomock"
)

func validRequest() domain.InquiryRequest {
	return domain.InquiryRequest{
		FullName: "Ahmad Fauzi",
		Phone:    "081234567890",
		Domicile: "Banjarbaru",
		Reason:   "Ingin memasang jadwal sholat di TV masjid.",
		Source:   "Media sosial",
		Consent:  true,
	}
}

func newUsecase(t *testing.T, envTag string) (*UseCase, *mock.MockInquiryNotifierItf) {
	t.Helper()
	notifier := mock.NewMockInquiryNotifierItf(gomock.NewController(t))
	uc := New(notifier, envTag)
	uc.now = func() time.Time { return time.Date(2026, 11, 19, 7, 3, 0, 0, time.UTC) }
	return uc, notifier
}

func codeOf(err error) apperror.Code {
	if appErr, ok := errors.AsType[*apperror.AppError](err); ok {
		return appErr.Code
	}
	return ""
}

func TestSubmit(t *testing.T) {
	tests := []struct {
		name, envTag, want string
	}{
		{"production has no tag", "", "📩 Inquiry baru dari halaman publik\n\n" +
			"Nama: Ahmad Fauzi\n" +
			"Ponsel: +6281234567890\n" +
			"WhatsApp: https://wa.me/6281234567890\n" +
			"Domisili: Banjarbaru\n" +
			"Tahu dari: Media sosial\n\n" +
			"Alasan:\nIngin memasang jadwal sholat di TV masjid.\n\n" +
			"Dikirim 19 Nov 2026 14:03 WIB"},
		{"staging is tagged", " staging ", "[STAGING] 📩 Inquiry baru dari halaman publik\n\n" +
			"Nama: Ahmad Fauzi\n" +
			"Ponsel: +6281234567890\n" +
			"WhatsApp: https://wa.me/6281234567890\n" +
			"Domisili: Banjarbaru\n" +
			"Tahu dari: Media sosial\n\n" +
			"Alasan:\nIngin memasang jadwal sholat di TV masjid.\n\n" +
			"Dikirim 19 Nov 2026 14:03 WIB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, notifier := newUsecase(t, tt.envTag)
			notifier.EXPECT().Send(gomock.Any(), tt.want).Return(nil)

			if err := uc.Submit(context.Background(), validRequest()); err != nil {
				t.Fatalf("Submit: %v", err)
			}
		})
	}
}

func TestSubmitWithoutSource(t *testing.T) {
	uc, notifier := newUsecase(t, "")
	notifier.EXPECT().Send(gomock.Any(), gomock.Not(gomock.Regex("Tahu dari"))).Return(nil)

	req := validRequest()
	req.Source = ""
	if err := uc.Submit(context.Background(), req); err != nil {
		t.Fatalf("Submit: %v", err)
	}
}

func TestSubmitRejectsInvalidInput(t *testing.T) {
	uc, _ := newUsecase(t, "") // no Send expected

	req := validRequest()
	req.Consent = false
	if err := uc.Submit(context.Background(), req); codeOf(err) != apperror.CodeInvalidInput {
		t.Errorf("Submit error = %v, want INVALID_INPUT", err)
	}
}

func TestSubmitDropsBots(t *testing.T) {
	uc, _ := newUsecase(t, "") // no Send expected

	req := validRequest()
	req.Website = "http://spam.example"
	if err := uc.Submit(context.Background(), req); err != nil {
		t.Errorf("Submit error = %v, want nil so the bot cannot tell", err)
	}
}

func TestSubmitSendFails(t *testing.T) {
	uc, notifier := newUsecase(t, "")
	notifier.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("telegram: status 502"))

	if err := uc.Submit(context.Background(), validRequest()); codeOf(err) != apperror.CodeUnavailable {
		t.Errorf("Submit error = %v, want UNAVAILABLE", err)
	}
}

func TestSubmitNotConfigured(t *testing.T) {
	uc := New(nil, "")
	if err := uc.Submit(context.Background(), validRequest()); codeOf(err) != apperror.CodeUnavailable {
		t.Errorf("Submit error = %v, want UNAVAILABLE", err)
	}
}
