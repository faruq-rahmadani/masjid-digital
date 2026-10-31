package domain

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

// The mosque's profile and logo are stored once per deployment, since each one
// serves a single mosque. Each is missing until first saved.
var (
	ErrMosqueProfileNotFound = errors.New("mosque profile not saved yet")
	ErrMosqueLogoNotFound    = errors.New("mosque logo not uploaded yet")
)

// MosqueTimezone is one of Indonesia's three time zones, as an IANA name
type MosqueTimezone string

const (
	MosqueTimezoneWIB  MosqueTimezone = "Asia/Jakarta"
	MosqueTimezoneWITA MosqueTimezone = "Asia/Makassar"
	MosqueTimezoneWIT  MosqueTimezone = "Asia/Jayapura"
)

func (tz MosqueTimezone) IsValid() bool {
	switch tz {
	case MosqueTimezoneWIB, MosqueTimezoneWITA, MosqueTimezoneWIT:
		return true
	default:
		return false
	}
}

// MosqueProfile is who and where the mosque is. Prayer times are computed from
// its coordinates and time zone. It is empty until first saved.
type MosqueProfile struct {
	Name         string         `json:"name"`
	Address      string         `json:"address"`       // optional
	ContactPhone string         `json:"contact_phone"` // optional, E.164
	Latitude     float64        `json:"latitude"`
	Longitude    float64        `json:"longitude"`
	Timezone     MosqueTimezone `json:"timezone"`
}

// UpdateMosqueProfileRequest replaces the whole profile (PUT)
type UpdateMosqueProfileRequest struct {
	// UpdatedBy is the logged in user making the change; the controller sets it from the session
	UpdatedBy int64 `json:"-"`

	Name         string         `json:"name"`
	Address      string         `json:"address"`       // empty clears it
	ContactPhone string         `json:"contact_phone"` // empty clears it
	Latitude     float64        `json:"latitude"`
	Longitude    float64        `json:"longitude"`
	Timezone     MosqueTimezone `json:"timezone"`
}

func (r *UpdateMosqueProfileRequest) ValidateAndNormalize() error {
	const maxAddressLength = 500

	var err error

	if r.Name, err = validateName(r.Name); err != nil {
		return err
	}

	r.Address = strings.TrimSpace(r.Address)
	if utf8.RuneCountInString(r.Address) > maxAddressLength {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("address must be less than %d characters long", maxAddressLength), fmt.Sprintf("alamat maksimal %d karakter", maxAddressLength))
	}

	if strings.TrimSpace(r.ContactPhone) == "" {
		r.ContactPhone = ""
	} else if r.ContactPhone, err = validatePhone(r.ContactPhone); err != nil {
		return apperror.New(apperror.CodeInvalidInput, err, "kontak pengurus harus nomor HP yang valid")
	}

	// 0,0 is in the Atlantic: it means the client never picked a location
	if r.Latitude == 0 && r.Longitude == 0 {
		return apperror.New(apperror.CodeInvalidInput, errors.New("missing location"), "lokasi masjid wajib di isi")
	}
	if r.Latitude < -90 || r.Latitude > 90 || r.Longitude < -180 || r.Longitude > 180 {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("coordinates out of range: %f, %f", r.Latitude, r.Longitude), "koordinat lokasi tidak valid")
	}

	if !r.Timezone.IsValid() {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("unknown timezone %q", r.Timezone), "zona waktu harus WIB, WITA, atau WIT")
	}

	return nil
}

// ToMosqueProfile returns the profile the request describes
func (r *UpdateMosqueProfileRequest) ToMosqueProfile() MosqueProfile {
	return MosqueProfile{
		Name:         r.Name,
		Address:      r.Address,
		ContactPhone: r.ContactPhone,
		Latitude:     r.Latitude,
		Longitude:    r.Longitude,
		Timezone:     r.Timezone,
	}
}

// MaxMosqueLogoSize is the largest logo accepted, in bytes
const MaxMosqueLogoSize = 2 << 20 // 2 MB

// mosqueLogoTypes are the accepted logo formats, as detected from the file's
// content. SVG is left out: it can carry scripts and has no magic bytes.
var mosqueLogoTypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/webp": true,
}

// MosqueLogo is the mosque's logo image
type MosqueLogo struct {
	ContentType string
	Data        []byte
}

// DataURI returns the logo as a data: URI, usable directly as an <img> src
func (l MosqueLogo) DataURI() string {
	return "data:" + l.ContentType + ";base64," + base64.StdEncoding.EncodeToString(l.Data)
}

// UpdateMosqueLogoRequest replaces the logo with an uploaded image
type UpdateMosqueLogoRequest struct {
	// UpdatedBy is the logged in user making the change; the controller sets it from the session
	UpdatedBy int64
	Data      []byte
}

// ValidateAndDetect checks the image's size and format, and returns the logo
// with its content type detected from the bytes, whatever the client claimed
func (r *UpdateMosqueLogoRequest) ValidateAndDetect() (MosqueLogo, error) {
	if len(r.Data) == 0 {
		return MosqueLogo{}, apperror.New(apperror.CodeInvalidInput, errors.New("empty logo file"), "file logo wajib diunggah")
	}

	if len(r.Data) > MaxMosqueLogoSize {
		return MosqueLogo{}, apperror.New(apperror.CodeInvalidInput, fmt.Errorf("logo is %d bytes, more than %d", len(r.Data), MaxMosqueLogoSize), "ukuran logo maksimal 2 MB")
	}

	contentType := http.DetectContentType(r.Data)
	if !mosqueLogoTypes[contentType] {
		return MosqueLogo{}, apperror.New(apperror.CodeInvalidInput, fmt.Errorf("logo type %q not allowed", contentType), "format logo harus PNG, JPG, atau WebP")
	}

	return MosqueLogo{ContentType: contentType, Data: r.Data}, nil
}
