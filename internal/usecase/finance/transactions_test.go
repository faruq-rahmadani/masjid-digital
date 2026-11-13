package finance

import (
	"context"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"go.uber.org/mock/gomock"
)

// fixture: Tunai and BSI; the four built-in kas and Renovasi Atap (terikat, target 3.000.000);
// the opening of `opening` (Tunai 1.000.000 in Kas Umum, BSI 600.000 in Zakat Mal)
var (
	fxAccounts = []domain.Account{{ID: 1, Name: "Tunai", Kind: domain.AccountKindCash}, {ID: 2, Name: "BSI", Kind: domain.AccountKindBank}}
	target     = int64(3_000_000)
	fxPockets  = []domain.Pocket{
		{ID: 1, Kind: domain.PocketKindUmum, Name: "Kas Umum", Status: domain.PocketStatusActive},
		{ID: 2, Kind: domain.PocketKindZakatMal, Name: "Zakat Mal", Status: domain.PocketStatusActive},
		{ID: 3, Kind: domain.PocketKindZakatFitrah, Name: "Zakat Fitrah", Status: domain.PocketStatusActive},
		{ID: 4, Kind: domain.PocketKindFidyah, Name: "Fidyah", Status: domain.PocketStatusActive},
		{ID: 5, Kind: domain.PocketKindTerikat, Name: "Renovasi Atap", Status: domain.PocketStatusActive, TargetAmount: &target},
	}
	fxCategories = []domain.Category{
		{ID: 10, Name: "Infaq Jumat", Type: domain.CategoryTypeIn, Scope: domain.CategoryScopeUmumTerikat},
		{ID: 11, Name: "Listrik", Type: domain.CategoryTypeOut, Scope: domain.CategoryScopeUmumTerikat},
		{ID: 12, Name: "Zakat Mal", Type: domain.CategoryTypeIn, Scope: domain.CategoryScopeZakat, Code: new(domain.CategoryCodeZakatMal)},
		{ID: 13, Name: "Fakir", Type: domain.CategoryTypeOut, Scope: domain.CategoryScopeZakat, Code: new("asnaf_fakir")},
		{ID: 14, Name: "Amil", Type: domain.CategoryTypeOut, Scope: domain.CategoryScopeZakat, Code: new(domain.CategoryCodeAmil)},
		{ID: 20, Name: "Pindah dana", Type: domain.CategoryTypeOut, Scope: domain.CategoryScopeAll, Code: new(domain.CategoryCodePindahOut)},
		{ID: 21, Name: "Pindah dana", Type: domain.CategoryTypeIn, Scope: domain.CategoryScopeAll, Code: new(domain.CategoryCodePindahIn)},
	}
)

// fixture sets every read the ledger makes; txs are the transactions after the opening
func fixture(m mocks, txs ...domain.Transaction) {
	m.closings.EXPECT().FindLatest(gomock.Any()).Return(opening, nil).AnyTimes()
	m.closings.EXPECT().FindAll(gomock.Any()).Return([]domain.BookClosing{opening}, nil).AnyTimes()
	m.profiles.EXPECT().Get(gomock.Any()).Return(domain.MosqueProfile{Timezone: domain.MosqueTimezoneWIB}, nil).AnyTimes()
	m.transactions.EXPECT().Find(gomock.Any(), domain.TransactionFilter{From: "2026-11-01"}).Return(txs, nil).AnyTimes()
	m.accounts.EXPECT().FindAll(gomock.Any()).Return(fxAccounts, nil).AnyTimes()
	m.pockets.EXPECT().FindAll(gomock.Any()).Return(fxPockets, nil).AnyTimes()
	m.categories.EXPECT().FindAll(gomock.Any()).Return(fxCategories, nil).AnyTimes()
	m.categories.EXPECT().FindByCode(gomock.Any(), domain.CategoryCodePindahOut).Return(fxCategories[5], nil).AnyTimes()
	m.categories.EXPECT().FindByCode(gomock.Any(), domain.CategoryCodePindahIn).Return(fxCategories[6], nil).AnyTimes()
	m.activityLog.EXPECT().Publish(gomock.Any(), gomock.Any()).AnyTimes()
}

func income(pocket, category, amount int64) domain.SaveTransactionRequest {
	return domain.SaveTransactionRequest{UserID: 7, Kind: domain.CategoryTypeIn, Date: "2026-11-06", AccountID: 1,
		Lines: []domain.TransactionLine{{PocketID: pocket, CategoryID: category, Amount: amount}}}
}

func echoCreate(m mocks) {
	id := int64(100)
	m.transactions.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, tx domain.Transaction) (domain.Transaction, error) {
		id++
		tx.ID = id
		return tx, nil
	}).AnyTimes()
}

func TestCreateTransactionRules(t *testing.T) {
	expense := income(1, 11, 1_200_000)
	expense.Kind = domain.CategoryTypeOut

	tests := []struct {
		name string
		req  domain.SaveTransactionRequest
		code apperror.Code
	}{
		{"infaq into Kas Umum", income(1, 10, 500_000), ""},
		{"infaq into Zakat Mal", income(2, 10, 500_000), apperror.CodeBusinessRule},
		{"zakat mal into Kas Umum", income(1, 12, 500_000), apperror.CodeBusinessRule},
		{"Pindah dana as a transaction", income(1, 21, 500_000), apperror.CodeBusinessRule},
		{"expense above the balance", expense, apperror.CodeBusinessRule},
		{"dated tomorrow", func() domain.SaveTransactionRequest { r := income(1, 10, 1); r.Date = "2026-11-12"; return r }(), apperror.CodeInvalidInput},
		{"dated in the closed month", func() domain.SaveTransactionRequest { r := income(1, 10, 1); r.Date = "2026-10-31"; return r }(), apperror.CodeBusinessRule},
		{"above the max", income(1, 10, domain.MaxTransactionAmount+1), apperror.CodeInvalidInput},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, m := newTestUsecase(t)
			fixture(m)
			echoCreate(m)

			_, err := uc.CreateTransaction(context.Background(), tt.req)
			if tt.code == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			ruleError(t, err, tt.code)
		})
	}
}

func TestSplitIncomeSharesAGroup(t *testing.T) {
	uc, m := newTestUsecase(t)
	fixture(m)
	echoCreate(m)
	req := domain.SaveTransactionRequest{UserID: 7, Kind: domain.CategoryTypeIn, Date: "2026-11-07", AccountID: 2, Lines: []domain.TransactionLine{
		{PocketID: 2, CategoryID: 12, Amount: 600_000}, {PocketID: 5, CategoryID: 10, Amount: 400_000},
	}}

	rows, err := uc.CreateTransaction(context.Background(), req)
	if err != nil || len(rows) != 2 || rows[0].GroupID == nil || rows[1].GroupID == nil || *rows[0].GroupID != *rows[1].GroupID {
		t.Fatalf("rows %+v, %v", rows, err)
	}
}

func TestMoveFundsRules(t *testing.T) {
	move := func(from, fromAccount, to, toAccount int64) domain.MoveFundsRequest {
		return domain.MoveFundsRequest{UserID: 7, FromPocketID: from, FromAccountID: fromAccount, ToPocketID: to, ToAccountID: toAccount, Amount: 100_000, Date: "2026-11-08"}
	}
	fromTerikat := move(5, 1, 1, 1)
	consented := fromTerikat
	consented.DonorConsent, consented.Note = true, "Atap selesai, donatur setuju"

	tests := []struct {
		name string
		req  domain.MoveFundsRequest
		code apperror.Code
	}{
		{"Kas Umum to terikat is free", move(1, 1, 5, 1), ""},
		{"setor tunai within Zakat Mal", move(2, 2, 2, 1), ""},
		{"out of Zakat Mal", move(2, 2, 1, 2), apperror.CodeBusinessRule},
		{"into Fidyah", move(1, 1, 4, 1), apperror.CodeBusinessRule},
		{"out of terikat without consent", fromTerikat, apperror.CodeBusinessRule},
		{"out of terikat with consent", consented, ""},
		{"same kas and account", move(1, 1, 1, 1), apperror.CodeInvalidInput},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, m := newTestUsecase(t)
			// Renovasi holds 200.000 from Tunai so it can give
			fixture(m, domain.Transaction{ID: 50, Kind: domain.CategoryTypeIn, Amount: 200_000, Date: "2026-11-02", AccountID: 1, PocketID: 5, CategoryID: 10})
			echoCreate(m)

			rows, err := uc.MoveFunds(context.Background(), tt.req)
			if tt.code == "" {
				if err != nil || len(rows) != 2 || rows[0].Kind != domain.CategoryTypeOut || rows[0].CategoryID != 20 || rows[1].CategoryID != 21 {
					t.Fatalf("rows %+v, %v", rows, err)
				}
				return
			}
			ruleError(t, err, tt.code)
		})
	}
}

func TestDeleteTransactionKeepsRevisionsAndBalances(t *testing.T) {
	group := "g1"
	split := []domain.Transaction{
		{ID: 60, Kind: domain.CategoryTypeIn, Amount: 300_000, Date: "2026-11-03", AccountID: 2, PocketID: 2, CategoryID: 12, GroupID: &group},
		{ID: 61, Kind: domain.CategoryTypeIn, Amount: 200_000, Date: "2026-11-03", AccountID: 2, PocketID: 5, CategoryID: 10, GroupID: &group},
	}

	t.Run("the whole split goes, each row with a revision", func(t *testing.T) {
		uc, m := newTestUsecase(t)
		fixture(m, split...)
		m.transactions.EXPECT().FindByID(gomock.Any(), int64(61)).Return(split[1], nil)
		m.transactions.EXPECT().Find(gomock.Any(), domain.TransactionFilter{GroupID: "g1"}).Return(split, nil)
		m.transactions.EXPECT().AddRevision(gomock.Any(), gomock.Any()).Return(nil).Times(2)
		m.transactions.EXPECT().SoftDelete(gomock.Any(), int64(60), int64(7)).Return(nil)
		m.transactions.EXPECT().SoftDelete(gomock.Any(), int64(61), int64(7)).Return(nil)

		if err := uc.DeleteTransaction(context.Background(), domain.DeleteTransactionRequest{UserID: 7, ID: 61, Reason: "Salah input"}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("not when the money was already spent", func(t *testing.T) {
		uc, m := newTestUsecase(t)
		spent := domain.Transaction{ID: 62, Kind: domain.CategoryTypeOut, Amount: 150_000, Date: "2026-11-04", AccountID: 2, PocketID: 5, CategoryID: 11}
		fixture(m, split[0], split[1], spent)
		m.transactions.EXPECT().FindByID(gomock.Any(), int64(61)).Return(split[1], nil)
		m.transactions.EXPECT().Find(gomock.Any(), domain.TransactionFilter{GroupID: "g1"}).Return(split, nil)

		ruleError(t, uc.DeleteTransaction(context.Background(), domain.DeleteTransactionRequest{UserID: 7, ID: 61, Reason: "Salah input"}), apperror.CodeBusinessRule)
	})

	t.Run("a reason is required", func(t *testing.T) {
		uc, _ := newTestUsecase(t)
		ruleError(t, uc.DeleteTransaction(context.Background(), domain.DeleteTransactionRequest{ID: 61}), apperror.CodeInvalidInput)
	})
}

func TestUpdateTransaction(t *testing.T) {
	saved := domain.Transaction{ID: 70, Kind: domain.CategoryTypeIn, Amount: 100_000, Date: "2026-11-05", AccountID: 1, PocketID: 1, CategoryID: 10}

	t.Run("keeps the old values as a revision", func(t *testing.T) {
		uc, m := newTestUsecase(t)
		fixture(m, saved)
		m.transactions.EXPECT().FindByID(gomock.Any(), int64(70)).Return(saved, nil)
		m.transactions.EXPECT().AddRevision(gomock.Any(), domain.TransactionRevision{TransactionID: 70, Action: domain.RevisionEdit, Previous: saved, Reason: "Nominal salah", UserID: 7}).Return(nil)
		m.transactions.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, tx domain.Transaction) (domain.Transaction, error) {
			if tx.Amount != 150_000 || tx.ID != 70 {
				t.Errorf("saved %+v", tx)
			}
			return tx, nil
		})

		req := income(1, 10, 150_000)
		req.ID, req.Date, req.Reason = 70, "2026-11-05", "Nominal salah"
		if _, err := uc.UpdateTransaction(context.Background(), req); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("the kind never changes", func(t *testing.T) {
		uc, m := newTestUsecase(t)
		fixture(m, saved)
		m.transactions.EXPECT().FindByID(gomock.Any(), int64(70)).Return(saved, nil)
		req := income(1, 11, 50_000)
		req.ID, req.Kind, req.Reason = 70, domain.CategoryTypeOut, "Salah jenis"

		_, err := uc.UpdateTransaction(context.Background(), req)
		ruleError(t, err, apperror.CodeBusinessRule)
	})
}

func TestCloseMonth(t *testing.T) {
	novemberTx := domain.Transaction{ID: 80, Kind: domain.CategoryTypeIn, Amount: 250_000, Date: "2026-11-06", AccountID: 1, PocketID: 1, CategoryID: 10}
	octoberOpening := opening
	octoberOpening.Month = "2026-09" // so October has ended and is next

	setup := func(t *testing.T) (*UseCase, mocks) {
		uc, m := newTestUsecase(t)
		m.closings.EXPECT().FindLatest(gomock.Any()).Return(octoberOpening, nil).AnyTimes()
		m.closings.EXPECT().FindAll(gomock.Any()).Return([]domain.BookClosing{octoberOpening}, nil).AnyTimes()
		m.profiles.EXPECT().Get(gomock.Any()).Return(domain.MosqueProfile{}, nil).AnyTimes()
		m.categories.EXPECT().FindAll(gomock.Any()).Return(fxCategories, nil).AnyTimes()
		m.accounts.EXPECT().FindAll(gomock.Any()).Return(fxAccounts, nil).AnyTimes()
		m.transactions.EXPECT().Find(gomock.Any(), domain.TransactionFilter{From: "2026-10-01", To: "2026-10-31"}).Return([]domain.Transaction{
			{ID: 81, Kind: domain.CategoryTypeOut, Amount: 100_000, Date: "2026-10-20", AccountID: 1, PocketID: 1, CategoryID: 11},
		}, nil).AnyTimes()
		m.activityLog.EXPECT().Publish(gomock.Any(), gomock.Any()).AnyTimes()
		_ = novemberTx
		return uc, m
	}

	t.Run("in order", func(t *testing.T) {
		uc, _ := setup(t)
		_, err := uc.CloseMonth(context.Background(), domain.CloseMonthRequest{Month: "2026-11"})
		ruleError(t, err, apperror.CodeBusinessRule)
	})

	t.Run("a counted difference is refused", func(t *testing.T) {
		uc, _ := setup(t)
		_, err := uc.CloseMonth(context.Background(), domain.CloseMonthRequest{Month: "2026-10", Accounts: []domain.CountedBalance{{AccountID: 1, Counted: 1_000_000}, {AccountID: 2, Counted: 600_000}}})
		ruleError(t, err, apperror.CodeBusinessRule)
	})

	t.Run("matching counts close the month with its totals", func(t *testing.T) {
		uc, m := setup(t)
		m.closings.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, c domain.BookClosing) (domain.BookClosing, error) {
			if c.Month != "2026-10" || c.TotalOut != 100_000 || c.ClosingBalance != 1_500_000 || len(c.Accounts) != 2 || c.Accounts[0].ClosingBalance != 900_000 {
				t.Errorf("closing %+v", c)
			}
			return c, nil
		})
		_, err := uc.CloseMonth(context.Background(), domain.CloseMonthRequest{UserID: 7, Month: "2026-10", Accounts: []domain.CountedBalance{{AccountID: 1, Counted: 900_000}, {AccountID: 2, Counted: 600_000}}})
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestSummaryLeavesPindahOutOfTotalsAndWarnsAboutAmil(t *testing.T) {
	uc, m := newTestUsecase(t)
	pair := "p1"
	txs := []domain.Transaction{
		{ID: 90, Kind: domain.CategoryTypeIn, Amount: 400_000, Date: "2026-11-11", AccountID: 1, PocketID: 1, CategoryID: 10},
		{ID: 91, Kind: domain.CategoryTypeOut, Amount: 100_000, Date: "2026-11-09", AccountID: 1, PocketID: 1, CategoryID: 20, GroupID: &pair},
		{ID: 92, Kind: domain.CategoryTypeIn, Amount: 100_000, Date: "2026-11-09", AccountID: 1, PocketID: 5, CategoryID: 21, GroupID: &pair},
		{ID: 93, Kind: domain.CategoryTypeOut, Amount: 100_000, Date: "2026-11-10", AccountID: 2, PocketID: 2, CategoryID: 14},
	}
	fixture(m, txs...)
	m.transactions.EXPECT().Find(gomock.Any(), domain.TransactionFilter{From: "2026-11-01", To: "2026-11-30"}).Return(txs, nil)
	m.prayerTimes.EXPECT().GetSettings(gomock.Any()).Return(domain.PrayerSettings{}, nil)
	m.categories.EXPECT().FindByCode(gomock.Any(), domain.CategoryCodeAmil).Return(fxCategories[4], nil)
	// this Hijri year: Zakat Mal took in 600.000 (1/8 = 75.000) and paid amil 100.000
	m.transactions.EXPECT().Find(gomock.Any(), gomock.Cond(func(x any) bool { f := x.(domain.TransactionFilter); return f.PocketID == 2 })).
		Return([]domain.Transaction{{Kind: domain.CategoryTypeIn, Amount: 600_000, PocketID: 2, CategoryID: 12}, txs[3]}, nil)
	m.transactions.EXPECT().Find(gomock.Any(), gomock.Cond(func(x any) bool { f := x.(domain.TransactionFilter); return f.PocketID == 3 })).Return(nil, nil)

	got, err := uc.Summary(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Month != "2026-11" || got.TotalIn != 400_000 || got.TotalOut != 100_000 {
		t.Errorf("totals %+v", got)
	}
	umum := got.Pockets[0]
	if umum.Opening != 1_000_000 || umum.In != 400_000 || umum.Moved != -100_000 || umum.Closing != 1_300_000 || umum.TodayIn != 400_000 {
		t.Errorf("kas umum %+v", umum)
	}
	if len(got.Warnings) != 1 || got.Warnings[0].Kind != domain.WarningAmilCap || got.Warnings[0].PocketID != 2 {
		t.Errorf("warnings %+v", got.Warnings)
	}
}

func TestTVPeriodFrom(t *testing.T) {
	wednesday := now // 2026-11-11 is a Wednesday
	for period, want := range map[domain.TVPeriod]string{domain.TVPeriodToday: "2026-11-11", domain.TVPeriodWeek: "2026-11-09", domain.TVPeriodMonth: "2026-11-01"} {
		if got := period.From(wednesday); got != want {
			t.Errorf("%s from = %s, want %s", period, got, want)
		}
	}
}

func TestTVColumnsAddUpWithPindah(t *testing.T) {
	f := domain.Flow{Opening: 1_000_000, Closing: 1_000_000}
	f.Add(domain.Transaction{Kind: domain.CategoryTypeIn, Amount: 400_000}, false)
	f.Add(domain.Transaction{Kind: domain.CategoryTypeOut, Amount: 100_000}, true)
	tv := f.WithMoves()
	if tv.Opening+tv.In-tv.Out != tv.Closing || tv.In != 400_000 || tv.Out != 100_000 {
		t.Errorf("tv flow %+v", tv)
	}
	if f.In != 400_000 || f.Out != 0 || f.Moved != -100_000 {
		t.Errorf("summary flow %+v", f)
	}
}
