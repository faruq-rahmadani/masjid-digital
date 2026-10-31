package users

import (
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

// The repository returns AppErrors ready for the client. Each keeps the domain
// sentinel as its internal error, so a usecase can still errors.Is it and
// answer differently, e.g. login hides that a phone number is not registered.

func errUserNotFound() *apperror.AppError {
	return apperror.New(apperror.CodeNotFound, domain.ErrUserNotFound, "user tidak ditemukan")
}

func errUserAlreadyExists() *apperror.AppError {
	return apperror.New(apperror.CodeAlreadyExists, domain.ErrUserAlreadyExists, "nomor HP atau email sudah terdaftar")
}

func errInternal(err error) *apperror.AppError {
	return apperror.New(apperror.CodeInternal, err, "gagal mengakses data user")
}
