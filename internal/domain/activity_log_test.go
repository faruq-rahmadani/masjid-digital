package domain

import (
	"context"
	"strings"
	"testing"
)

func TestQuoteForSummary(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{in: "Kajian Tafsir", want: `"Kajian Tafsir"`},
		{in: "  Infaq \n via   QRIS ", want: `"Infaq via QRIS"`},
		{in: strings.Repeat("a", 40), want: `"` + strings.Repeat("a", 40) + `"`},
		{in: strings.Repeat("é", 41), want: `"` + strings.Repeat("é", 39) + `…"`},
	}
	for _, tt := range tests {
		if got := quoteForSummary(tt.in); got != tt.want {
			t.Errorf("quoteForSummary(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestDateForSummary(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{in: "2026-11-06", want: "6 Nov 2026"},
		{in: "2027-05-14", want: "14 Mei 2027"},
		{in: "06-11-2026", want: "06-11-2026"},
	}
	for _, tt := range tests {
		if got := dateForSummary(tt.in); got != tt.want {
			t.Errorf("dateForSummary(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestActor(t *testing.T) {
	if got := ActorFrom(context.Background()); got != 0 {
		t.Errorf("no actor = %d, want 0", got)
	}
	if got := ActorFrom(WithActor(context.Background(), 7)); got != 7 {
		t.Errorf("actor = %d, want 7", got)
	}
}

func TestScreenLabel(t *testing.T) {
	tests := []struct {
		screen Screen
		want   string
	}{
		{screen: Screen{Name: "TV Teras", NameHint: "teras"}, want: "TV Teras"},
		{screen: Screen{NameHint: "teras"}, want: "teras"},
		{screen: Screen{}, want: "baru"},
	}
	for _, tt := range tests {
		if got := tt.screen.Label(); got != tt.want {
			t.Errorf("Label(%+v) = %q, want %q", tt.screen, got, tt.want)
		}
	}
}

func TestPrayerLabel(t *testing.T) {
	if got := PrayerMaghrib.Label(); got != "Maghrib" {
		t.Errorf("label = %q, want Maghrib", got)
	}
}
