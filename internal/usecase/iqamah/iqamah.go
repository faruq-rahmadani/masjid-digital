package iqamah

import (
	"context"
	"errors"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase"
)

type UseCase struct {
	iqamah usecase.IqamahRepositoryItf
}

func NewIqamahUsecase(iqamah usecase.IqamahRepositoryItf) *UseCase {
	return &UseCase{iqamah: iqamah}
}

// GetSettings returns the saved settings, or the defaults when never saved
func (u *UseCase) GetSettings(ctx context.Context) (domain.IqamahSettings, error) {
	settings, err := u.iqamah.GetSettings(ctx)
	if errors.Is(err, domain.ErrIqamahSettingsNotFound) {
		return domain.DefaultIqamahSettings(), nil
	}
	return settings, err
}

// UpdateSettings validates req and replaces every setting with it
func (u *UseCase) UpdateSettings(ctx context.Context, req domain.UpdateIqamahSettingsRequest) (domain.IqamahSettings, error) {
	if err := req.Validate(); err != nil {
		return domain.IqamahSettings{}, err
	}

	if err := u.iqamah.UpdateSettings(ctx, req.IqamahSettings, req.UpdatedBy); err != nil {
		return domain.IqamahSettings{}, err
	}
	return req.IqamahSettings, nil
}
