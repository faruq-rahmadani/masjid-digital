package domain

import (
	"errors"
	"fmt"

	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

// What the TV does around each prayer (E5): after the adzan time it shows the adzan screen, counts
// down to iqamah, then dims for the prayer. On Friday, Dzuhur is Jumat: a Jumat screen before it,
// the khutbah after the adzan, then the prayer. Who leads it is in friday_duty.go.

var ErrIqamahSettingsNotFound = errors.New("iqamah settings not saved yet")

// SholatDisplay is how the TV looks while a prayer is on
type SholatDisplay string

const (
	SholatDisplayDim   SholatDisplay = "dim"   // "Sholat sedang berlangsung" and a small clock
	SholatDisplayBlack SholatDisplay = "black" // a black screen
)

func (d SholatDisplay) IsValid() bool {
	return d == SholatDisplayDim || d == SholatDisplayBlack
}

// AlarmSound is the sound the TV plays for an alarm
type AlarmSound string

const (
	AlarmSoundBeep AlarmSound = "beep"
	AlarmSoundBell AlarmSound = "bell"
)

func (s AlarmSound) IsValid() bool {
	return s == AlarmSoundBeep || s == AlarmSoundBell
}

const (
	minIqamahMin       = 1
	maxIqamahMin       = 30
	minSholatMin       = 5
	maxSholatMin       = 30
	maxJumatSholatMin  = 60
	minAdzanDisplayMin = 1
	maxAdzanDisplayMin = 10
	maxJumatBeforeMin  = 120 // 0 turns the Jumat screen before Dzuhur off
	minKhutbahMin      = 10
	maxKhutbahMin      = 60
	maxVolume          = 100
)

// PrayerMinutes is a number of minutes for each of the five prayers
type PrayerMinutes struct {
	Subuh   int `json:"subuh"`
	Dzuhur  int `json:"dzuhur"`
	Ashar   int `json:"ashar"`
	Maghrib int `json:"maghrib"`
	Isya    int `json:"isya"`
}

func (m PrayerMinutes) values() []int {
	return []int{m.Subuh, m.Dzuhur, m.Ashar, m.Maghrib, m.Isya}
}

// SholatMinutes is how long mode sholat lasts per prayer, Jumat included
type SholatMinutes struct {
	PrayerMinutes
	Jumat int `json:"jumat"`
}

// AlarmSettings is when and how the TV sounds an alarm
type AlarmSettings struct {
	Enabled      bool       `json:"enabled"`
	AtAdzan      bool       `json:"at_adzan"`
	BeforeIqamah bool       `json:"before_iqamah"` // the last 10 seconds before iqamah
	Sound        AlarmSound `json:"sound"`
	Volume       int        `json:"volume"` // 0..100
}

// IqamahSettings is what the TV does around each prayer. Until first saved, the defaults apply.
type IqamahSettings struct {
	IqamahMin       PrayerMinutes `json:"iqamah_min"` // counted from the adzan time
	SholatMin       SholatMinutes `json:"sholat_min"`
	AdzanDisplayMin int           `json:"adzan_display_min"`
	SholatDisplay   SholatDisplay `json:"sholat_display"`
	JumatBeforeMin  int           `json:"jumat_before_min"` // the Jumat screen shows this long before Dzuhur
	KhutbahMin      int           `json:"khutbah_min"`
	Alarm           AlarmSettings `json:"alarm"`
}

// DefaultIqamahSettings are the settings of a mosque that never changed them, as in the mockup
func DefaultIqamahSettings() IqamahSettings {
	return IqamahSettings{
		IqamahMin: PrayerMinutes{Subuh: 15, Dzuhur: 10, Ashar: 10, Maghrib: 7, Isya: 10},
		SholatMin: SholatMinutes{
			PrayerMinutes: PrayerMinutes{Subuh: 15, Dzuhur: 12, Ashar: 12, Maghrib: 12, Isya: 15},
			Jumat:         40,
		},
		AdzanDisplayMin: 4,
		SholatDisplay:   SholatDisplayDim,
		JumatBeforeMin:  60,
		KhutbahMin:      30,
		Alarm:           AlarmSettings{Enabled: true, AtAdzan: true, BeforeIqamah: true, Sound: AlarmSoundBeep, Volume: 70},
	}
}

// UpdateIqamahSettingsRequest replaces every setting (PUT)
type UpdateIqamahSettingsRequest struct {
	// UpdatedBy is the logged in user making the change; the controller sets it from the session
	UpdatedBy int64 `json:"-"`

	IqamahSettings
}

func (r *UpdateIqamahSettingsRequest) Validate() error {
	if !allBetween(r.IqamahMin.values(), minIqamahMin, maxIqamahMin) {
		return invalid("iqamah", fmt.Sprintf("iqamah harus %d sampai %d menit", minIqamahMin, maxIqamahMin))
	}
	if !allBetween(r.SholatMin.values(), minSholatMin, maxSholatMin) {
		return invalid("sholat", fmt.Sprintf("mode sholat harus %d sampai %d menit", minSholatMin, maxSholatMin))
	}
	if !between(r.SholatMin.Jumat, minSholatMin, maxJumatSholatMin) {
		return invalid("jumat sholat", fmt.Sprintf("mode sholat Jumat harus %d sampai %d menit", minSholatMin, maxJumatSholatMin))
	}
	if !between(r.AdzanDisplayMin, minAdzanDisplayMin, maxAdzanDisplayMin) {
		return invalid("adzan display", fmt.Sprintf("durasi tampilan adzan harus %d sampai %d menit", minAdzanDisplayMin, maxAdzanDisplayMin))
	}
	if !r.SholatDisplay.IsValid() {
		return invalid("sholat display", "tampilan saat sholat harus teks redup atau gelap total")
	}
	if !between(r.JumatBeforeMin, 0, maxJumatBeforeMin) {
		return invalid("jumat before", fmt.Sprintf("layar Jumat harus 0 sampai %d menit sebelum Dzuhur", maxJumatBeforeMin))
	}
	if !between(r.KhutbahMin, minKhutbahMin, maxKhutbahMin) {
		return invalid("khutbah", fmt.Sprintf("durasi khutbah harus %d sampai %d menit", minKhutbahMin, maxKhutbahMin))
	}
	if !r.Alarm.Sound.IsValid() {
		return invalid("alarm sound", "bunyi alarm harus beep atau bel")
	}
	if !between(r.Alarm.Volume, 0, maxVolume) {
		return invalid("alarm volume", fmt.Sprintf("volume harus 0 sampai %d", maxVolume))
	}
	return nil
}

func between(v, lo, hi int) bool {
	return v >= lo && v <= hi
}

func allBetween(values []int, lo, hi int) bool {
	for _, v := range values {
		if !between(v, lo, hi) {
			return false
		}
	}
	return true
}

func invalid(internal, message string) *apperror.AppError {
	return apperror.New(apperror.CodeInvalidInput, errors.New(internal), message)
}
