package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// Who leads each Friday's prayer (petugas Jumat). The TV shows them on the Jumat screen; when
// that screen runs is in iqamah.go.

var ErrFridayDutyNotFound = errors.New("friday duty not found")

const (
	maxDutyNameLength  = 100
	maxDutyThemeLength = 200
)

// FridayDuty is who leads one Friday's prayer (petugas Jumat)
type FridayDuty struct {
	Date      string    `json:"date"` // YYYY-MM-DD, a Friday, local
	Khatib    string    `json:"khatib"`
	Imam      string    `json:"imam"`
	Muadzin   string    `json:"muadzin"`
	Theme     string    `json:"theme"` // tema khutbah
	UpdatedBy int64     `json:"updated_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SaveFridayDutyRequest sets the petugas of one Friday, replacing what was there (PUT by date)
type SaveFridayDutyRequest struct {
	// UpdatedBy is the logged in user; the controller sets it from the session
	UpdatedBy int64 `json:"-"`
	// Date comes from the path; the controller sets it
	Date string `json:"-"`

	Khatib  string `json:"khatib"`
	Imam    string `json:"imam"`
	Muadzin string `json:"muadzin"`
	Theme   string `json:"theme"`
}

func (r *SaveFridayDutyRequest) ValidateAndNormalize() error {
	var err error
	if r.Date, err = normalizeDate(r.Date); err != nil {
		return err
	}
	date, _ := time.Parse(dateLayout, r.Date)
	if date.Weekday() != time.Friday {
		return invalid("not a friday "+r.Date, "petugas Jumat hanya untuk hari Jumat")
	}

	r.Khatib, r.Imam, r.Muadzin = strings.TrimSpace(r.Khatib), strings.TrimSpace(r.Imam), strings.TrimSpace(r.Muadzin)
	r.Theme = strings.TrimSpace(r.Theme)
	for _, name := range []string{r.Khatib, r.Imam, r.Muadzin} {
		if utf8.RuneCountInString(name) > maxDutyNameLength {
			return invalid("name too long", fmt.Sprintf("nama petugas maksimal %d karakter", maxDutyNameLength))
		}
	}
	if utf8.RuneCountInString(r.Theme) > maxDutyThemeLength {
		return invalid("theme too long", fmt.Sprintf("tema khutbah maksimal %d karakter", maxDutyThemeLength))
	}
	if r.Khatib == "" && r.Imam == "" && r.Muadzin == "" {
		return invalid("no names", "isi minimal satu petugas")
	}
	return nil
}

// ToFridayDuty returns the duty the request describes
func (r *SaveFridayDutyRequest) ToFridayDuty() FridayDuty {
	return FridayDuty{Date: r.Date, Khatib: r.Khatib, Imam: r.Imam, Muadzin: r.Muadzin, Theme: r.Theme, UpdatedBy: r.UpdatedBy}
}

// FridayDutyRange is the dates to list petugas Jumat for, from..to inclusive (YYYY-MM-DD)
type FridayDutyRange struct {
	From string
	To   string
}

// ValidateAndNormalize applies the same rules as any date range (see PrayerTimeOverrideRange)
func (r *FridayDutyRange) ValidateAndNormalize() error {
	dates := PrayerTimeOverrideRange(*r)
	if err := dates.ValidateAndNormalize(); err != nil {
		return err
	}
	*r = FridayDutyRange(dates)
	return nil
}
