package domain

import (
	"strings"
	"testing"
	"time"
)

func TestConnectScreenRequest(t *testing.T) {
	req := ConnectScreenRequest{
		DeviceID:   "2f1c9a7e-5b8d-4c3a-9e21-7d6f0b4a8c11",
		NameHint:   "  teras  ",
		AppVersion: "1.4.0",
		UserAgent:  strings.Repeat("a", 400),
	}
	if err := req.ValidateAndNormalize(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if req.NameHint != "teras" || len(req.UserAgent) != maxScreenUserAgentLen {
		t.Errorf("normalized = %+v", req)
	}

	for _, id := range []string{"", "short", "has space in it", "../../etc/passwd", strings.Repeat("a", 65)} {
		bad := ConnectScreenRequest{DeviceID: id}
		if got := appMessage(t, bad.ValidateAndNormalize()); got != "id layar tidak valid" {
			t.Errorf("device id %q: message = %q", id, got)
		}
	}
}

func TestUpdateScreenRequest(t *testing.T) {
	req := UpdateScreenRequest{Name: " TV Teras ", ForcedMode: ScreenModeDark}
	if err := req.ValidateAndNormalize(); err != nil || req.Name != "TV Teras" {
		t.Errorf("got %+v, %v", req, err)
	}

	tests := map[string]struct {
		req  UpdateScreenRequest
		want string
	}{
		"no name":   {UpdateScreenRequest{Name: "  ", ForcedMode: ScreenModeAuto}, "nama layar wajib diisi"},
		"long name": {UpdateScreenRequest{Name: strings.Repeat("a", 51), ForcedMode: ScreenModeAuto}, "nama layar maksimal 50 karakter"},
		"bad mode":  {UpdateScreenRequest{Name: "TV", ForcedMode: "off"}, "mode tampilan tidak valid"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := appMessage(t, tt.req.ValidateAndNormalize()); got != tt.want {
				t.Errorf("message = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestScreenModeAt(t *testing.T) {
	now := time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC)
	later := now.Add(time.Hour)

	screen := Screen{ForcedMode: ScreenModeDark, ForcedUntil: &later}
	if got := screen.ModeAt(now); got != ScreenModeDark {
		t.Errorf("before until = %s, want dark", got)
	}
	if got := screen.ModeAt(later); got != ScreenModeAuto {
		t.Errorf("at until = %s, want auto", got)
	}
	if got := (&Screen{ForcedMode: ScreenModeDark}).ModeAt(now); got != ScreenModeAuto {
		t.Errorf("no until = %s, want auto", got)
	}
}

func TestMosqueNextMidnight(t *testing.T) {
	// 18:30 UTC is 01:30 the next day in WIB, 02:30 in WITA
	at := time.Date(2026, 11, 2, 18, 30, 0, 0, time.UTC)

	if got, want := MosqueTimezoneWIB.NextMidnight(at), time.Date(2026, 11, 3, 17, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("WIB = %s, want %s", got.UTC(), want)
	}
	if got, want := MosqueTimezoneWIT.NextMidnight(at), time.Date(2026, 11, 3, 15, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("WIT = %s, want %s", got.UTC(), want)
	}
	if got, want := MosqueTimezone("").NextMidnight(at), MosqueTimezoneWIB.NextMidnight(at); !got.Equal(want) {
		t.Errorf("unknown zone = %s, want WIB's %s", got.UTC(), want.UTC())
	}
}
