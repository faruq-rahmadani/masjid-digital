package domain

import (
	"strings"
	"testing"
)

var slidePNG = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")

func TestCreateSlideTitle(t *testing.T) {
	tests := []struct {
		name, title, fileName, want string
	}{
		{"given title wins", " Kajian Ahad ", "poster.png", "Kajian Ahad"},
		{"file name without extension", "", "Poster Idul Adha 1447.webp", "Poster Idul Adha 1447"},
		{"long file name is cut", "", strings.Repeat("é", 120) + ".png", strings.Repeat("é", MaxSlideTitleLength)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateSlideRequest{Title: tt.title, FileName: tt.fileName, Data: slidePNG}
			image, err := req.ValidateAndDetect()
			if err != nil {
				t.Fatalf("validate: %v", err)
			}
			if req.Title != tt.want {
				t.Errorf("title = %q, want %q", req.Title, tt.want)
			}
			if image.ContentType != "image/png" {
				t.Errorf("content type = %q, want image/png", image.ContentType)
			}
			if req.DurationSeconds != DefaultSlideDuration {
				t.Errorf("duration = %d, want the default %d", req.DurationSeconds, DefaultSlideDuration)
			}
		})
	}
}

func TestCreateSlideKeepsCaptionAndDuration(t *testing.T) {
	req := CreateSlideRequest{Title: "Infaq", Caption: " Transfer ke BSI ", DurationSeconds: 30, Data: slidePNG}
	if _, err := req.ValidateAndDetect(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if req.Caption != "Transfer ke BSI" || req.DurationSeconds != 30 {
		t.Errorf("caption, duration = %q, %d", req.Caption, req.DurationSeconds)
	}
}

func TestCreateSlideRejects(t *testing.T) {
	tests := []struct {
		name        string
		req         CreateSlideRequest
		wantMessage string
	}{
		{"no title or file name", CreateSlideRequest{FileName: ".png", Data: slidePNG}, "judul wajib diisi"},
		{"title too long", CreateSlideRequest{Title: strings.Repeat("a", 101), Data: slidePNG}, "judul maksimal 100 karakter"},
		{"caption too long", CreateSlideRequest{Title: "Kajian", Caption: strings.Repeat("a", 201), Data: slidePNG}, "keterangan maksimal 200 karakter"},
		{"duration not offered", CreateSlideRequest{Title: "Kajian", DurationSeconds: 12, Data: slidePNG}, "durasi harus 5, 10, 15, 20, atau 30 detik"},
		{"no image", CreateSlideRequest{Title: "Kajian"}, "file gambar wajib diunggah"},
		{"image too large", CreateSlideRequest{Title: "Kajian", Data: append(slidePNG, make([]byte, MaxSlideImageSize)...)}, "ukuran gambar maksimal 5 MB"},
		{"not an allowed image", CreateSlideRequest{Title: "Kajian", Data: []byte("GIF89a")}, "format gambar harus PNG, JPG, atau WebP"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.req.ValidateAndDetect()
			if got := appMessage(t, err); got != tt.wantMessage {
				t.Errorf("message = %q, want %q", got, tt.wantMessage)
			}
		})
	}
}

func TestUpdateSlideValidation(t *testing.T) {
	valid := func() UpdateSlideRequest {
		return UpdateSlideRequest{ID: 1, Title: "Kajian", Caption: "Ahad", DurationSeconds: 10, Active: true}
	}

	tests := []struct {
		name        string
		edit        func(*UpdateSlideRequest)
		wantMessage string
	}{
		{"valid", func(*UpdateSlideRequest) {}, ""},
		{"empty caption", func(r *UpdateSlideRequest) { r.Caption = "  " }, ""},
		{"blank title", func(r *UpdateSlideRequest) { r.Title = " " }, "judul wajib diisi"},
		{"caption too long", func(r *UpdateSlideRequest) { r.Caption = strings.Repeat("a", 201) }, "keterangan maksimal 200 karakter"},
		{"duration not offered", func(r *UpdateSlideRequest) { r.DurationSeconds = 12 }, "durasi harus 5, 10, 15, 20, atau 30 detik"},
		{"no duration", func(r *UpdateSlideRequest) { r.DurationSeconds = 0 }, "durasi harus 5, 10, 15, 20, atau 30 detik"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := valid()
			tt.edit(&req)
			err := req.ValidateAndNormalize()
			if tt.wantMessage == "" {
				if err != nil {
					t.Errorf("err = %v, want nil", err)
				}
				return
			}
			if got := appMessage(t, err); got != tt.wantMessage {
				t.Errorf("message = %q, want %q", got, tt.wantMessage)
			}
		})
	}
}
