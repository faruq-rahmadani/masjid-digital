package iqamahsettings

import "github.com/faruq-rahmadani/masjid-digital/internal/domain"

// settingsValue is the stored JSON of the iqamah settings. It stays inside this package; other
// layers work with domain.IqamahSettings.
type settingsValue struct {
	IqamahMin       minutesValue `json:"iqamah_min"`
	SholatMin       minutesValue `json:"sholat_min"`
	SholatMinJumat  int          `json:"sholat_min_jumat"`
	AdzanDisplayMin int          `json:"adzan_display_min"`
	SholatDisplay   string       `json:"sholat_display"`
	JumatBeforeMin  int          `json:"jumat_before_min"`
	KhutbahMin      int          `json:"khutbah_min"`
	Alarm           alarmValue   `json:"alarm"`
}

type minutesValue struct {
	Subuh   int `json:"subuh"`
	Dzuhur  int `json:"dzuhur"`
	Ashar   int `json:"ashar"`
	Maghrib int `json:"maghrib"`
	Isya    int `json:"isya"`
}

type alarmValue struct {
	Enabled      bool   `json:"enabled"`
	AtAdzan      bool   `json:"at_adzan"`
	BeforeIqamah bool   `json:"before_iqamah"`
	Sound        string `json:"sound"`
	Volume       int    `json:"volume"`
}

func newSettingsValue(s domain.IqamahSettings) settingsValue {
	return settingsValue{
		IqamahMin:       minutesValue(s.IqamahMin),
		SholatMin:       minutesValue(s.SholatMin.PrayerMinutes),
		SholatMinJumat:  s.SholatMin.Jumat,
		AdzanDisplayMin: s.AdzanDisplayMin,
		SholatDisplay:   string(s.SholatDisplay),
		JumatBeforeMin:  s.JumatBeforeMin,
		KhutbahMin:      s.KhutbahMin,
		Alarm: alarmValue{
			Enabled:      s.Alarm.Enabled,
			AtAdzan:      s.Alarm.AtAdzan,
			BeforeIqamah: s.Alarm.BeforeIqamah,
			Sound:        string(s.Alarm.Sound),
			Volume:       s.Alarm.Volume,
		},
	}
}

func (v settingsValue) toDomain() domain.IqamahSettings {
	return domain.IqamahSettings{
		IqamahMin:       domain.PrayerMinutes(v.IqamahMin),
		SholatMin:       domain.SholatMinutes{PrayerMinutes: domain.PrayerMinutes(v.SholatMin), Jumat: v.SholatMinJumat},
		AdzanDisplayMin: v.AdzanDisplayMin,
		SholatDisplay:   domain.SholatDisplay(v.SholatDisplay),
		JumatBeforeMin:  v.JumatBeforeMin,
		KhutbahMin:      v.KhutbahMin,
		Alarm: domain.AlarmSettings{
			Enabled:      v.Alarm.Enabled,
			AtAdzan:      v.Alarm.AtAdzan,
			BeforeIqamah: v.Alarm.BeforeIqamah,
			Sound:        domain.AlarmSound(v.Alarm.Sound),
			Volume:       v.Alarm.Volume,
		},
	}
}
