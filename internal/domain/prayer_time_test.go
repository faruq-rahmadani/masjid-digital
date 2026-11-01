package domain

import (
	"errors"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

func validSettingsRequest() UpdatePrayerSettingsRequest {
	return UpdatePrayerSettingsRequest{
		UpdatedBy: 7,
		PrayerSettings: PrayerSettings{
			Corrections:     PrayerCorrections{Subuh: 2, Dzuhur: 2, Ashar: 2, Maghrib: 2, Isya: 2},
			ImsakOffsetMin:  10,
			WeeklyOverrides: []WeeklyPrayerOverride{{Weekday: "friday", Prayer: PrayerDzuhur, Time: " 12:00 ", Remarks: " Sholat Jumat "}},
			HijriOffsetDays: -1,
			HijriOffsetFrom: "2027-02-08",
		},
	}
}

func appMessage(t *testing.T, err error) string {
	t.Helper()
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("err = %v, want *apperror.AppError", err)
	}
	return appErr.MessageError
}

func TestUpdatePrayerSettingsNormalizes(t *testing.T) {
	req := validSettingsRequest()
	if err := req.ValidateAndNormalize(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if got := req.WeeklyOverrides[0]; got.Time != "12:00" || got.Remarks != "Sholat Jumat" {
		t.Errorf("weekly override = %+v, want trimmed", got)
	}

	empty := UpdatePrayerSettingsRequest{PrayerSettings: PrayerSettings{ImsakOffsetMin: 10}}
	if err := empty.ValidateAndNormalize(); err != nil {
		t.Fatalf("validate empty: %v", err)
	}
	if empty.WeeklyOverrides == nil {
		t.Error("nil weekly overrides should become an empty list")
	}
}

func TestUpdatePrayerSettingsRejects(t *testing.T) {
	tests := map[string]struct {
		edit func(*UpdatePrayerSettingsRequest)
		want string
	}{
		"correction too big":  {func(r *UpdatePrayerSettingsRequest) { r.Corrections.Maghrib = 11 }, "koreksi waktu harus antara -10 dan +10 menit"},
		"correction too low":  {func(r *UpdatePrayerSettingsRequest) { r.Corrections.Terbit = -11 }, "koreksi waktu harus antara -10 dan +10 menit"},
		"imsak zero":          {func(r *UpdatePrayerSettingsRequest) { r.ImsakOffsetMin = 0 }, "imsak harus 1 sampai 30 menit sebelum subuh"},
		"unknown weekday":     {func(r *UpdatePrayerSettingsRequest) { r.WeeklyOverrides[0].Weekday = "jumat" }, "hari tidak valid"},
		"terbit not override": {func(r *UpdatePrayerSettingsRequest) { r.WeeklyOverrides[0].Prayer = "terbit" }, "waktu sholat tidak valid"},
		"bad time":            {func(r *UpdatePrayerSettingsRequest) { r.WeeklyOverrides[0].Time = "25:00" }, "format jam harus HH:MM"},
		"duplicate weekly": {func(r *UpdatePrayerSettingsRequest) {
			r.WeeklyOverrides = append(r.WeeklyOverrides, r.WeeklyOverrides[0])
		}, "satu waktu sholat hanya boleh punya satu jadwal mingguan per hari"},
		"hijri too far":  {func(r *UpdatePrayerSettingsRequest) { r.HijriOffsetDays = 3 }, "koreksi tanggal hijriah harus antara -2 dan +2 hari"},
		"hijri bad date": {func(r *UpdatePrayerSettingsRequest) { r.HijriOffsetFrom = "08-02-2027" }, "format tanggal harus YYYY-MM-DD"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			req := validSettingsRequest()
			tt.edit(&req)
			if got := appMessage(t, req.ValidateAndNormalize()); got != tt.want {
				t.Errorf("message = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCreatePrayerTimeOverride(t *testing.T) {
	req := CreatePrayerTimeOverrideRequest{Date: " 2026-11-06 ", Prayer: PrayerDzuhur, Time: "11:58", Remarks: " Ikut jadwal Jumat "}
	if err := req.ValidateAndNormalize(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if req.Date != "2026-11-06" || req.Remarks != "Ikut jadwal Jumat" {
		t.Errorf("normalized = %+v", req)
	}

	bad := []CreatePrayerTimeOverrideRequest{
		{Date: "6 Nov 2026", Prayer: PrayerDzuhur, Time: "11:58"},
		{Date: "2026-11-06", Prayer: "imsak", Time: "11:58"},
		{Date: "2026-11-06", Prayer: PrayerDzuhur, Time: "noon"},
	}
	for _, b := range bad {
		if err := b.ValidateAndNormalize(); err == nil {
			t.Errorf("%+v: want error", b)
		}
	}
}

func TestPrayerTimeOverrideRange(t *testing.T) {
	ok := PrayerTimeOverrideRange{From: "2026-11-01", To: "2026-11-30"}
	if err := ok.ValidateAndNormalize(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	backwards := PrayerTimeOverrideRange{From: "2026-11-30", To: "2026-11-01"}
	if got := appMessage(t, backwards.ValidateAndNormalize()); got != "tanggal akhir harus setelah tanggal awal" {
		t.Errorf("backwards = %q", got)
	}

	tooLong := PrayerTimeOverrideRange{From: "2026-01-01", To: "2027-06-01"}
	if got := appMessage(t, tooLong.ValidateAndNormalize()); got != "rentang tanggal maksimal 366 hari" {
		t.Errorf("too long = %q", got)
	}
}
