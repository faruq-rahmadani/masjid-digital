package slide

import (
	"context"
	"fmt"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

type useCase struct {
	slides      usecase.SlideRepositoryItf
	activityLog usecase.ActivityLogPublisherItf
}

func New(slides usecase.SlideRepositoryItf, activityLog usecase.ActivityLogPublisherItf) *useCase {
	return &useCase{slides: slides, activityLog: activityLog}
}

func (u *useCase) List(ctx context.Context) ([]domain.Slide, error) {
	return u.slides.FindAll(ctx)
}

func (u *useCase) Create(ctx context.Context, req domain.CreateSlideRequest) (domain.Slide, error) {
	image, err := req.ValidateAndDetect()
	if err != nil {
		return domain.Slide{}, err
	}
	slide, err := u.slides.Create(ctx, domain.Slide{
		Title:           req.Title,
		Caption:         req.Caption,
		DurationSeconds: req.DurationSeconds,
		ShowCaption:     req.ShowCaption,
		CreatedBy:       req.CreatedBy,
	}, image)
	if err != nil {
		return domain.Slide{}, err
	}
	u.activityLog.Publish(ctx, req.ToActivityLog(slide.ID))
	return slide, nil
}

func (u *useCase) Update(ctx context.Context, req domain.UpdateSlideRequest) (domain.Slide, error) {
	if err := req.ValidateAndNormalize(); err != nil {
		return domain.Slide{}, err
	}
	slide, err := u.slides.Update(ctx, domain.Slide{
		ID:              req.ID,
		Title:           req.Title,
		Caption:         req.Caption,
		DurationSeconds: req.DurationSeconds,
		ShowCaption:     req.ShowCaption,
		Active:          req.Active,
	})
	if err != nil {
		return domain.Slide{}, err
	}
	u.activityLog.Publish(ctx, req.ToActivityLog())
	return slide, nil
}

func (u *useCase) Delete(ctx context.Context, id int64) error {
	if err := u.slides.Delete(ctx, id); err != nil {
		return err
	}
	u.activityLog.Publish(ctx, domain.SlideDeletedActivityLog(id))
	return nil
}

func (u *useCase) GetImage(ctx context.Context, id int64) (domain.SlideImage, error) {
	return u.slides.GetImage(ctx, id)
}

// Reorder saves a new display order. ids must hold every slide exactly once, so an admin working
// on a stale list (one added or deleted meanwhile) reloads instead of losing a slide.
func (u *useCase) Reorder(ctx context.Context, req domain.ReorderSlidesRequest) ([]domain.Slide, error) {
	current, err := u.slides.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	remaining := make(map[int64]bool, len(current))
	for _, s := range current {
		remaining[s.ID] = true
	}
	for _, id := range req.IDs {
		if !remaining[id] {
			return nil, errStaleOrder(fmt.Errorf("id %d unknown or repeated", id))
		}
		delete(remaining, id)
	}
	if len(remaining) > 0 {
		return nil, errStaleOrder(fmt.Errorf("%d slides missing from the order", len(remaining)))
	}

	if err := u.slides.Reorder(ctx, req.IDs); err != nil {
		return nil, err
	}
	u.activityLog.Publish(ctx, req.ToActivityLog())
	return u.slides.FindAll(ctx)
}

// ListDisplay returns what the TV shows: the active slides in order
func (u *useCase) ListDisplay(ctx context.Context) ([]domain.DisplaySlide, error) {
	all, err := u.slides.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	slides := make([]domain.DisplaySlide, 0, len(all))
	for _, s := range all {
		if s.IsActive() {
			slides = append(slides, s.ToDisplaySlide())
		}
	}
	return slides, nil
}

func errStaleOrder(err error) *apperror.AppError {
	return apperror.New(apperror.CodeInvalidInput, err, "daftar slide sudah berubah, muat ulang halaman lalu urutkan lagi")
}
