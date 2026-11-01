package domain

import "testing"

func TestDefaultIqamahSettingsAreValid(t *testing.T) {
	req := UpdateIqamahSettingsRequest{IqamahSettings: DefaultIqamahSettings()}
	if err := req.Validate(); err != nil {
		t.Fatalf("defaults: %v", err)
	}
}

func TestUpdateIqamahSettingsRejects(t *testing.T) {
	tests := map[string]struct {
		edit func(*IqamahSettings)
		want string
	}{
		"iqamah zero":           {func(s *IqamahSettings) { s.IqamahMin.Subuh = 0 }, "iqamah harus 1 sampai 30 menit"},
		"iqamah too long":       {func(s *IqamahSettings) { s.IqamahMin.Isya = 31 }, "iqamah harus 1 sampai 30 menit"},
		"sholat too short":      {func(s *IqamahSettings) { s.SholatMin.Maghrib = 4 }, "mode sholat harus 5 sampai 30 menit"},
		"sholat 31":             {func(s *IqamahSettings) { s.SholatMin.Dzuhur = 31 }, "mode sholat harus 5 sampai 30 menit"},
		"jumat sholat too long": {func(s *IqamahSettings) { s.SholatMin.Jumat = 61 }, "mode sholat Jumat harus 5 sampai 60 menit"},
		"adzan display":         {func(s *IqamahSettings) { s.AdzanDisplayMin = 0 }, "durasi tampilan adzan harus 1 sampai 10 menit"},
		"sholat display":        {func(s *IqamahSettings) { s.SholatDisplay = "off" }, "tampilan saat sholat harus teks redup atau gelap total"},
		"jumat before":          {func(s *IqamahSettings) { s.JumatBeforeMin = -1 }, "layar Jumat harus 0 sampai 120 menit sebelum Dzuhur"},
		"khutbah":               {func(s *IqamahSettings) { s.KhutbahMin = 5 }, "durasi khutbah harus 10 sampai 60 menit"},
		"sound":                 {func(s *IqamahSettings) { s.Alarm.Sound = "siren" }, "bunyi alarm harus beep atau bel"},
		"volume":                {func(s *IqamahSettings) { s.Alarm.Volume = 101 }, "volume harus 0 sampai 100"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			req := UpdateIqamahSettingsRequest{IqamahSettings: DefaultIqamahSettings()}
			tt.edit(&req.IqamahSettings)
			if got := appMessage(t, req.Validate()); got != tt.want {
				t.Errorf("message = %q, want %q", got, tt.want)
			}
		})
	}

	// Jumat may go past 30, the other prayers may not
	req := UpdateIqamahSettingsRequest{IqamahSettings: DefaultIqamahSettings()}
	req.SholatMin.Jumat = 60
	if err := req.Validate(); err != nil {
		t.Errorf("jumat 60: %v", err)
	}
}
