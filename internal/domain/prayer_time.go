package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

// Prayer times are computed on each device from the mosque's location with the Kemenag method
// (Subuh 20°, Isya 18°); the server only stores how a mosque adjusts them.

var (
	ErrPrayerSettingsNotFound     = errors.New("prayer settings not saved yet")
	ErrPrayerTimeOverrideNotFound = errors.New("prayer time override not found")
	ErrPrayerTimeOverrideExists   = errors.New("prayer already has an override on that date")
)

type Prayer string

const (
	PrayerSubuh   Prayer = "subuh"
	PrayerDzuhur  Prayer = "dzuhur"
	PrayerAshar   Prayer = "ashar"
	PrayerMaghrib Prayer = "maghrib"
	PrayerIsya    Prayer = "isya"
)

func (p Prayer) IsValid() bool {
	switch p {
	case PrayerSubuh, PrayerDzuhur, PrayerAshar, PrayerMaghrib, PrayerIsya:
		return true
	default:
		return false
	}
}

type Weekday string

var weekdays = map[Weekday]bool{
	"monday": true, "tuesday": true, "wednesday": true, "thursday": true,
	"friday": true, "saturday": true, "sunday": true,
}

func (w Weekday) IsValid() bool {
	return weekdays[w]
}

const (
	maxPrayerCorrection = 10 // minutes either way
	minImsakOffset      = 1
	maxImsakOffset      = 30
	maxHijriOffset      = 2 // days either way
	maxRemarksLength    = 100

	DefaultImsakOffsetMin = 10
)

type PrayerCorrections struct {
	Subuh   int `json:"subuh"`
	Terbit  int `json:"terbit"`
	Dzuhur  int `json:"dzuhur"`
	Ashar   int `json:"ashar"`
	Maghrib int `json:"maghrib"`
	Isya    int `json:"isya"`
}

type WeeklyPrayerOverride struct {
	Weekday Weekday `json:"weekday"`
	Prayer  Prayer  `json:"prayer"`
	Time    string  `json:"time"` // HH:MM, local
	Remarks string  `json:"remarks"`
}

type PrayerSettings struct {
	Corrections     PrayerCorrections      `json:"corrections"`
	ImsakOffsetMin  int                    `json:"imsak_offset_min"`
	WeeklyOverrides []WeeklyPrayerOverride `json:"weekly_overrides"`
	// HijriOffsetDays shifts the computed Hijri date, e.g. when the sidang isbat starts a month
	// a day later. It applies from HijriOffsetFrom (YYYY-MM-DD), or always when that is empty.
	HijriOffsetDays int    `json:"hijri_offset_days"`
	HijriOffsetFrom string `json:"hijri_offset_from"`
}

func DefaultPrayerSettings() PrayerSettings {
	return PrayerSettings{
		ImsakOffsetMin:  DefaultImsakOffsetMin,
		WeeklyOverrides: []WeeklyPrayerOverride{},
	}
}

type UpdatePrayerSettingsRequest struct {
	UpdatedBy int64 `json:"-"`

	PrayerSettings
}

func (r *UpdatePrayerSettingsRequest) ValidateAndNormalize() error {
	c := r.Corrections
	for _, v := range []int{c.Subuh, c.Terbit, c.Dzuhur, c.Ashar, c.Maghrib, c.Isya} {
		if v < -maxPrayerCorrection || v > maxPrayerCorrection {
			return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("correction %d out of range", v), fmt.Sprintf("koreksi waktu harus antara -%d dan +%d menit", maxPrayerCorrection, maxPrayerCorrection))
		}
	}

	if r.ImsakOffsetMin < minImsakOffset || r.ImsakOffsetMin > maxImsakOffset {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("imsak offset %d out of range", r.ImsakOffsetMin), fmt.Sprintf("imsak harus %d sampai %d menit sebelum subuh", minImsakOffset, maxImsakOffset))
	}

	if r.WeeklyOverrides == nil {
		r.WeeklyOverrides = []WeeklyPrayerOverride{}
	}
	seen := make(map[string]bool, len(r.WeeklyOverrides))
	for i := range r.WeeklyOverrides {
		o := &r.WeeklyOverrides[i]
		if !o.Weekday.IsValid() {
			return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("unknown weekday %q", o.Weekday), "hari tidak valid")
		}
		if !o.Prayer.IsValid() {
			return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("unknown prayer %q", o.Prayer), "waktu sholat tidak valid")
		}
		var err error
		if o.Time, err = validateClock(o.Time); err != nil {
			return err
		}
		if o.Remarks, err = validateRemarks(o.Remarks); err != nil {
			return err
		}
		key := string(o.Weekday) + "/" + string(o.Prayer)
		if seen[key] {
			return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("duplicate weekly override %s", key), "satu waktu sholat hanya boleh punya satu jadwal mingguan per hari")
		}
		seen[key] = true
	}

	if r.HijriOffsetDays < -maxHijriOffset || r.HijriOffsetDays > maxHijriOffset {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("hijri offset %d out of range", r.HijriOffsetDays), fmt.Sprintf("koreksi tanggal hijriah harus antara -%d dan +%d hari", maxHijriOffset, maxHijriOffset))
	}
	r.HijriOffsetFrom = strings.TrimSpace(r.HijriOffsetFrom)
	if r.HijriOffsetFrom != "" {
		if _, err := parseDate(r.HijriOffsetFrom); err != nil {
			return err
		}
	}

	return nil
}

type PrayerTimeOverride struct {
	ID        int64     `json:"id"`
	Date      string    `json:"date"` // YYYY-MM-DD, local
	Prayer    Prayer    `json:"prayer"`
	Time      string    `json:"time"` // HH:MM, local
	Remarks   string    `json:"remarks"`
	CreatedBy int64     `json:"created_by"` // 0 when nobody created it
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreatePrayerTimeOverrideRequest struct {
	CreatedBy int64 `json:"-"`

	Date    string `json:"date"`
	Prayer  Prayer `json:"prayer"`
	Time    string `json:"time"`
	Remarks string `json:"remarks"`
}

func (r *CreatePrayerTimeOverrideRequest) ValidateAndNormalize() error {
	var err error
	if r.Date, err = normalizeDate(r.Date); err != nil {
		return err
	}
	if !r.Prayer.IsValid() {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("unknown prayer %q", r.Prayer), "waktu sholat tidak valid")
	}
	if r.Time, err = validateClock(r.Time); err != nil {
		return err
	}
	if r.Remarks, err = validateRemarks(r.Remarks); err != nil {
		return err
	}
	return nil
}

func (r *CreatePrayerTimeOverrideRequest) ToPrayerTimeOverride() PrayerTimeOverride {
	return PrayerTimeOverride{Date: r.Date, Prayer: r.Prayer, Time: r.Time, Remarks: r.Remarks, CreatedBy: r.CreatedBy}
}

const MaxPrayerOverrideRange = 366

type PrayerTimeOverrideRange struct {
	From string
	To   string
}

func (r *PrayerTimeOverrideRange) ValidateAndNormalize() error {
	from, err := parseDate(strings.TrimSpace(r.From))
	if err != nil {
		return err
	}
	to, err := parseDate(strings.TrimSpace(r.To))
	if err != nil {
		return err
	}
	if to.Before(from) {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("range ends %s before it starts %s", r.To, r.From), "tanggal akhir harus setelah tanggal awal")
	}
	if to.Sub(from) > MaxPrayerOverrideRange*24*time.Hour {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("range of %s is too long", to.Sub(from)), fmt.Sprintf("rentang tanggal maksimal %d hari", MaxPrayerOverrideRange))
	}
	r.From, r.To = from.Format(dateLayout), to.Format(dateLayout)
	return nil
}

type PrayerConfig struct {
	Settings  PrayerSettings       `json:"settings"`
	Overrides []PrayerTimeOverride `json:"overrides"`
}

const (
	dateLayout  = "2006-01-02"
	clockLayout = "15:04"
)

func parseDate(s string) (time.Time, error) {
	d, err := time.Parse(dateLayout, s)
	if err != nil {
		return time.Time{}, apperror.New(apperror.CodeInvalidInput, fmt.Errorf("invalid date %q: %w", s, err), "format tanggal harus YYYY-MM-DD")
	}
	return d, nil
}

func normalizeDate(s string) (string, error) {
	d, err := parseDate(strings.TrimSpace(s))
	if err != nil {
		return "", err
	}
	return d.Format(dateLayout), nil
}

func validateClock(s string) (string, error) {
	t, err := time.Parse(clockLayout, strings.TrimSpace(s))
	if err != nil {
		return "", apperror.New(apperror.CodeInvalidInput, fmt.Errorf("invalid time %q: %w", s, err), "format jam harus HH:MM")
	}
	return t.Format(clockLayout), nil
}

func validateRemarks(s string) (string, error) {
	trimmed := strings.TrimSpace(s)
	if utf8.RuneCountInString(trimmed) > maxRemarksLength {
		return "", apperror.New(apperror.CodeInvalidInput, fmt.Errorf("remarks longer than %d characters", maxRemarksLength), fmt.Sprintf("keterangan maksimal %d karakter", maxRemarksLength))
	}
	return trimmed, nil
}
