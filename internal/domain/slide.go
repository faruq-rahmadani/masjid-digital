package domain

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

// Slides rotate in the TV's main panel, active ones only, in position order, each for its own
// duration. Every screen shows the same slides.

var ErrSlideNotFound = errors.New("slide not found")

const (
	// MaxSlideImageSize is the largest slide image accepted, in bytes
	MaxSlideImageSize     = 5 << 20 // 5 MB
	MaxSlideTitleLength   = 100
	MaxSlideCaptionLength = 200
	// DefaultSlideDuration is how long a new slide stays on screen, in seconds
	DefaultSlideDuration = 10
)

// slideDurations are the durations an admin can pick, in seconds
var slideDurations = map[int]bool{5: true, 10: true, 15: true, 20: true, 30: true}

// slideImageTypes are the accepted image formats, as detected from the file's bytes
var slideImageTypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/webp": true,
}

type Slide struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	Caption         string    `json:"caption"` // optional
	DurationSeconds int       `json:"duration_seconds"`
	ShowCaption     bool      `json:"show_caption"` // whether the TV shows the title and caption
	Active          bool      `json:"active"`
	Position        int       `json:"position"`
	CreatedBy       int64     `json:"created_by"`      // 0 when nobody created it
	CreatedByName   string    `json:"created_by_name"` // "" when nobody created it
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (s *Slide) IsActive() bool {
	return s.Active
}

func (s *Slide) ToDisplaySlide() DisplaySlide {
	return DisplaySlide{
		ID:              s.ID,
		Title:           s.Title,
		Caption:         s.Caption,
		DurationSeconds: s.DurationSeconds,
		ShowCaption:     s.ShowCaption,
	}
}

// SlideImage is a slide's picture. It never changes: a new picture is a new slide.
type SlideImage struct {
	ContentType string
	Data        []byte
}

// CreateSlideRequest adds an active slide from an uploaded image. The controller binds the
// tagged fields from the multipart form; the untagged ones are set by the server.
type CreateSlideRequest struct {
	CreatedBy int64
	FileName  string
	Data      []byte

	// Title is optional: it defaults to FileName without its extension
	Title   string `form:"title"`
	Caption string `form:"caption"` // optional
	// DurationSeconds is optional: 0 means DefaultSlideDuration
	DurationSeconds int  `form:"duration_seconds"`
	ShowCaption     bool `form:"show_caption"`
}

// ValidateAndDetect normalizes the fields and checks the image's size and format, returning the
// image with its content type detected from the bytes, whatever the client claimed
func (r *CreateSlideRequest) ValidateAndDetect() (SlideImage, error) {
	r.Title = strings.TrimSpace(r.Title)
	if r.Title == "" {
		r.Title = truncateRunes(strings.TrimSpace(strings.TrimSuffix(r.FileName, filepath.Ext(r.FileName))), MaxSlideTitleLength)
	}
	if r.DurationSeconds == 0 {
		r.DurationSeconds = DefaultSlideDuration
	}

	var err error
	if r.Title, err = validateSlideTitle(r.Title); err != nil {
		return SlideImage{}, err
	}
	if r.Caption, err = validateSlideCaption(r.Caption); err != nil {
		return SlideImage{}, err
	}
	if err := validateSlideDuration(r.DurationSeconds); err != nil {
		return SlideImage{}, err
	}

	if len(r.Data) == 0 {
		return SlideImage{}, apperror.New(apperror.CodeInvalidInput, errors.New("empty slide image"), "file gambar wajib diunggah")
	}

	if len(r.Data) > MaxSlideImageSize {
		return SlideImage{}, apperror.New(apperror.CodeInvalidInput, fmt.Errorf("slide image is %d bytes, more than %d", len(r.Data), MaxSlideImageSize), "ukuran gambar maksimal 5 MB")
	}

	contentType := http.DetectContentType(r.Data)
	if !slideImageTypes[contentType] {
		return SlideImage{}, apperror.New(apperror.CodeInvalidInput, fmt.Errorf("slide image type %q not allowed", contentType), "format gambar harus PNG, JPG, atau WebP")
	}

	return SlideImage{ContentType: contentType, Data: r.Data}, nil
}

type UpdateSlideRequest struct {
	ID int64 `json:"-"`

	Title           string `json:"title"`
	Caption         string `json:"caption"` // empty clears it
	DurationSeconds int    `json:"duration_seconds"`
	ShowCaption     bool   `json:"show_caption"`
	Active          bool   `json:"active"`
}

func (r *UpdateSlideRequest) ValidateAndNormalize() error {
	var err error
	if r.Title, err = validateSlideTitle(r.Title); err != nil {
		return err
	}

	if r.Caption, err = validateSlideCaption(r.Caption); err != nil {
		return err
	}
	return validateSlideDuration(r.DurationSeconds)
}

// ReorderSlidesRequest lists every slide ID in the new display order
type ReorderSlidesRequest struct {
	IDs []int64 `json:"ids"`
}

// DisplaySlide is a slide as the TV shows it; the image is at /api/public/v1/slides/{id}/image
type DisplaySlide struct {
	ID              int64  `json:"id"`
	Title           string `json:"title"`
	Caption         string `json:"caption"`
	DurationSeconds int    `json:"duration_seconds"`
	ShowCaption     bool   `json:"show_caption"`
}

func validateSlideTitle(s string) (string, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return "", apperror.New(apperror.CodeInvalidInput, errors.New("empty slide title"), "judul wajib diisi")
	}
	if utf8.RuneCountInString(trimmed) > MaxSlideTitleLength {
		return "", apperror.New(apperror.CodeInvalidInput, fmt.Errorf("slide title longer than %d characters", MaxSlideTitleLength), fmt.Sprintf("judul maksimal %d karakter", MaxSlideTitleLength))
	}
	return trimmed, nil
}

func validateSlideCaption(s string) (string, error) {
	trimmed := strings.TrimSpace(s)
	if utf8.RuneCountInString(trimmed) > MaxSlideCaptionLength {
		return "", apperror.New(apperror.CodeInvalidInput, fmt.Errorf("slide caption longer than %d characters", MaxSlideCaptionLength), fmt.Sprintf("keterangan maksimal %d karakter", MaxSlideCaptionLength))
	}
	return trimmed, nil
}

func validateSlideDuration(seconds int) error {
	if !slideDurations[seconds] {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("slide duration %d not allowed", seconds), "durasi harus 5, 10, 15, 20, atau 30 detik")
	}
	return nil
}

// truncateRunes cuts s to at most n characters
func truncateRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return strings.TrimSpace(string([]rune(s)[:n]))
}

// ToActivityLog is the log of creating the slide that got id; call after ValidateAndDetect
func (r *CreateSlideRequest) ToActivityLog(id int64) ActivityLog {
	return ActivityLog{Action: ActivitySlideCreated, EntityID: id, Summary: "menambah slide " + quoteForSummary(r.Title)}
}

// ToActivityLog is the log of this update; call after ValidateAndNormalize
func (r *UpdateSlideRequest) ToActivityLog() ActivityLog {
	return ActivityLog{Action: ActivitySlideUpdated, EntityID: r.ID, Summary: "mengubah slide " + quoteForSummary(r.Title)}
}

func (r *ReorderSlidesRequest) ToActivityLog() ActivityLog {
	return ActivityLog{Action: ActivitySlidesReordered, Summary: "mengurutkan ulang slide"}
}

func SlideDeletedActivityLog(id int64) ActivityLog {
	return ActivityLog{Action: ActivitySlideDeleted, EntityID: id, Summary: "menghapus slide"}
}
