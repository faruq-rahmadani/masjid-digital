package controller

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"go.uber.org/mock/gomock"
)

func newRunningTextController(t *testing.T) (*Controller, *mock.MockRunningTextUsecaseItf) {
	t.Helper()
	runningTexts := mock.NewMockRunningTextUsecaseItf(gomock.NewController(t))
	return NewController(nil, nil, nil, nil, runningTexts), runningTexts
}

func TestCreateRunningTextTakesCreatorFromSession(t *testing.T) {
	created := time.Date(2026, 11, 1, 8, 0, 0, 0, time.UTC)
	ctrl, runningTexts := newRunningTextController(t)
	runningTexts.EXPECT().Create(gomock.Any(), domain.CreateRunningTextRequest{CreatedBy: loggedInAs.UserID, Text: "Kerja bakti"}).
		Return(domain.RunningText{ID: 4, Text: "Kerja bakti", Active: true, Position: 3, CreatedBy: loggedInAs.UserID, CreatedAt: created, UpdatedAt: created}, nil)

	rec, err := serve(authenticated(t, ctrl.CreateRunningText), http.MethodPost, "/api/v1/running-texts", `{"text":"Kerja bakti","created_by":99}`, "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rec.Code)
	}
	if got := rec.Body.String(); !strings.Contains(got, `"id":4`) || !strings.Contains(got, `"active":true`) || !strings.Contains(got, `"position":3`) {
		t.Errorf("body = %s", got)
	}
}

func TestUpdateRunningTextTakesIDFromPath(t *testing.T) {
	ctrl, runningTexts := newRunningTextController(t)
	runningTexts.EXPECT().Update(gomock.Any(), domain.UpdateRunningTextRequest{ID: 2, Text: "Kajian", Active: false}).
		Return(domain.RunningText{ID: 2, Text: "Kajian"}, nil)

	rec, err := serve(authenticated(t, ctrl.UpdateRunningText), http.MethodPut, "/api/v1/running-texts/2", `{"id":9,"text":"Kajian","active":false}`, "Bearer valid-token", "id", "2")
	if err != nil || rec.Code != http.StatusOK {
		t.Fatalf("handler: %d, %v", rec.Code, err)
	}
}

func TestDeleteRunningText(t *testing.T) {
	ctrl, runningTexts := newRunningTextController(t)
	runningTexts.EXPECT().Delete(gomock.Any(), int64(4)).Return(nil)

	rec, err := serve(authenticated(t, ctrl.DeleteRunningText), http.MethodDelete, "/api/v1/running-texts/4", "", "Bearer valid-token", "id", "4")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"message":"running text dihapus"}` {
		t.Errorf("body = %s", got)
	}
}

func TestDeleteRunningTextBadID(t *testing.T) {
	ctrl, _ := newRunningTextController(t)

	_, err := serve(authenticated(t, ctrl.DeleteRunningText), http.MethodDelete, "/api/v1/running-texts/x", "", "Bearer valid-token", "id", "x")
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeInvalidInput {
		t.Errorf("err = %v, want INVALID_INPUT", err)
	}
}

func TestReorderRunningTexts(t *testing.T) {
	ctrl, runningTexts := newRunningTextController(t)
	runningTexts.EXPECT().Reorder(gomock.Any(), domain.ReorderRunningTextsRequest{IDs: []int64{3, 1, 2}}).Return([]domain.RunningText{}, nil)

	rec, err := serve(authenticated(t, ctrl.ReorderRunningTexts), http.MethodPut, "/api/v1/running-texts/order", `{"ids":[3,1,2]}`, "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"data":[]}` {
		t.Errorf("body = %s", got)
	}
}

func TestUpdateRunningTextSettingTakesUpdaterFromSession(t *testing.T) {
	ctrl, runningTexts := newRunningTextController(t)
	fast := domain.RunningTextSetting{Speed: domain.ScrollSpeedFast}
	runningTexts.EXPECT().UpdateSetting(gomock.Any(), domain.UpdateRunningTextSettingRequest{UpdatedBy: loggedInAs.UserID, RunningTextSetting: fast}).Return(fast, nil)

	rec, err := serve(authenticated(t, ctrl.UpdateRunningTextSetting), http.MethodPut, "/api/v1/running-text/setting", `{"speed":"fast"}`, "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"data":{"speed":"fast"}}` {
		t.Errorf("body = %s", got)
	}
}

func TestGetRunningTextDisplay(t *testing.T) {
	ctrl, runningTexts := newRunningTextController(t)
	runningTexts.EXPECT().GetDisplay(gomock.Any()).Return(domain.RunningTextDisplay{Speed: domain.ScrollSpeedMedium, Texts: []string{"Kajian Tafsir"}}, nil)

	rec, err := serve(ctrl.GetRunningTextDisplay, http.MethodGet, "/api/public/v1/running-texts", "", "")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"data":{"speed":"medium","texts":["Kajian Tafsir"]}}` {
		t.Errorf("body = %s", got)
	}
}
