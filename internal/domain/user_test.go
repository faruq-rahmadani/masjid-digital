package domain

import (
	"errors"
	"testing"
)

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		phone   string
		want    string
		wantErr bool
	}{
		{phone: "081234567890", want: "+6281234567890"},
		{phone: "0812-3456-7890", want: "+6281234567890"},
		{phone: " 0812 3456 7890 ", want: "+6281234567890"},
		{phone: "6281234567890", want: "+6281234567890"},
		{phone: "+62 812.3456.7890", want: "+6281234567890"},
		{phone: "0812345678", want: "+62812345678"},   // shortest mobile number
		{phone: "+60123456789", want: "+60123456789"}, // Malaysia, any E.164 is accepted
		{phone: "", wantErr: true},
		{phone: "812345678", wantErr: true},        // no prefix
		{phone: "021234567", wantErr: true},        // Jakarta landline
		{phone: "08123", wantErr: true},            // too short
		{phone: "0812345678901234", wantErr: true}, // too long
		{phone: "0812abc67890", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.phone, func(t *testing.T) {
			got, err := NormalizePhone(tt.phone)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidPhone) {
					t.Errorf("NormalizePhone(%q) = %q, %v; want ErrInvalidPhone", tt.phone, got, err)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Errorf("NormalizePhone(%q) = %q, %v; want %q", tt.phone, got, err, tt.want)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		email   string
		want    string
		wantErr bool
	}{
		{email: "", want: ""},
		{email: "   ", want: ""},
		{email: " Ahmad@Example.COM ", want: "ahmad@example.com"},
		{email: "ahmad", wantErr: true},
		{email: "Ahmad <ahmad@example.com>", wantErr: true},
		{email: "ahmad@", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			got, err := NormalizeEmail(tt.email)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidEmail) {
					t.Errorf("NormalizeEmail(%q) = %q, %v; want ErrInvalidEmail", tt.email, got, err)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Errorf("NormalizeEmail(%q) = %q, %v; want %q", tt.email, got, err, tt.want)
			}
		})
	}
}
