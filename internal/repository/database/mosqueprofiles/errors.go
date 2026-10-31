package mosqueprofiles

import (
	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

// The repository returns AppErrors ready for the client. NOT_FOUND keeps the
// domain sentinel inside, so a usecase can errors.Is it, e.g. to answer with an
// empty profile instead.

func errProfileNotFound() *apperror.AppError {
	return apperror.New(apperror.CodeNotFound, domain.ErrMosqueProfileNotFound, "profil masjid belum diisi")
}

func errLogoNotFound() *apperror.AppError {
	return apperror.New(apperror.CodeNotFound, domain.ErrMosqueLogoNotFound, "logo masjid belum diunggah")
}

func errInternal(err error) *apperror.AppError {
	return apperror.New(apperror.CodeInternal, err, "gagal mengakses data profil masjid")
}
