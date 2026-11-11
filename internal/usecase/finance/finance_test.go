package finance

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/mock"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"go.uber.org/mock/gomock"
)

type mocks struct {
	tx           *mock.MockTransactorItf
	accounts     *mock.MockAccountRepositoryItf
	pockets      *mock.MockPocketRepositoryItf
	categories   *mock.MockCategoryRepositoryItf
	transactions *mock.MockTransactionRepositoryItf
	closings     *mock.MockBookClosingRepositoryItf
	profiles     *mock.MockMosqueProfileRepositoryItf
	activityLog  *mock.MockActivityLogPublisherItf
}

// 2026-11-11 08:00 WIB
var now = time.Date(2026, 11, 11, 1, 0, 0, 0, time.UTC)

func newTestUsecase(t *testing.T) (*UseCase, mocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	m := mocks{
		tx:           mock.NewMockTransactorItf(ctrl),
		accounts:     mock.NewMockAccountRepositoryItf(ctrl),
		pockets:      mock.NewMockPocketRepositoryItf(ctrl),
		categories:   mock.NewMockCategoryRepositoryItf(ctrl),
		transactions: mock.NewMockTransactionRepositoryItf(ctrl),
		closings:     mock.NewMockBookClosingRepositoryItf(ctrl),
		profiles:     mock.NewMockMosqueProfileRepositoryItf(ctrl),
		activityLog:  mock.NewMockActivityLogPublisherItf(ctrl),
	}
	m.tx.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	}).AnyTimes()
	uc := New(Repositories{
		Transactor: m.tx, Accounts: m.accounts, Pockets: m.pockets, Categories: m.categories,
		Transactions: m.transactions, BookClosings: m.closings, MosqueProfiles: m.profiles,
	}, m.activityLog)
	uc.now = func() time.Time { return now }
	return uc, m
}

func ruleError(t *testing.T, err error, code apperror.Code) {
	t.Helper()
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != code {
		t.Fatalf("err = %v, want %s", err, code)
	}
}

// opening at the end of October: Tunai 1.000.000 in Kas Umum, BSI 600.000 in Zakat Mal
var opening = domain.BookClosing{
	ID: 1, Month: "2026-10", Kind: domain.BookClosingOpening, ClosingBalance: 1_600_000,
	Accounts: []domain.BookClosingAccount{{AccountID: 1, ClosingBalance: 1_000_000}, {AccountID: 2, ClosingBalance: 600_000}},
	Pockets:  []domain.BookClosingPocket{{PocketID: 1, ClosingBalance: 1_000_000}, {PocketID: 2, ClosingBalance: 600_000}},
}

func expectBalances(m mocks, txs ...domain.Transaction) {
	m.closings.EXPECT().FindLatest(gomock.Any()).Return(opening, nil)
	m.transactions.EXPECT().Find(gomock.Any(), domain.TransactionFilter{From: "2026-11-01"}).Return(txs, nil)
}

func TestListAccountsAddsTransactionsAfterTheLastClosing(t *testing.T) {
	uc, m := newTestUsecase(t)
	m.accounts.EXPECT().FindAll(gomock.Any()).Return([]domain.Account{{ID: 1, Name: "Tunai"}, {ID: 2, Name: "BSI"}, {ID: 3, Name: "BRI"}}, nil)
	expectBalances(m,
		domain.Transaction{Kind: domain.CategoryTypeIn, Amount: 250_000, AccountID: 1, PocketID: 1},
		domain.Transaction{Kind: domain.CategoryTypeOut, Amount: 100_000, AccountID: 2, PocketID: 2},
	)

	got, err := uc.ListAccounts(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := []int64{1_250_000, 500_000, 0}
	for i, a := range got {
		if a.Balance != want[i] {
			t.Errorf("%s balance = %d, want %d", a.Name, a.Balance, want[i])
		}
	}
}

func TestBalancesAreZeroBeforeBookkeepingStarts(t *testing.T) {
	uc, m := newTestUsecase(t)
	m.pockets.EXPECT().FindAll(gomock.Any()).Return([]domain.Pocket{{ID: 1}}, nil)
	m.closings.EXPECT().FindLatest(gomock.Any()).Return(domain.BookClosing{}, apperror.New(apperror.CodeNotFound, domain.ErrBookClosingNotFound, "belum ada"))

	got, err := uc.ListPockets(context.Background())
	if err != nil || got[0].Balance != 0 {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestArchiveAccount(t *testing.T) {
	tunai := domain.Account{ID: 1, Name: "Tunai", Kind: domain.AccountKindCash}
	bsi := domain.Account{ID: 2, Name: "BSI", Kind: domain.AccountKindBank}
	bri := domain.Account{ID: 3, Name: "BRI", Kind: domain.AccountKindBank}

	t.Run("an empty account is archived", func(t *testing.T) {
		uc, m := newTestUsecase(t)
		m.accounts.EXPECT().FindAll(gomock.Any()).Return([]domain.Account{tunai, bsi, bri}, nil)
		expectBalances(m)
		m.accounts.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, a domain.Account) (domain.Account, error) {
			if a.ArchivedAt == nil || !a.ArchivedAt.Equal(now) {
				t.Errorf("archived_at = %v", a.ArchivedAt)
			}
			return a, nil
		})
		m.activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityAccountArchived, EntityID: 3, Summary: `mengarsipkan akun "BRI"`})

		if _, err := uc.ArchiveAccount(context.Background(), 3); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("an account holding money stays", func(t *testing.T) {
		uc, m := newTestUsecase(t)
		m.accounts.EXPECT().FindAll(gomock.Any()).Return([]domain.Account{tunai, bsi, bri}, nil)
		expectBalances(m)

		_, err := uc.ArchiveAccount(context.Background(), 2)
		ruleError(t, err, apperror.CodeBusinessRule)
	})

	t.Run("the last cash account stays", func(t *testing.T) {
		uc, m := newTestUsecase(t)
		m.accounts.EXPECT().FindAll(gomock.Any()).Return([]domain.Account{tunai, bsi}, nil)

		_, err := uc.ArchiveAccount(context.Background(), 1)
		ruleError(t, err, apperror.CodeBusinessRule)
	})
}

func TestCreateAccountDropsTheNumberOfCash(t *testing.T) {
	uc, m := newTestUsecase(t)
	m.accounts.EXPECT().Create(gomock.Any(), domain.Account{Name: "Kotak Jumat", Kind: domain.AccountKindCash, CreatedBy: 7}).Return(domain.Account{ID: 4, Name: "Kotak Jumat"}, nil)
	m.activityLog.EXPECT().Publish(gomock.Any(), gomock.Any())

	if _, err := uc.CreateAccount(context.Background(), domain.CreateAccountRequest{CreatedBy: 7, Name: " Kotak  Jumat ", Kind: domain.AccountKindCash, Number: "123"}); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateBuiltInPocketOnlyChangesShowOnTV(t *testing.T) {
	uc, m := newTestUsecase(t)
	umum := domain.Pocket{ID: 1, Kind: domain.PocketKindUmum, Name: "Kas Umum", Status: domain.PocketStatusActive}
	m.pockets.EXPECT().FindByID(gomock.Any(), int64(1)).Return(umum, nil)
	want := umum
	want.ShowOnTV = true
	m.pockets.EXPECT().Update(gomock.Any(), want).Return(want, nil)
	m.activityLog.EXPECT().Publish(gomock.Any(), gomock.Any())

	req := domain.UpdatePocketRequest{ID: 1, TerikatPocketFields: domain.TerikatPocketFields{Name: "Kas Bebas", ShowOnTV: true}}
	if _, err := uc.UpdatePocket(context.Background(), req); err != nil {
		t.Fatal(err)
	}
}

func TestCreatePocketIsTerikat(t *testing.T) {
	uc, m := newTestUsecase(t)
	target := int64(3_000_000)
	plan := domain.SurplusPlanToUmum
	m.pockets.EXPECT().Create(gomock.Any(), domain.Pocket{
		Kind: domain.PocketKindTerikat, Name: "Renovasi Atap", Purpose: "Ganti genteng", TargetAmount: &target,
		SurplusPlan: &plan, Status: domain.PocketStatusActive, CreatedBy: 7,
	}).Return(domain.Pocket{ID: 5}, nil)
	m.activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityPocketCreated, EntityID: 5, Summary: `menambah kas "Renovasi Atap"`})

	req := domain.CreatePocketRequest{CreatedBy: 7, TerikatPocketFields: domain.TerikatPocketFields{
		Name: "Renovasi Atap", Purpose: "Ganti genteng", TargetAmount: &target, SurplusPlan: plan,
	}}
	if _, err := uc.CreatePocket(context.Background(), req); err != nil {
		t.Fatal(err)
	}

	req.SurplusPlan = ""
	_, err := uc.CreatePocket(context.Background(), req)
	ruleError(t, err, apperror.CodeInvalidInput)
}

func TestArchivePocket(t *testing.T) {
	t.Run("built-in kas stay", func(t *testing.T) {
		uc, m := newTestUsecase(t)
		m.pockets.EXPECT().FindByID(gomock.Any(), int64(2)).Return(domain.Pocket{ID: 2, Kind: domain.PocketKindZakatMal}, nil)

		_, err := uc.ArchivePocket(context.Background(), 2)
		ruleError(t, err, apperror.CodeBusinessRule)
	})

	t.Run("a terikat kas holding money stays", func(t *testing.T) {
		uc, m := newTestUsecase(t)
		m.pockets.EXPECT().FindByID(gomock.Any(), int64(5)).Return(domain.Pocket{ID: 5, Kind: domain.PocketKindTerikat, Status: domain.PocketStatusActive}, nil)
		expectBalances(m, domain.Transaction{Kind: domain.CategoryTypeIn, Amount: 400_000, AccountID: 2, PocketID: 5})

		_, err := uc.ArchivePocket(context.Background(), 5)
		ruleError(t, err, apperror.CodeBusinessRule)
	})

	t.Run("an empty terikat kas is archived", func(t *testing.T) {
		uc, m := newTestUsecase(t)
		m.pockets.EXPECT().FindByID(gomock.Any(), int64(5)).Return(domain.Pocket{ID: 5, Kind: domain.PocketKindTerikat, Name: "Renovasi", Status: domain.PocketStatusActive}, nil)
		expectBalances(m,
			domain.Transaction{Kind: domain.CategoryTypeIn, Amount: 400_000, AccountID: 2, PocketID: 5},
			domain.Transaction{Kind: domain.CategoryTypeOut, Amount: 400_000, AccountID: 2, PocketID: 5},
		)
		m.pockets.EXPECT().Update(gomock.Any(), domain.Pocket{ID: 5, Kind: domain.PocketKindTerikat, Name: "Renovasi", Status: domain.PocketStatusArchived}).
			Return(domain.Pocket{ID: 5, Name: "Renovasi", Status: domain.PocketStatusArchived}, nil)
		m.activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityPocketArchived, EntityID: 5, Summary: `mengarsipkan kas "Renovasi"`})

		if _, err := uc.ArchivePocket(context.Background(), 5); err != nil {
			t.Fatal(err)
		}
	})
}

func code(c string) *string { return &c }

func TestCategoryRules(t *testing.T) {
	pindah := domain.Category{ID: 1, Name: "Pindah dana", Type: domain.CategoryTypeOut, Scope: domain.CategoryScopeAll, Code: code(domain.CategoryCodePindahOut)}
	fakir := domain.Category{ID: 2, Name: "Fakir", Type: domain.CategoryTypeOut, Scope: domain.CategoryScopeZakat, Code: code("asnaf_fakir")}
	listrik := domain.Category{ID: 3, Name: "Listrik", Type: domain.CategoryTypeOut, Scope: domain.CategoryScopeUmumTerikat}

	t.Run("a system category keeps its name", func(t *testing.T) {
		uc, m := newTestUsecase(t)
		m.categories.EXPECT().FindByID(gomock.Any(), int64(2)).Return(fakir, nil)

		_, err := uc.UpdateCategory(context.Background(), domain.UpdateCategoryRequest{ID: 2, Name: "Faqir"})
		ruleError(t, err, apperror.CodeBusinessRule)
	})

	t.Run("a used category is archived, not deleted", func(t *testing.T) {
		uc, m := newTestUsecase(t)
		m.categories.EXPECT().FindByID(gomock.Any(), int64(3)).Return(listrik, nil)
		m.transactions.EXPECT().UsesCategory(gomock.Any(), int64(3)).Return(true, nil)

		ruleError(t, uc.DeleteCategory(context.Background(), 3), apperror.CodeBusinessRule)
	})

	t.Run("an unused category is deleted", func(t *testing.T) {
		uc, m := newTestUsecase(t)
		m.categories.EXPECT().FindByID(gomock.Any(), int64(3)).Return(listrik, nil)
		m.transactions.EXPECT().UsesCategory(gomock.Any(), int64(3)).Return(false, nil)
		m.categories.EXPECT().Delete(gomock.Any(), int64(3)).Return(nil)
		m.activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityCategoryDeleted, EntityID: 3, Summary: `menghapus kategori "Listrik"`})

		if err := uc.DeleteCategory(context.Background(), 3); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Pindah dana is never archived", func(t *testing.T) {
		uc, m := newTestUsecase(t)
		m.categories.EXPECT().FindByID(gomock.Any(), int64(1)).Return(pindah, nil)

		_, err := uc.ArchiveCategory(context.Background(), 1)
		ruleError(t, err, apperror.CodeBusinessRule)
	})

	t.Run("an asnaf may be hidden and restored", func(t *testing.T) {
		uc, m := newTestUsecase(t)
		m.categories.EXPECT().FindByID(gomock.Any(), int64(2)).Return(fakir, nil)
		m.categories.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, c domain.Category) (domain.Category, error) { return c, nil })
		m.activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityCategoryArchived, EntityID: 2, Summary: `mengarsipkan kategori "Fakir"`})

		archived, err := uc.ArchiveCategory(context.Background(), 2)
		if err != nil || archived.ArchivedAt == nil {
			t.Fatalf("archive: %+v, %v", archived, err)
		}

		m.categories.EXPECT().FindByID(gomock.Any(), int64(2)).Return(archived, nil)
		m.categories.EXPECT().Update(gomock.Any(), fakir).Return(fakir, nil)
		m.activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityCategoryRestored, EntityID: 2, Summary: `memulihkan kategori "Fakir"`})
		if _, err := uc.RestoreCategory(context.Background(), 2); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("counts transactions per category", func(t *testing.T) {
		uc, m := newTestUsecase(t)
		m.categories.EXPECT().FindAll(gomock.Any()).Return([]domain.Category{fakir, listrik}, nil)
		m.transactions.EXPECT().Find(gomock.Any(), domain.TransactionFilter{}).Return([]domain.Transaction{{CategoryID: 3}, {CategoryID: 3}, {CategoryID: 2}}, nil)

		got, err := uc.ListCategories(context.Background())
		if err != nil || got[0].TransactionCount != 1 || got[1].TransactionCount != 2 {
			t.Errorf("got %+v, %v", got, err)
		}
	})
}

func TestStartBookkeeping(t *testing.T) {
	archived := now
	accounts := []domain.Account{{ID: 1, Name: "Tunai"}, {ID: 2, Name: "BSI"}, {ID: 3, Name: "Lama", ArchivedAt: &archived}}
	pockets := []domain.Pocket{
		{ID: 1, Kind: domain.PocketKindUmum, Status: domain.PocketStatusActive},
		{ID: 2, Kind: domain.PocketKindZakatMal, Status: domain.PocketStatusActive},
	}
	valid := func() domain.StartBookkeepingRequest {
		return domain.StartBookkeepingRequest{
			ClosedBy: 7, StartMonth: "2026-11",
			Accounts: []domain.OpeningBalance{{ID: 1, Amount: 1_000_000}, {ID: 2, Amount: 600_000}},
			Pockets:  []domain.OpeningBalance{{ID: 1, Amount: 1_000_000}, {ID: 2, Amount: 600_000}},
		}
	}

	t.Run("saved as the closing of the month before", func(t *testing.T) {
		uc, m := newTestUsecase(t)
		m.profiles.EXPECT().Get(gomock.Any()).Return(domain.MosqueProfile{Timezone: domain.MosqueTimezoneWIB}, nil)
		m.closings.EXPECT().FindAll(gomock.Any()).Return(nil, nil)
		m.accounts.EXPECT().FindAll(gomock.Any()).Return(accounts, nil)
		m.pockets.EXPECT().FindAll(gomock.Any()).Return(pockets, nil)
		m.closings.EXPECT().Create(gomock.Any(), domain.BookClosing{
			Month: "2026-10", Kind: domain.BookClosingOpening, ClosingBalance: 1_600_000, ClosedBy: 7,
			Accounts: []domain.BookClosingAccount{{AccountID: 1, ClosingBalance: 1_000_000}, {AccountID: 2, ClosingBalance: 600_000}},
			Pockets:  []domain.BookClosingPocket{{PocketID: 1, ClosingBalance: 1_000_000}, {PocketID: 2, ClosingBalance: 600_000}},
		}).Return(opening, nil)
		m.activityLog.EXPECT().Publish(gomock.Any(), domain.ActivityLog{Action: domain.ActivityBookkeepingStarted, EntityID: 1, Summary: "memulai pembukuan dari Nov 2026"})

		got, err := uc.StartBookkeeping(context.Background(), valid())
		if err != nil || !got.Started || got.StartMonth != "2026-11" {
			t.Errorf("got %+v, %v", got, err)
		}
	})

	t.Run("only once", func(t *testing.T) {
		uc, m := newTestUsecase(t)
		m.profiles.EXPECT().Get(gomock.Any()).Return(domain.MosqueProfile{}, nil)
		m.closings.EXPECT().FindAll(gomock.Any()).Return([]domain.BookClosing{opening}, nil)

		_, err := uc.StartBookkeeping(context.Background(), valid())
		ruleError(t, err, apperror.CodeBusinessRule)
	})

	t.Run("not into an archived account", func(t *testing.T) {
		uc, m := newTestUsecase(t)
		m.profiles.EXPECT().Get(gomock.Any()).Return(domain.MosqueProfile{}, nil)
		m.closings.EXPECT().FindAll(gomock.Any()).Return(nil, nil)
		m.accounts.EXPECT().FindAll(gomock.Any()).Return(accounts, nil)
		req := valid()
		req.Accounts[1].ID = 3

		_, err := uc.StartBookkeeping(context.Background(), req)
		ruleError(t, err, apperror.CodeBusinessRule)
	})

	invalid := map[string]func(r *domain.StartBookkeepingRequest){
		"next month":       func(r *domain.StartBookkeepingRequest) { r.StartMonth = "2026-12" },
		"not a month":      func(r *domain.StartBookkeepingRequest) { r.StartMonth = "Nov 2026" },
		"totals differ":    func(r *domain.StartBookkeepingRequest) { r.Pockets[1].Amount = 500_000 },
		"negative":         func(r *domain.StartBookkeepingRequest) { r.Accounts[0].Amount = -1 },
		"repeated account": func(r *domain.StartBookkeepingRequest) { r.Accounts[1].ID = 1 },
		"above the max":    func(r *domain.StartBookkeepingRequest) { r.Accounts[0].Amount = domain.MaxTransactionAmount + 1 },
	}
	for name, change := range invalid {
		t.Run(name, func(t *testing.T) {
			uc, m := newTestUsecase(t)
			m.profiles.EXPECT().Get(gomock.Any()).Return(domain.MosqueProfile{}, apperror.New(apperror.CodeNotFound, domain.ErrMosqueProfileNotFound, "belum ada"))
			req := valid()
			change(&req)

			_, err := uc.StartBookkeeping(context.Background(), req)
			ruleError(t, err, apperror.CodeInvalidInput)
		})
	}
}
