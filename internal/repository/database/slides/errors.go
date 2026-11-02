package slides

import (
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

func errNotFound() *apperror.AppError {
	return apperror.New(apperror.CodeNotFound, domain.ErrSlideNotFound, "slide tidak ditemukan")
}

func errInternal(err error) *apperror.AppError {
	return apperror.New(apperror.CodeInternal, err, "gagal mengakses data slide")
}
