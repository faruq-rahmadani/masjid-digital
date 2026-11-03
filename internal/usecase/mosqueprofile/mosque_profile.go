package mosqueprofile

import (
	"context"
	"errors"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase"
)

type UseCase struct {
	profiles    usecase.MosqueProfileRepositoryItf
	activityLog usecase.ActivityLogPublisherItf
}

func NewMosqueProfileUsecase(profiles usecase.MosqueProfileRepositoryItf, activityLog usecase.ActivityLogPublisherItf) *UseCase {
	return &UseCase{profiles: profiles, activityLog: activityLog}
}

// Get returns the mosque's profile, empty when never saved
func (u *UseCase) Get(ctx context.Context) (domain.MosqueProfile, error) {
	profile, err := u.profiles.Get(ctx)
	if errors.Is(err, domain.ErrMosqueProfileNotFound) {
		return domain.MosqueProfile{}, nil
	}
	return profile, err
}

// Update validates req and replaces the whole profile with it
func (u *UseCase) Update(ctx context.Context, req domain.UpdateMosqueProfileRequest) (domain.MosqueProfile, error) {
	if err := req.ValidateAndNormalize(); err != nil {
		return domain.MosqueProfile{}, err
	}

	profile := req.ToMosqueProfile()
	if err := u.profiles.Update(ctx, profile, req.UpdatedBy); err != nil {
		return domain.MosqueProfile{}, err
	}
	u.activityLog.Publish(ctx, req.ToActivityLog())

	return profile, nil
}

// GetLogo returns the mosque's logo, or NOT_FOUND when none was uploaded
func (u *UseCase) GetLogo(ctx context.Context) (domain.MosqueLogo, error) {
	return u.profiles.GetLogo(ctx)
}

// UpdateLogo validates the uploaded image and replaces the logo with it
func (u *UseCase) UpdateLogo(ctx context.Context, req domain.UpdateMosqueLogoRequest) error {
	logo, err := req.ValidateAndDetect()
	if err != nil {
		return err
	}

	if err := u.profiles.UpdateLogo(ctx, logo, req.UpdatedBy); err != nil {
		return err
	}
	u.activityLog.Publish(ctx, req.ToActivityLog())
	return nil
}

// DeleteLogo removes the logo; removing a missing logo is not an error
func (u *UseCase) DeleteLogo(ctx context.Context) error {
	if err := u.profiles.DeleteLogo(ctx); err != nil {
		return err
	}
	u.activityLog.Publish(ctx, domain.MosqueLogoDeletedActivityLog())
	return nil
}
