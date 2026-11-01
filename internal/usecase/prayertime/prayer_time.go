package prayertime

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

const ConfigDays = 30

const dateLayout = "2006-01-02"

type UseCase struct {
	prayerTimes usecase.PrayerTimeRepositoryItf
	now         func() time.Time
}

func NewPrayerTimeUsecase(prayerTimes usecase.PrayerTimeRepositoryItf) *UseCase {
	return &UseCase{prayerTimes: prayerTimes, now: time.Now}
}

func (u *UseCase) GetSettings(ctx context.Context) (domain.PrayerSettings, error) {
	settings, err := u.prayerTimes.GetSettings(ctx)
	if errors.Is(err, domain.ErrPrayerSettingsNotFound) {
		return domain.DefaultPrayerSettings(), nil
	}
	return settings, err
}

func (u *UseCase) UpdateSettings(ctx context.Context, req domain.UpdatePrayerSettingsRequest) (domain.PrayerSettings, error) {
	if err := req.ValidateAndNormalize(); err != nil {
		return domain.PrayerSettings{}, err
	}

	if err := u.prayerTimes.UpdateSettings(ctx, req.PrayerSettings, req.UpdatedBy); err != nil {
		return domain.PrayerSettings{}, err
	}
	return req.PrayerSettings, nil
}

func (u *UseCase) ListOverrides(ctx context.Context, r domain.PrayerTimeOverrideRange) ([]domain.PrayerTimeOverride, error) {
	if err := r.ValidateAndNormalize(); err != nil {
		return nil, err
	}
	return u.prayerTimes.FindOverridesBetween(ctx, r.From, r.To)
}

func (u *UseCase) CreateOverride(ctx context.Context, req domain.CreatePrayerTimeOverrideRequest) (domain.PrayerTimeOverride, error) {
	if err := req.ValidateAndNormalize(); err != nil {
		return domain.PrayerTimeOverride{}, err
	}
	if earliest := u.earliestToday(); req.Date < earliest {
		return domain.PrayerTimeOverride{}, apperror.New(apperror.CodeInvalidInput, fmt.Errorf("date %s is before %s", req.Date, earliest), "tanggal sudah lewat")
	}

	return u.prayerTimes.CreateOverride(ctx, req.ToPrayerTimeOverride())
}

func (u *UseCase) DeleteOverride(ctx context.Context, id int64) error {
	return u.prayerTimes.DeleteOverride(ctx, id)
}

func (u *UseCase) GetConfig(ctx context.Context) (domain.PrayerConfig, error) {
	settings, err := u.GetSettings(ctx)
	if err != nil {
		return domain.PrayerConfig{}, err
	}

	from := u.earliestToday()
	to := u.now().UTC().AddDate(0, 0, ConfigDays).Format(dateLayout)
	overrides, err := u.prayerTimes.FindOverridesBetween(ctx, from, to)
	if err != nil {
		return domain.PrayerConfig{}, err
	}

	return domain.PrayerConfig{Settings: settings, Overrides: overrides}, nil
}

func (u *UseCase) earliestToday() string {
	return u.now().UTC().Format(dateLayout)
}
