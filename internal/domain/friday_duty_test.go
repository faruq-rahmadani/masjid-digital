package domain

import "testing"

func TestSaveFridayDuty(t *testing.T) {
	req := SaveFridayDutyRequest{Date: "2026-11-06", Khatib: " Ustadz Abdullah ", Theme: " Amanah "}
	if err := req.ValidateAndNormalize(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if req.Khatib != "Ustadz Abdullah" || req.Theme != "Amanah" {
		t.Errorf("normalized = %+v", req)
	}

	tests := map[string]struct {
		req  SaveFridayDutyRequest
		want string
	}{
		"thursday":  {SaveFridayDutyRequest{Date: "2026-11-05", Khatib: "A"}, "petugas Jumat hanya untuk hari Jumat"},
		"bad date":  {SaveFridayDutyRequest{Date: "6 Nov", Khatib: "A"}, "format tanggal harus YYYY-MM-DD"},
		"no names":  {SaveFridayDutyRequest{Date: "2026-11-06", Theme: "Amanah"}, "isi minimal satu petugas"},
		"long name": {SaveFridayDutyRequest{Date: "2026-11-06", Imam: string(make([]rune, 101))}, "nama petugas maksimal 100 karakter"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := appMessage(t, tt.req.ValidateAndNormalize()); got != tt.want {
				t.Errorf("message = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFridayDutyRange(t *testing.T) {
	r := FridayDutyRange{From: " 2026-11-01 ", To: "2026-11-30"}
	if err := r.ValidateAndNormalize(); err != nil || r.From != "2026-11-01" {
		t.Errorf("got %+v, %v", r, err)
	}

	backwards := FridayDutyRange{From: "2026-11-30", To: "2026-11-01"}
	if got := appMessage(t, backwards.ValidateAndNormalize()); got != "tanggal akhir harus setelah tanggal awal" {
		t.Errorf("backwards = %q", got)
	}
}
