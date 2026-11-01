package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

// Running texts scroll in the TV's INFO bar, active ones only, in position order.

var (
	ErrRunningTextNotFound        = errors.New("running text not found")
	ErrRunningTextSettingNotFound = errors.New("running text setting not saved yet")
)

const MaxRunningTextLength = 200

type ScrollSpeed string

const (
	ScrollSpeedSlow   ScrollSpeed = "slow"
	ScrollSpeedMedium ScrollSpeed = "medium"
	ScrollSpeedFast   ScrollSpeed = "fast"
)

func (s ScrollSpeed) IsValid() bool {
	switch s {
	case ScrollSpeedSlow, ScrollSpeedMedium, ScrollSpeedFast:
		return true
	default:
		return false
	}
}

type RunningText struct {
	ID            int64  `json:"id"`
	Text          string `json:"text"`
	Active        bool   `json:"active"`
	Position      int    `json:"position"`
	CreatedBy     int64  `json:"created_by"`      // 0 when nobody created it
	CreatedByName string `json:"created_by_name"` // "" when nobody created it
	// DeactivatedAt is when the text was last turned off, i.e. last shown on the TV; nil while
	// active or if never turned off
	DeactivatedAt *time.Time `json:"deactivated_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type CreateRunningTextRequest struct {
	CreatedBy int64 `json:"-"`

	Text string `json:"text"`
}

func (r *CreateRunningTextRequest) ValidateAndNormalize() error {
	var err error
	r.Text, err = validateRunningText(r.Text)
	return err
}

type UpdateRunningTextRequest struct {
	ID int64 `json:"-"`

	Text   string `json:"text"`
	Active bool   `json:"active"`
}

func (r *UpdateRunningTextRequest) ValidateAndNormalize() error {
	var err error
	r.Text, err = validateRunningText(r.Text)
	return err
}

// ReorderRunningTextsRequest lists every running text ID in the new display order
type ReorderRunningTextsRequest struct {
	IDs []int64 `json:"ids"`
}

type RunningTextSetting struct {
	Speed ScrollSpeed `json:"speed"`
}

func DefaultRunningTextSetting() RunningTextSetting {
	return RunningTextSetting{Speed: ScrollSpeedMedium}
}

type UpdateRunningTextSettingRequest struct {
	UpdatedBy int64 `json:"-"`

	RunningTextSetting
}

func (r *UpdateRunningTextSettingRequest) ValidateAndNormalize() error {
	if !r.Speed.IsValid() {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("unknown scroll speed %q", r.Speed), "kecepatan gulir tidak valid")
	}
	return nil
}

// RunningTextDisplay is what the TV shows: the active texts in order and how fast they scroll
type RunningTextDisplay struct {
	Speed ScrollSpeed `json:"speed"`
	Texts []string    `json:"texts"`
}

func validateRunningText(s string) (string, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return "", apperror.New(apperror.CodeInvalidInput, errors.New("empty running text"), "teks wajib diisi")
	}
	if utf8.RuneCountInString(trimmed) > MaxRunningTextLength {
		return "", apperror.New(apperror.CodeInvalidInput, fmt.Errorf("running text longer than %d characters", MaxRunningTextLength), fmt.Sprintf("teks maksimal %d karakter", MaxRunningTextLength))
	}
	return trimmed, nil
}
