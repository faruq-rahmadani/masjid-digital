package runningtext

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
	errSettingNotFound = apperror.New(apperror.CodeNotFound, domain.ErrRunningTextSettingNotFound, "pengaturan running text belum disimpan")
	errDatabase        = apperror.New(apperror.CodeInternal, errors.New("db down"), "gagal mengakses data running text")
)

var saved = []domain.RunningText{
	{ID: 1, Text: "Kajian Tafsir", Active: true, Position: 0},
	{ID: 2, Text: "Pendaftaran qurban", Active: false, Position: 1},
	{ID: 3, Text: "Infaq via QRIS", Active: true, Position: 2},
}

func newTestUsecase(t *testing.T) (*useCase, *mock.MockRunningTextRepositoryItf) {
	t.Helper()

	repo := mock.NewMockRunningTextRepositoryItf(gomock.NewController(t))
	return New(repo), repo
}

func TestCreate(t *testing.T) {
	uc, repo := newTestUsecase(t)
	want := domain.RunningText{ID: 4, Text: "Kerja bakti", Active: true, Position: 3, CreatedBy: 7}
	repo.EXPECT().Create(gomock.Any(), domain.RunningText{Text: "Kerja bakti", CreatedBy: 7}).Return(want, nil)

	got, err := uc.Create(context.Background(), domain.CreateRunningTextRequest{CreatedBy: 7, Text: " Kerja bakti "})
	if err != nil || got != want {
		t.Errorf("got %+v, %v; want %+v", got, err, want)
	}
}

func TestCreateInvalidNeverSaves(t *testing.T) {
	uc, _ := newTestUsecase(t)

	if _, err := uc.Create(context.Background(), domain.CreateRunningTextRequest{Text: " "}); err == nil {
		t.Error("want a validation error")
	}
}

func TestUpdate(t *testing.T) {
	uc, repo := newTestUsecase(t)
	want := domain.RunningText{ID: 2, Text: "Pendaftaran qurban", Active: true, Position: 1}
	repo.EXPECT().Update(gomock.Any(), domain.RunningText{ID: 2, Text: "Pendaftaran qurban", Active: true}).Return(want, nil)

	got, err := uc.Update(context.Background(), domain.UpdateRunningTextRequest{ID: 2, Text: "Pendaftaran qurban", Active: true})
	if err != nil || got != want {
		t.Errorf("got %+v, %v; want %+v", got, err, want)
	}
}

func TestReorder(t *testing.T) {
	uc, repo := newTestUsecase(t)
	reordered := []domain.RunningText{saved[2], saved[0], saved[1]}
	gomock.InOrder(
		repo.EXPECT().FindAll(gomock.Any()).Return(saved, nil),
		repo.EXPECT().Reorder(gomock.Any(), []int64{3, 1, 2}).Return(nil),
		repo.EXPECT().FindAll(gomock.Any()).Return(reordered, nil),
	)

	got, err := uc.Reorder(context.Background(), domain.ReorderRunningTextsRequest{IDs: []int64{3, 1, 2}})
	if err != nil || !reflect.DeepEqual(got, reordered) {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestReorderStaleList(t *testing.T) {
	tests := []struct {
		name string
		ids  []int64
	}{
		{"missing one", []int64{3, 1}},
		{"unknown id", []int64{3, 1, 2, 9}},
		{"repeated id", []int64{3, 1, 1}},
		{"empty", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, repo := newTestUsecase(t)
			repo.EXPECT().FindAll(gomock.Any()).Return(saved, nil)

			_, err := uc.Reorder(context.Background(), domain.ReorderRunningTextsRequest{IDs: tt.ids})
			var appErr *apperror.AppError
			if !errors.As(err, &appErr) || appErr.Code != apperror.CodeInvalidInput {
				t.Errorf("err = %v, want INVALID_INPUT", err)
			}
		})
	}
}

func TestGetSettingDefaultsWhenNeverSaved(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().GetSetting(gomock.Any()).Return(domain.RunningTextSetting{}, errSettingNotFound)

	got, err := uc.GetSetting(context.Background())
	if err != nil || got != domain.DefaultRunningTextSetting() {
		t.Errorf("got %+v, %v; want default", got, err)
	}
}

func TestGetSettingPassesOtherErrors(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().GetSetting(gomock.Any()).Return(domain.RunningTextSetting{}, errDatabase)

	if _, err := uc.GetSetting(context.Background()); !errors.Is(err, errDatabase) {
		t.Errorf("err = %v, want errDatabase", err)
	}
}

func TestUpdateSetting(t *testing.T) {
	uc, repo := newTestUsecase(t)
	fast := domain.RunningTextSetting{Speed: domain.ScrollSpeedFast}
	repo.EXPECT().UpdateSetting(gomock.Any(), fast, int64(7)).Return(nil)

	got, err := uc.UpdateSetting(context.Background(), domain.UpdateRunningTextSettingRequest{UpdatedBy: 7, RunningTextSetting: fast})
	if err != nil || got != fast {
		t.Errorf("got %+v, %v", got, err)
	}

	if _, err := uc.UpdateSetting(context.Background(), domain.UpdateRunningTextSettingRequest{RunningTextSetting: domain.RunningTextSetting{Speed: "turbo"}}); err == nil {
		t.Error("turbo: want a validation error")
	}
}

func TestGetDisplayKeepsActiveInOrder(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().GetSetting(gomock.Any()).Return(domain.RunningTextSetting{}, errSettingNotFound)
	repo.EXPECT().FindAll(gomock.Any()).Return(saved, nil)

	got, err := uc.GetDisplay(context.Background())
	want := domain.RunningTextDisplay{Speed: domain.ScrollSpeedMedium, Texts: []string{"Kajian Tafsir", "Infaq via QRIS"}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, %v; want %+v", got, err, want)
	}
}
