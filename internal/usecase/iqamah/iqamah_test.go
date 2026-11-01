package iqamah

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"go.uber.org/mock/gomock"
)

// repository errors, shaped like the real repository's AppErrors
var (
	errSettingsNotFound = apperror.New(apperror.CodeNotFound, domain.ErrIqamahSettingsNotFound, "pengaturan iqamah belum disimpan")
	errDatabase         = apperror.New(apperror.CodeInternal, errors.New("db down"), "gagal mengakses data iqamah")
)

func newTestUsecase(t *testing.T) (*UseCase, *mock.MockIqamahRepositoryItf) {
	t.Helper()

	repo := mock.NewMockIqamahRepositoryItf(gomock.NewController(t))
	return NewIqamahUsecase(repo), repo
}

func TestGetSettings(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().GetSettings(gomock.Any()).Return(domain.IqamahSettings{}, errSettingsNotFound)

	got, err := uc.GetSettings(context.Background())
	if err != nil || !reflect.DeepEqual(got, domain.DefaultIqamahSettings()) {
		t.Errorf("never saved: got %+v, %v; want defaults", got, err)
	}

	repo.EXPECT().GetSettings(gomock.Any()).Return(domain.IqamahSettings{}, errDatabase)
	if _, err := uc.GetSettings(context.Background()); !errors.Is(err, errDatabase) {
		t.Errorf("err = %v, want errDatabase", err)
	}
}

func TestUpdateSettings(t *testing.T) {
	uc, repo := newTestUsecase(t)
	settings := domain.DefaultIqamahSettings()
	settings.SholatMin.Jumat = 50
	repo.EXPECT().UpdateSettings(gomock.Any(), settings, int64(7)).Return(nil)

	got, err := uc.UpdateSettings(context.Background(), domain.UpdateIqamahSettingsRequest{UpdatedBy: 7, IqamahSettings: settings})
	if err != nil || !reflect.DeepEqual(got, settings) {
		t.Errorf("got %+v, %v", got, err)
	}

	// invalid: never reaches the repository
	settings.IqamahMin.Subuh = 0
	if _, err := uc.UpdateSettings(context.Background(), domain.UpdateIqamahSettingsRequest{IqamahSettings: settings}); err == nil {
		t.Error("want a validation error")
	}
}
