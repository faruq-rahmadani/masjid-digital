package prayertimes

import (
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

func errSettingsNotFound() *apperror.AppError {
	return apperror.New(apperror.CodeNotFound, domain.ErrPrayerSettingsNotFound, "pengaturan waktu sholat belum disimpan")
}

func errOverrideNotFound() *apperror.AppError {
	return apperror.New(apperror.CodeNotFound, domain.ErrPrayerTimeOverrideNotFound, "jadwal khusus tidak ditemukan")
}

func errOverrideExists() *apperror.AppError {
	return apperror.New(apperror.CodeAlreadyExists, domain.ErrPrayerTimeOverrideExists, "waktu sholat ini sudah punya jadwal khusus di tanggal tersebut")
}

func errInternal(err error) *apperror.AppError {
	return apperror.New(apperror.CodeInternal, err, "gagal mengakses data waktu sholat")
}
