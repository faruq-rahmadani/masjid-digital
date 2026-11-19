package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

func validInquiryRequest() InquiryRequest {
	return InquiryRequest{
		FullName: "  Ahmad   Fauzi ",
		Phone:    "0812-3456-7890",
		Domicile: "Banjarbaru\nKalimantan Selatan",
		Reason:   "  Ingin memasang jadwal sholat di TV masjid.\nKami punya 2 TV.  ",
		Source:   "Media sosial",
		Consent:  true,
	}
}

func TestInquiryRequestValidate(t *testing.T) {
	got, err := validInquiryRequest().Validate()
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}

	want := Inquiry{
		FullName: "Ahmad Fauzi",
		Phone:    "+6281234567890",
		Domicile: "Banjarbaru Kalimantan Selatan",
		Reason:   "Ingin memasang jadwal sholat di TV masjid.\nKami punya 2 TV.",
		Source:   "Media sosial",
	}
	if got != want {
		t.Errorf("Validate = %+v, want %+v", got, want)
	}
	if url := got.WhatsAppURL(); url != "https://wa.me/6281234567890" {
		t.Errorf("WhatsAppURL = %q", url)
	}
}

func TestInquiryRequestValidateRejects(t *testing.T) {
	tests := []struct {
		name    string
		change  func(*InquiryRequest)
		message string
	}{
		{"no name", func(r *InquiryRequest) { r.FullName = "  " }, "nama lengkap wajib diisi"},
		{"no phone", func(r *InquiryRequest) { r.Phone = "" }, "nomor ponsel wajib diisi"},
		{"bad phone", func(r *InquiryRequest) { r.Phone = "12345" }, "nomor ponsel tidak valid"},
		{"no domicile", func(r *InquiryRequest) { r.Domicile = "\n" }, "domisili wajib diisi"},
		{"no reason", func(r *InquiryRequest) { r.Reason = "" }, "alasan wajib diisi"},
		{"long name", func(r *InquiryRequest) { r.FullName = strings.Repeat("a", 101) }, "nama lengkap maksimal 100 karakter"},
		{"long source", func(r *InquiryRequest) { r.Source = strings.Repeat("a", 101) }, "informasi lainnya maksimal 100 karakter"},
		{"long reason", func(r *InquiryRequest) { r.Reason = strings.Repeat("a", 1001) }, "alasan maksimal 1000 karakter"},
		{"no consent", func(r *InquiryRequest) { r.Consent = false }, "centang persetujuan untuk dihubungi"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := validInquiryRequest()
			tt.change(&r)
			_, err := r.Validate()
			appErr, ok := errors.AsType[*apperror.AppError](err)
			if !ok || appErr.Code != apperror.CodeInvalidInput || appErr.MessageError != tt.message {
				t.Errorf("Validate error = %v, want INVALID_INPUT %q", err, tt.message)
			}
		})
	}
}

func TestInquiryRequestIsBot(t *testing.T) {
	r := validInquiryRequest()
	if r.IsBot() {
		t.Error("IsBot = true without the honeypot")
	}
	r.Website = "http://spam.example"
	if !r.IsBot() {
		t.Error("IsBot = false with the honeypot filled in")
	}
}
