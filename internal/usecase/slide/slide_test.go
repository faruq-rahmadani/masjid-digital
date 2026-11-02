package slide

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

// a repository error, shaped like the real repository's AppErrors
var errDatabase = apperror.New(apperror.CodeInternal, errors.New("db down"), "gagal mengakses data slide")

// pngData is the start of a PNG file, enough for content type detection
var pngData = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")

var saved = []domain.Slide{
	{ID: 1, Title: "Kajian", Caption: "Ahad pagi", DurationSeconds: 10, ShowCaption: true, Active: true, Position: 0},
	{ID: 2, Title: "Idul Adha", DurationSeconds: 15, Active: false, Position: 1},
	{ID: 3, Title: "Infaq", DurationSeconds: 30, Active: true, Position: 2},
}

func newTestUsecase(t *testing.T) (*useCase, *mock.MockSlideRepositoryItf) {
	t.Helper()

	repo := mock.NewMockSlideRepositoryItf(gomock.NewController(t))
	return New(repo), repo
}

func TestCreate(t *testing.T) {
	uc, repo := newTestUsecase(t)
	want := domain.Slide{ID: 4, Title: "Poster Kajian", DurationSeconds: 10, Active: true, CreatedBy: 7}
	repo.EXPECT().Create(gomock.Any(),
		domain.Slide{Title: "Poster Kajian", DurationSeconds: domain.DefaultSlideDuration, CreatedBy: 7},
		domain.SlideImage{ContentType: "image/png", Data: pngData},
	).Return(want, nil)

	// no title: the file name stands in
	got, err := uc.Create(context.Background(), domain.CreateSlideRequest{CreatedBy: 7, FileName: "Poster Kajian.png", Data: pngData})
	if err != nil || got != want {
		t.Errorf("create = %+v, %v; want %+v", got, err, want)
	}
}

func TestCreateWithCaptionAndDuration(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().Create(gomock.Any(),
		domain.Slide{Title: "Infaq", Caption: "Transfer ke BSI", DurationSeconds: 20},
		domain.SlideImage{ContentType: "image/png", Data: pngData},
	).Return(domain.Slide{ID: 5}, nil)

	if _, err := uc.Create(context.Background(), domain.CreateSlideRequest{Title: "Infaq", Caption: "Transfer ke BSI ", DurationSeconds: 20, Data: pngData}); err != nil {
		t.Errorf("create: %v", err)
	}
}

func TestCreateRejectsBadImage(t *testing.T) {
	uc, _ := newTestUsecase(t)

	_, err := uc.Create(context.Background(), domain.CreateSlideRequest{Title: "Kajian", Data: []byte("<svg></svg>")})
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeInvalidInput {
		t.Errorf("err = %v, want INVALID_INPUT", err)
	}
}

func TestUpdate(t *testing.T) {
	uc, repo := newTestUsecase(t)
	want := domain.Slide{ID: 2, Title: "Infaq Atap", Caption: "Transfer ke BSI", DurationSeconds: 15, ShowCaption: true, Active: true}
	repo.EXPECT().Update(gomock.Any(), want).Return(want, nil)

	got, err := uc.Update(context.Background(), domain.UpdateSlideRequest{ID: 2, Title: " Infaq Atap ", Caption: " Transfer ke BSI ", DurationSeconds: 15, ShowCaption: true, Active: true})
	if err != nil || got != want {
		t.Errorf("update = %+v, %v; want %+v", got, err, want)
	}
}

func TestReorder(t *testing.T) {
	tests := []struct {
		name    string
		ids     []int64
		wantErr bool
	}{
		{"every slide once", []int64{3, 1, 2}, false},
		{"one missing", []int64{3, 1}, true},
		{"repeated", []int64{3, 1, 1}, true},
		{"unknown", []int64{3, 1, 2, 9}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, repo := newTestUsecase(t)
			repo.EXPECT().FindAll(gomock.Any()).Return(saved, nil)
			if !tt.wantErr {
				repo.EXPECT().Reorder(gomock.Any(), tt.ids).Return(nil)
				repo.EXPECT().FindAll(gomock.Any()).Return(saved, nil)
			}

			_, err := uc.Reorder(context.Background(), domain.ReorderSlidesRequest{IDs: tt.ids})
			var appErr *apperror.AppError
			if tt.wantErr && (!errors.As(err, &appErr) || appErr.Code != apperror.CodeInvalidInput) {
				t.Errorf("err = %v, want INVALID_INPUT", err)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("err = %v", err)
			}
		})
	}
}

func TestListDisplay(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().FindAll(gomock.Any()).Return(saved, nil)

	got, err := uc.ListDisplay(context.Background())
	want := []domain.DisplaySlide{
		{ID: 1, Title: "Kajian", Caption: "Ahad pagi", DurationSeconds: 10, ShowCaption: true},
		{ID: 3, Title: "Infaq", DurationSeconds: 30},
	}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("display = %+v, %v; want %+v", got, err, want)
	}
}

func TestListDisplayRepositoryError(t *testing.T) {
	uc, repo := newTestUsecase(t)
	repo.EXPECT().FindAll(gomock.Any()).Return(nil, errDatabase)

	if _, err := uc.ListDisplay(context.Background()); !errors.Is(err, errDatabase) {
		t.Errorf("err = %v, want the repository's", err)
	}
}
