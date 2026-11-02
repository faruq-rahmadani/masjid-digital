package runningtext

import (
	"context"
	"errors"
	"fmt"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/internal/usecase"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

type useCase struct {
	runningTexts usecase.RunningTextRepositoryItf
}

func New(runningTexts usecase.RunningTextRepositoryItf) *useCase {
	return &useCase{runningTexts: runningTexts}
}

func (u *useCase) List(ctx context.Context) ([]domain.RunningText, error) {
	return u.runningTexts.FindAll(ctx)
}

func (u *useCase) Create(ctx context.Context, req domain.CreateRunningTextRequest) (domain.RunningText, error) {
	if err := req.ValidateAndNormalize(); err != nil {
		return domain.RunningText{}, err
	}
	return u.runningTexts.Create(ctx, domain.RunningText{Text: req.Text, CreatedBy: req.CreatedBy})
}

func (u *useCase) Update(ctx context.Context, req domain.UpdateRunningTextRequest) (domain.RunningText, error) {
	if err := req.ValidateAndNormalize(); err != nil {
		return domain.RunningText{}, err
	}
	return u.runningTexts.Update(ctx, domain.RunningText{ID: req.ID, Text: req.Text, Active: req.Active})
}

func (u *useCase) Delete(ctx context.Context, id int64) error {
	return u.runningTexts.Delete(ctx, id)
}

// Reorder saves a new display order. ids must hold every running text exactly once, so an admin
// working on a stale list (one added or deleted meanwhile) reloads instead of losing a text.
func (u *useCase) Reorder(ctx context.Context, req domain.ReorderRunningTextsRequest) ([]domain.RunningText, error) {
	current, err := u.runningTexts.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	remaining := make(map[int64]bool, len(current))
	for _, rt := range current {
		remaining[rt.ID] = true
	}
	for _, id := range req.IDs {
		if !remaining[id] {
			return nil, errStaleOrder(fmt.Errorf("id %d unknown or repeated", id))
		}
		delete(remaining, id)
	}
	if len(remaining) > 0 {
		return nil, errStaleOrder(fmt.Errorf("%d running texts missing from the order", len(remaining)))
	}

	if err := u.runningTexts.Reorder(ctx, req.IDs); err != nil {
		return nil, err
	}
	return u.runningTexts.FindAll(ctx)
}

func (u *useCase) GetSetting(ctx context.Context) (domain.RunningTextSetting, error) {
	setting, err := u.runningTexts.GetSetting(ctx)
	if errors.Is(err, domain.ErrRunningTextSettingNotFound) {
		return domain.DefaultRunningTextSetting(), nil
	}
	return setting, err
}

func (u *useCase) UpdateSetting(ctx context.Context, req domain.UpdateRunningTextSettingRequest) (domain.RunningTextSetting, error) {
	if err := req.ValidateAndNormalize(); err != nil {
		return domain.RunningTextSetting{}, err
	}

	if err := u.runningTexts.UpdateSetting(ctx, req.RunningTextSetting, req.UpdatedBy); err != nil {
		return domain.RunningTextSetting{}, err
	}
	return req.RunningTextSetting, nil
}

// GetDisplay returns what the TV shows: the active texts in order and the scroll speed
func (u *useCase) GetDisplay(ctx context.Context) (domain.RunningTextDisplay, error) {
	setting, err := u.GetSetting(ctx)
	if err != nil {
		return domain.RunningTextDisplay{}, err
	}

	all, err := u.runningTexts.FindAll(ctx)
	if err != nil {
		return domain.RunningTextDisplay{}, err
	}

	texts := make([]string, 0, len(all))
	for _, rt := range all {
		if rt.Active {
			texts = append(texts, rt.Text)
		}
	}
	return domain.RunningTextDisplay{Speed: setting.Speed, Texts: texts}, nil
}

func errStaleOrder(err error) *apperror.AppError {
	return apperror.New(apperror.CodeInvalidInput, err, "daftar running text sudah berubah, muat ulang halaman lalu urutkan lagi")
}
