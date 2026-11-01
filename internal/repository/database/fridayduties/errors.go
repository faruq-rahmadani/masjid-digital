package fridayduties

import (
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

// The repository returns AppErrors ready for the client. NOT_FOUND keeps the domain sentinel
// inside, so a caller can errors.Is it.

func errNotFound() *apperror.AppError {
	return apperror.New(apperror.CodeNotFound, domain.ErrFridayDutyNotFound, "petugas Jumat tidak ditemukan")
}

func errInternal(err error) *apperror.AppError {
	return apperror.New(apperror.CodeInternal, err, "gagal mengakses data petugas Jumat")
}
