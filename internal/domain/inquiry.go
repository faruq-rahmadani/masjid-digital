package domain

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

const (
	maxInquiryLineLength   = 100  // full name, domicile and source
	maxInquiryReasonLength = 1000 // fits a Telegram message (4096) with room for the rest
)

// InquiryRequest is what the public site's inquiry form sends. Nothing of it is
// stored: the usecase forwards it to the team's Telegram chat.
type InquiryRequest struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Domicile string `json:"domicile"`
	Reason   string `json:"reason"`
	Source   string `json:"source"` // optional: where they heard of Masjid Digital
	Consent  bool   `json:"consent"`
	// Website is a honeypot: the form hides it from people, so only bots fill it in
	Website string `json:"website"`
}

// Inquiry is a validated InquiryRequest, ready to forward
type Inquiry struct {
	FullName string
	Phone    string // E.164, e.g. +6281234567890
	Domicile string
	Reason   string
	Source   string
}

// IsBot tells whether the honeypot field was filled in
func (r InquiryRequest) IsBot() bool {
	return strings.TrimSpace(r.Website) != ""
}

// Validate trims and checks every field, normalizing the phone number to E.164.
// One-line fields have their whitespace collapsed, so a name cannot add lines to the message.
func (r InquiryRequest) Validate() (Inquiry, error) {
	inquiry := Inquiry{
		FullName: oneLine(r.FullName),
		Domicile: oneLine(r.Domicile),
		Reason:   strings.TrimSpace(r.Reason),
		Source:   oneLine(r.Source),
	}

	required := []struct{ value, label string }{
		{inquiry.FullName, "nama lengkap"},
		{r.Phone, "nomor ponsel"},
		{inquiry.Domicile, "domisili"},
		{inquiry.Reason, "alasan"},
	}
	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			return Inquiry{}, apperror.New(apperror.CodeInvalidInput, fmt.Errorf("missing %s", field.label), field.label+" wajib diisi")
		}
	}

	limits := []struct {
		value, label string
		max          int
	}{
		{inquiry.FullName, "nama lengkap", maxInquiryLineLength},
		{inquiry.Domicile, "domisili", maxInquiryLineLength},
		{inquiry.Source, "informasi lainnya", maxInquiryLineLength},
		{inquiry.Reason, "alasan", maxInquiryReasonLength},
	}
	for _, field := range limits {
		if utf8.RuneCountInString(field.value) > field.max {
			return Inquiry{}, apperror.New(apperror.CodeInvalidInput, fmt.Errorf("%s longer than %d characters", field.label, field.max), fmt.Sprintf("%s maksimal %d karakter", field.label, field.max))
		}
	}

	phone, err := NormalizePhone(r.Phone)
	if err != nil {
		return Inquiry{}, apperror.New(apperror.CodeInvalidInput, err, "nomor ponsel tidak valid")
	}
	inquiry.Phone = phone

	if !r.Consent {
		return Inquiry{}, apperror.New(apperror.CodeInvalidInput, errors.New("consent not given"), "centang persetujuan untuk dihubungi")
	}

	return inquiry, nil
}

// WhatsAppURL opens a chat with the sender
func (i Inquiry) WhatsAppURL() string {
	return "https://wa.me/" + strings.TrimPrefix(i.Phone, "+")
}

// oneLine trims s and collapses every run of whitespace, line breaks included, to one space
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
