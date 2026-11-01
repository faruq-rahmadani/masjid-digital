package prayertimes

import (
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
)

type settingsValue struct {
	Corrections     correctionsValue `json:"corrections"`
	ImsakOffsetMin  int              `json:"imsak_offset_min"`
	WeeklyOverrides []weeklyValue    `json:"weekly_overrides"`
	HijriOffsetDays int              `json:"hijri_offset_days"`
	HijriOffsetFrom string           `json:"hijri_offset_from"`
}

type correctionsValue struct {
	Subuh   int `json:"subuh"`
	Terbit  int `json:"terbit"`
	Dzuhur  int `json:"dzuhur"`
	Ashar   int `json:"ashar"`
	Maghrib int `json:"maghrib"`
	Isya    int `json:"isya"`
}

type weeklyValue struct {
	Weekday string `json:"weekday"`
	Prayer  string `json:"prayer"`
	Time    string `json:"time"`
	Remarks string `json:"remarks"`
}

func newSettingsValue(s domain.PrayerSettings) settingsValue {
	weekly := make([]weeklyValue, 0, len(s.WeeklyOverrides))
	for _, o := range s.WeeklyOverrides {
		weekly = append(weekly, weeklyValue{Weekday: string(o.Weekday), Prayer: string(o.Prayer), Time: o.Time, Remarks: o.Remarks})
	}
	return settingsValue{
		Corrections:     correctionsValue(s.Corrections),
		ImsakOffsetMin:  s.ImsakOffsetMin,
		WeeklyOverrides: weekly,
		HijriOffsetDays: s.HijriOffsetDays,
		HijriOffsetFrom: s.HijriOffsetFrom,
	}
}

func (v settingsValue) toDomain() domain.PrayerSettings {
	weekly := make([]domain.WeeklyPrayerOverride, 0, len(v.WeeklyOverrides))
	for _, o := range v.WeeklyOverrides {
		weekly = append(weekly, domain.WeeklyPrayerOverride{Weekday: domain.Weekday(o.Weekday), Prayer: domain.Prayer(o.Prayer), Time: o.Time, Remarks: o.Remarks})
	}
	return domain.PrayerSettings{
		Corrections:     domain.PrayerCorrections(v.Corrections),
		ImsakOffsetMin:  v.ImsakOffsetMin,
		WeeklyOverrides: weekly,
		HijriOffsetDays: v.HijriOffsetDays,
		HijriOffsetFrom: v.HijriOffsetFrom,
	}
}

type overrideModel struct {
	ID        int64     `db:"id"`
	Date      string    `db:"date"`
	Prayer    string    `db:"prayer"`
	Time      string    `db:"time"`
	Remarks   string    `db:"remarks"`
	CreatedBy *int64    `db:"created_by"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

var overrideColumns = []string{"id", "date", "prayer", "time", "remarks", "created_by", "created_at", "updated_at"}

func (m overrideModel) toDomain() domain.PrayerTimeOverride {
	var createdBy int64
	if m.CreatedBy != nil {
		createdBy = *m.CreatedBy
	}
	return domain.PrayerTimeOverride{
		ID:        m.ID,
		Date:      m.Date,
		Prayer:    domain.Prayer(m.Prayer),
		Time:      m.Time,
		Remarks:   m.Remarks,
		CreatedBy: createdBy,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
