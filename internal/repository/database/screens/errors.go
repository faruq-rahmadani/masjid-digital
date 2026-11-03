package screens

import (
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

func errNotFound() *apperror.AppError {
	return apperror.New(apperror.CodeNotFound, domain.ErrScreenNotFound, "layar tidak ditemukan")
}

func errInternal(err error) *apperror.AppError {
	return apperror.New(apperror.CodeInternal, err, "gagal mengakses data layar")
}
