package iqamahsettings

import (
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

// The repository returns AppErrors ready for the client. NOT_FOUND keeps the domain sentinel
// inside, so a usecase can errors.Is it, e.g. to answer with the default settings instead.

func errSettingsNotFound() *apperror.AppError {
	return apperror.New(apperror.CodeNotFound, domain.ErrIqamahSettingsNotFound, "pengaturan iqamah belum disimpan")
}

func errInternal(err error) *apperror.AppError {
	return apperror.New(apperror.CodeInternal, err, "gagal mengakses data iqamah")
}
