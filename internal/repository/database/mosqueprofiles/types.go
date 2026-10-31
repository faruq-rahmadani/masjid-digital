package mosqueprofiles

import "github.com/faruq-rahmadani/masjid-digital/internal/domain"

// profileValue is the stored JSON of the mosque's profile. It stays inside this
// package; other layers work with domain.MosqueProfile.
type profileValue struct {
	Name         string  `json:"name"`
	Address      string  `json:"address"`
	ContactPhone string  `json:"contact_phone"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Timezone     string  `json:"timezone"`
}

func newProfileValue(profile domain.MosqueProfile) profileValue {
	return profileValue{
		Name:         profile.Name,
		Address:      profile.Address,
		ContactPhone: profile.ContactPhone,
		Latitude:     profile.Latitude,
		Longitude:    profile.Longitude,
		Timezone:     string(profile.Timezone),
	}
}

func (v profileValue) toDomain() domain.MosqueProfile {
	return domain.MosqueProfile{
		Name:         v.Name,
		Address:      v.Address,
		ContactPhone: v.ContactPhone,
		Latitude:     v.Latitude,
		Longitude:    v.Longitude,
		Timezone:     domain.MosqueTimezone(v.Timezone),
	}
}

// logoValue is the stored JSON of the mosque's logo; Data is base64 in JSON
type logoValue struct {
	ContentType string `json:"content_type"`
	Data        []byte `json:"data"`
}

func newLogoValue(logo domain.MosqueLogo) logoValue {
	return logoValue{ContentType: logo.ContentType, Data: logo.Data}
}

func (v logoValue) toDomain() domain.MosqueLogo {
	return domain.MosqueLogo{ContentType: v.ContentType, Data: v.Data}
}
