package domain

import (
	"strings"
	"testing"
)

func TestCreateRunningTextNormalizes(t *testing.T) {
	req := CreateRunningTextRequest{Text: "  Kajian Tafsir setiap Kamis  "}
	if err := req.ValidateAndNormalize(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if req.Text != "Kajian Tafsir setiap Kamis" {
		t.Errorf("text = %q, want trimmed", req.Text)
	}
}

func TestRunningTextLength(t *testing.T) {
	tests := []struct {
		name, text, wantMessage string
	}{
		{"empty", "", "teks wajib diisi"},
		{"blank", "   ", "teks wajib diisi"},
		{"200 characters", strings.Repeat("é", 200), ""}, // counted in characters, not bytes
		{"201 characters", strings.Repeat("a", 201), "teks maksimal 200 karakter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := UpdateRunningTextRequest{ID: 1, Text: tt.text, Active: true}
			err := req.ValidateAndNormalize()
			if tt.wantMessage == "" {
				if err != nil {
					t.Errorf("err = %v, want none", err)
				}
				return
			}
			if got := appMessage(t, err); got != tt.wantMessage {
				t.Errorf("message = %q, want %q", got, tt.wantMessage)
			}
		})
	}
}

func TestUpdateRunningTextSetting(t *testing.T) {
	for _, speed := range []ScrollSpeed{ScrollSpeedSlow, ScrollSpeedMedium, ScrollSpeedFast} {
		req := UpdateRunningTextSettingRequest{RunningTextSetting: RunningTextSetting{Speed: speed}}
		if err := req.ValidateAndNormalize(); err != nil {
			t.Errorf("%s: %v", speed, err)
		}
	}

	req := UpdateRunningTextSettingRequest{RunningTextSetting: RunningTextSetting{Speed: "turbo"}}
	if got := appMessage(t, req.ValidateAndNormalize()); got != "kecepatan gulir tidak valid" {
		t.Errorf("message = %q", got)
	}
}
