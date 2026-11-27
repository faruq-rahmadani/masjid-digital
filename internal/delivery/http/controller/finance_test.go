package controller

import (
	"net/http"
	"strings"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"go.uber.org/mock/gomock"
)

func newFinanceController(t *testing.T) (*Controller, *mock.MockFinanceUsecaseItf) {
	t.Helper()
	finance := mock.NewMockFinanceUsecaseItf(gomock.NewController(t))
	return NewController(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, finance), finance
}

func TestCreateAccountTakesTheCreatorFromTheSession(t *testing.T) {
	ctrl, finance := newFinanceController(t)
	finance.EXPECT().CreateAccount(gomock.Any(), domain.CreateAccountRequest{CreatedBy: loggedInAs.UserID, Name: "BSI Masjid", Kind: domain.AccountKindBank, Number: "7123"}).
		Return(domain.Account{ID: 2, Name: "BSI Masjid"}, nil)

	rec, err := serve(authenticated(t, ctrl.CreateAccount), http.MethodPost, "/api/v1/finance/accounts", `{"name":"BSI Masjid","kind":"bank","number":"7123"}`, "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if rec.Code != http.StatusCreated || !strings.Contains(rec.Body.String(), `"name":"BSI Masjid"`) {
		t.Errorf("status %d, body %s", rec.Code, rec.Body.String())
	}
}

func TestListPocketsFlattensTheBalance(t *testing.T) {
	ctrl, finance := newFinanceController(t)
	finance.EXPECT().ListPockets(gomock.Any()).Return([]domain.PocketWithBalance{{Pocket: domain.Pocket{ID: 1, Name: "Kas Umum"}, Balance: 2_250_000}}, nil)

	rec, err := serve(authenticated(t, ctrl.ListPockets), http.MethodGet, "/api/v1/finance/pockets", "", "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if body := rec.Body.String(); !strings.Contains(body, `"name":"Kas Umum"`) || !strings.Contains(body, `"balance":2250000`) {
		t.Errorf("body = %s", body)
	}
}

func TestArchivePocketReadsTheID(t *testing.T) {
	ctrl, finance := newFinanceController(t)
	finance.EXPECT().ArchivePocket(gomock.Any(), int64(5)).Return(domain.Pocket{ID: 5, Status: domain.PocketStatusArchived}, nil)

	rec, err := serve(authenticated(t, ctrl.ArchivePocket), http.MethodPost, "/api/v1/finance/pockets/5/archive", "", "Bearer valid-token", "id", "5")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if !strings.Contains(rec.Body.String(), `"status":"archived"`) {
		t.Errorf("body = %s", rec.Body.String())
	}

	if _, err := serve(authenticated(t, ctrl.ArchivePocket), http.MethodPost, "/api/v1/finance/pockets/x/archive", "", "Bearer valid-token", "id", "x"); err == nil {
		t.Error("id x: want an error")
	}
}

func TestStartBookkeepingTakesTheUserFromTheSession(t *testing.T) {
	ctrl, finance := newFinanceController(t)
	finance.EXPECT().StartBookkeeping(gomock.Any(), domain.StartBookkeepingRequest{
		ClosedBy: loggedInAs.UserID, StartMonth: "2026-11",
		Accounts: []domain.OpeningBalance{{ID: 1, Amount: 500_000}},
		Pockets:  []domain.OpeningBalance{{ID: 1, Amount: 500_000}},
	}).Return(domain.Bookkeeping{Started: true, StartMonth: "2026-11"}, nil)

	body := `{"start_month":"2026-11","accounts":[{"id":1,"amount":500000}],"pockets":[{"id":1,"amount":500000}]}`
	rec, err := serve(authenticated(t, ctrl.StartBookkeeping), http.MethodPost, "/api/v1/finance/bookkeeping", body, "Bearer valid-token")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if rec.Code != http.StatusCreated || !strings.Contains(rec.Body.String(), `"started":true`) {
		t.Errorf("status %d, body %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteCategory(t *testing.T) {
	ctrl, finance := newFinanceController(t)
	finance.EXPECT().DeleteCategory(gomock.Any(), int64(3)).Return(nil)

	rec, err := serve(authenticated(t, ctrl.DeleteCategory), http.MethodDelete, "/api/v1/finance/categories/3", "", "Bearer valid-token", "id", "3")
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"message":"kategori dihapus"}` {
		t.Errorf("body = %s", got)
	}
}
