package runningtexts

import (
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

func errNotFound() *apperror.AppError {
	return apperror.New(apperror.CodeNotFound, domain.ErrRunningTextNotFound, "running text tidak ditemukan")
}

func errSettingNotFound() *apperror.AppError {
	return apperror.New(apperror.CodeNotFound, domain.ErrRunningTextSettingNotFound, "pengaturan running text belum disimpan")
}

func errInternal(err error) *apperror.AppError {
	return apperror.New(apperror.CodeInternal, err, "gagal mengakses data running text")
}
