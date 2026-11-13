package finance

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	hijri "github.com/hablullah/go-hijri"
)

// flows is how every kas and account moved from..to (YYYY-MM-DD, inclusive), starting from the
// latest closing before from. Before pembukuan starts everything is zero.
type flows struct {
	pockets  map[int64]*domain.Flow
	accounts map[int64]*domain.Flow
	txs      []domain.Transaction // from..to
	pindah   map[int64]bool       // Pindah dana category IDs
}

func (u *UseCase) flows(ctx context.Context, from, to string) (flows, error) {
	f := flows{pockets: map[int64]*domain.Flow{}, accounts: map[int64]*domain.Flow{}, pindah: map[int64]bool{}}
	categories, err := u.Categories.FindAll(ctx)
	if err != nil {
		return flows{}, err
	}
	for _, c := range categories {
		if c.IsPindah() {
			f.pindah[c.ID] = true
		}
	}

	closings, err := u.BookClosings.FindAll(ctx)
	if err != nil || len(closings) == 0 {
		return f, err
	}
	base := closings[0]
	for _, c := range closings {
		if domain.FirstDayAfter(c.Month) <= from {
			base = c
		}
	}
	for _, a := range base.Accounts {
		f.account(a.AccountID).Opening = a.ClosingBalance
	}
	for _, p := range base.Pockets {
		f.pocket(p.PocketID).Opening = p.ClosingBalance
	}

	txs, err := u.Transactions.Find(ctx, domain.TransactionFilter{From: domain.FirstDayAfter(base.Month), To: to})
	if err != nil {
		return flows{}, err
	}
	for _, tx := range txs {
		if tx.Date < from {
			f.pocket(tx.PocketID).Opening += tx.Signed()
			f.account(tx.AccountID).Opening += tx.Signed()
			continue
		}
		f.txs = append(f.txs, tx)
		f.pocket(tx.PocketID).Add(tx, f.pindah[tx.CategoryID])
		f.account(tx.AccountID).Add(tx, f.pindah[tx.CategoryID])
	}
	for _, p := range f.pockets {
		p.Closing += p.Opening
	}
	for _, a := range f.accounts {
		a.Closing += a.Opening
	}
	return f, nil
}

func (f *flows) pocket(id int64) *domain.Flow {
	if f.pockets[id] == nil {
		f.pockets[id] = &domain.Flow{}
	}
	return f.pockets[id]
}

func (f *flows) account(id int64) *domain.Flow {
	if f.accounts[id] == nil {
		f.accounts[id] = &domain.Flow{}
	}
	return f.accounts[id]
}

// Summary is one month of Keuangan (the current one when month is empty): per kas and account,
// the month's totals without Pindah dana, today's in and out, and the warnings that apply today
func (u *UseCase) Summary(ctx context.Context, month string) (domain.FinanceSummary, error) {
	now, err := u.today(ctx)
	if err != nil {
		return domain.FinanceSummary{}, err
	}
	today := now.Format("2006-01-02")
	if month == "" {
		month = now.Format("2006-01")
	}
	start, err := time.Parse("2006-01", month)
	if err != nil {
		return domain.FinanceSummary{}, invalidInput(err, "bulan tidak valid")
	}

	f, err := u.flows(ctx, start.Format("2006-01-02"), start.AddDate(0, 1, -1).Format("2006-01-02"))
	if err != nil {
		return domain.FinanceSummary{}, err
	}
	summary := domain.FinanceSummary{Month: month, Today: today, Pockets: []domain.PocketFlow{}, Accounts: []domain.AccountFlow{}}

	latest, err := u.BookClosings.FindLatest(ctx)
	if err != nil && !errors.Is(err, domain.ErrBookClosingNotFound) {
		return domain.FinanceSummary{}, err
	}
	summary.Closed = err == nil && month <= latest.Month

	pockets, err := u.Pockets.FindAll(ctx)
	if err != nil {
		return domain.FinanceSummary{}, err
	}
	for _, p := range pockets {
		pf := domain.PocketFlow{PocketID: p.ID, Flow: *f.pocket(p.ID)}
		for _, tx := range f.txs {
			if tx.PocketID != p.ID || tx.Date != today || f.pindah[tx.CategoryID] {
				continue
			}
			if tx.Kind == domain.CategoryTypeIn {
				pf.TodayIn += tx.Amount
				continue
			}
			pf.TodayOut += tx.Amount
		}
		summary.TotalIn += pf.In
		summary.TotalOut += pf.Out
		summary.TodayIn += pf.TodayIn
		summary.TodayOut += pf.TodayOut
		summary.Balance += pf.Closing
		summary.Pockets = append(summary.Pockets, pf)
	}

	accounts, err := u.Accounts.FindAll(ctx)
	if err != nil {
		return domain.FinanceSummary{}, err
	}
	for _, a := range accounts {
		summary.Accounts = append(summary.Accounts, domain.AccountFlow{AccountID: a.ID, Flow: *f.account(a.ID)})
	}

	summary.Warnings, err = u.warnings(ctx, now, pockets)
	return summary, err
}

// progress is how much came into each terikat kas with a target, since it was created
func (u *UseCase) progress(ctx context.Context, pockets []domain.Pocket) (map[int64]domain.PocketProgress, error) {
	progress := map[int64]domain.PocketProgress{}
	for _, p := range pockets {
		if p.Kind != domain.PocketKindTerikat || p.TargetAmount == nil {
			continue
		}
		txs, err := u.Transactions.Find(ctx, domain.TransactionFilter{PocketID: p.ID})
		if err != nil {
			return nil, err
		}
		var collected int64
		for _, tx := range txs {
			if tx.Kind == domain.CategoryTypeIn {
				collected += tx.Amount
			}
		}
		progress[p.ID] = domain.PocketProgress{Collected: collected, TargetReached: collected >= *p.TargetAmount}
	}
	return progress, nil
}

// FinishPocket is "Selesaikan program": the remainder moves to Kas Umum as a Pindah dana pair,
// then the kas is archived. With rencana sisa dana "ask_donor" the donors' consent and a note
// are required.
func (u *UseCase) FinishPocket(ctx context.Context, req domain.FinishPocketRequest) (domain.Pocket, error) {
	l, err := u.ledger(ctx)
	if err != nil {
		return domain.Pocket{}, err
	}
	pocket, ok := l.pockets[req.ID]
	if !ok {
		return domain.Pocket{}, apperror.New(apperror.CodeNotFound, domain.ErrPocketNotFound, "kas tidak ditemukan")
	}
	if pocket.Kind != domain.PocketKindTerikat || !pocket.IsOpen() {
		return domain.Pocket{}, errRule(errors.New("not an open terikat pocket"), "hanya kas terikat yang masih aktif yang bisa diselesaikan")
	}
	plan := domain.SurplusPlanToUmum
	if pocket.SurplusPlan != nil {
		plan = *pocket.SurplusPlan
	}
	if err := req.ValidateAndNormalize(plan); err != nil {
		return domain.Pocket{}, err
	}

	var umum domain.Pocket
	for _, p := range l.pockets {
		if p.Kind == domain.PocketKindUmum {
			umum = p
		}
	}
	remainder := l.balances.pockets[pocket.ID]
	var rows []domain.Transaction
	if remainder > 0 {
		outCategory, inCategory, err := u.pindahCategories(ctx)
		if err != nil {
			return domain.Pocket{}, err
		}
		note := "Selesaikan program " + pocket.Name
		if req.Note != "" {
			note += ": " + req.Note
		}
		move := domain.MoveFundsRequest{UserID: req.UserID, FromPocketID: pocket.ID, ToPocketID: umum.ID, FromAccountID: req.AccountID,
			ToAccountID: req.AccountID, Amount: remainder, Date: l.today, Note: note, DonorConsent: true}
		rows = move.Rows(newGroupID(), outCategory, inCategory)
		if err := l.check(nil, rows, true); err != nil {
			return domain.Pocket{}, err
		}
	}

	pocket.Status = domain.PocketStatusArchived
	var saved domain.Pocket
	err = u.Transactor.WithTx(ctx, func(ctx context.Context) error {
		for _, r := range rows {
			if _, err := u.Transactions.Create(ctx, r); err != nil {
				return err
			}
		}
		var err error
		saved, err = u.Pockets.Update(ctx, pocket)
		return err
	})
	if err != nil {
		return domain.Pocket{}, err
	}
	u.activityLog.Publish(ctx, domain.PocketFinishedActivityLog(saved, remainder))
	return saved, nil
}

// ListClosings returns the opening and every closed month, oldest first
func (u *UseCase) ListClosings(ctx context.Context) ([]domain.BookClosing, error) {
	return u.BookClosings.FindAll(ctx)
}

// CloseMonth is Tutup buku: only the month after the last closing, only once it has ended, and
// only when every account's counted balance equals its computed one. It keeps the month's totals
// and every balance, and locks the month.
func (u *UseCase) CloseMonth(ctx context.Context, req domain.CloseMonthRequest) (domain.BookClosing, error) {
	if err := req.ValidateAndNormalize(); err != nil {
		return domain.BookClosing{}, err
	}
	latest, err := u.BookClosings.FindLatest(ctx)
	if errors.Is(err, domain.ErrBookClosingNotFound) {
		return domain.BookClosing{}, errRule(err, "mulai pembukuan dulu di menu Akun")
	}
	if err != nil {
		return domain.BookClosing{}, err
	}
	if next := domain.MonthAfter(latest.Month); req.Month != next {
		return domain.BookClosing{}, errRule(fmt.Errorf("closing %s, next is %s", req.Month, next), "tutup buku harus berurutan, berikutnya "+next)
	}
	now, err := u.today(ctx)
	if err != nil {
		return domain.BookClosing{}, err
	}
	if req.Month >= now.Format("2006-01") {
		return domain.BookClosing{}, errRule(errors.New("month not ended"), "bulan ini belum berakhir")
	}

	start, _ := time.Parse("2006-01", req.Month)
	f, err := u.flows(ctx, start.Format("2006-01-02"), start.AddDate(0, 1, -1).Format("2006-01-02"))
	if err != nil {
		return domain.BookClosing{}, err
	}
	accounts, err := u.Accounts.FindAll(ctx)
	if err != nil {
		return domain.BookClosing{}, err
	}
	counted := map[int64]int64{}
	given := map[int64]bool{}
	for _, c := range req.Accounts {
		counted[c.AccountID], given[c.AccountID] = c.Counted, true
	}

	closing := domain.BookClosing{Month: req.Month, Kind: domain.BookClosingClosing, ClosedBy: req.UserID}
	for _, a := range accounts {
		computed := f.account(a.ID).Closing
		if a.ArchivedAt != nil && computed == 0 && !given[a.ID] {
			continue
		}
		if !given[a.ID] {
			return domain.BookClosing{}, invalidInput(fmt.Errorf("account %d not counted", a.ID), "isi saldo hitung akun "+a.Name)
		}
		if diff := counted[a.ID] - computed; diff != 0 {
			return domain.BookClosing{}, errRule(fmt.Errorf("account %d differs by %d", a.ID, diff),
				fmt.Sprintf("saldo hitung akun %s selisih %s dari catatan (%s); catat transaksi penyesuaian dulu", a.Name, domain.RupiahForMessage(diff), domain.RupiahForMessage(computed)))
		}
		c := counted[a.ID]
		closing.Accounts = append(closing.Accounts, domain.BookClosingAccount{AccountID: a.ID, ClosingBalance: computed, CountedBalance: &c})
		closing.ClosingBalance += computed
	}
	ids := make([]int64, 0, len(f.pockets))
	for id := range f.pockets {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	for _, id := range ids {
		p := f.pockets[id]
		closing.TotalIn += p.In
		closing.TotalOut += p.Out
		closing.Pockets = append(closing.Pockets, domain.BookClosingPocket{PocketID: id, ClosingBalance: p.Closing})
	}

	var saved domain.BookClosing
	err = u.Transactor.WithTx(ctx, func(ctx context.Context) error {
		var err error
		saved, err = u.BookClosings.Create(ctx, closing)
		return err
	})
	if err != nil {
		return domain.BookClosing{}, err
	}
	u.activityLog.Publish(ctx, req.ToActivityLog(saved.ID))
	return saved, nil
}

// warnings are the syariah warnings that apply now: amil above 1/8 of a zakat kas's income this
// Hijri year, and Zakat Fitrah still holding money after 1 Syawal
func (u *UseCase) warnings(ctx context.Context, now time.Time, pockets []domain.Pocket) ([]domain.FinanceWarning, error) {
	warnings := []domain.FinanceWarning{}
	offset, err := u.hijriOffset(ctx, now)
	if err != nil {
		return nil, err
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	h, err := hijri.CreateUmmAlQuraDate(today.AddDate(0, 0, offset))
	if err != nil {
		return warnings, nil // outside 1356–1500 H: no warnings rather than a failed page
	}
	yearStart := hijri.UmmAlQuraDate{Day: 1, Month: 1, Year: h.Year}.ToGregorian().AddDate(0, 0, -offset).Format("2006-01-02")

	amil, err := u.Categories.FindByCode(ctx, domain.CategoryCodeAmil)
	if err != nil {
		return nil, err
	}
	b, err := u.balances(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range pockets {
		if p.Kind == domain.PocketKindZakatFitrah && b.pockets[p.ID] > 0 && (h.Month >= 10 || h.Month <= 7) {
			warnings = append(warnings, domain.FinanceWarning{Kind: domain.WarningFitrahDeadline, PocketID: p.ID,
				Message: fmt.Sprintf("Zakat Fitrah masih tersisa %s, seharusnya sudah disalurkan sebelum sholat Idul Fitri", domain.RupiahForMessage(b.pockets[p.ID]))})
		}
		if p.Kind != domain.PocketKindZakatMal && p.Kind != domain.PocketKindZakatFitrah {
			continue
		}
		txs, err := u.Transactions.Find(ctx, domain.TransactionFilter{PocketID: p.ID, From: yearStart})
		if err != nil {
			return nil, err
		}
		var income, amilOut int64
		for _, tx := range txs {
			if tx.Kind == domain.CategoryTypeIn {
				income += tx.Amount
			}
			if tx.CategoryID == amil.ID {
				amilOut += tx.Amount
			}
		}
		if amilOut*8 > income {
			warnings = append(warnings, domain.FinanceWarning{Kind: domain.WarningAmilCap, PocketID: p.ID,
				Message: fmt.Sprintf("Bagian amil %s melebihi 1/8 pemasukan %s tahun %d H (%s)", domain.RupiahForMessage(amilOut), p.Name, h.Year, domain.RupiahForMessage(income/8))})
		}
	}
	return warnings, nil
}

// hijriOffset is the mosque's Hijri correction in days on the given day
func (u *UseCase) hijriOffset(ctx context.Context, now time.Time) (int, error) {
	settings, err := u.PrayerTimes.GetSettings(ctx)
	if errors.Is(err, domain.ErrPrayerSettingsNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if settings.HijriOffsetFrom != "" && now.Format("2006-01-02") < settings.HijriOffsetFrom {
		return 0, nil
	}
	return settings.HijriOffsetDays, nil
}

func (u *UseCase) GetTVSetting(ctx context.Context) (domain.FinanceTVSetting, error) {
	setting, err := u.Settings.GetTVSetting(ctx)
	if errors.Is(err, domain.ErrFinanceTVSettingNotFound) {
		return domain.DefaultFinanceTVSetting(), nil
	}
	return setting, err
}

func (u *UseCase) UpdateTVSetting(ctx context.Context, req domain.UpdateFinanceTVSettingRequest) (domain.FinanceTVSetting, error) {
	if err := req.ValidateAndNormalize(); err != nil {
		return domain.FinanceTVSetting{}, err
	}
	if err := u.Settings.UpdateTVSetting(ctx, req.FinanceTVSetting, req.UpdatedBy); err != nil {
		return domain.FinanceTVSetting{}, err
	}
	u.activityLog.Publish(ctx, req.ToActivityLog())
	return req.FinanceTVSetting, nil
}

// GetDisplay is what the TV shows: each open kas marked tampil di TV with its four columns over
// the period set by pengurus
func (u *UseCase) GetDisplay(ctx context.Context) (domain.FinanceDisplay, error) {
	setting, err := u.GetTVSetting(ctx)
	if err != nil {
		return domain.FinanceDisplay{}, err
	}
	now, err := u.today(ctx)
	if err != nil {
		return domain.FinanceDisplay{}, err
	}
	display := domain.FinanceDisplay{Period: setting.Period, From: setting.Period.From(now), To: now.Format("2006-01-02"), Pockets: []domain.TVPocket{}}

	f, err := u.flows(ctx, display.From, display.To)
	if err != nil {
		return domain.FinanceDisplay{}, err
	}
	pockets, err := u.Pockets.FindAll(ctx)
	if err != nil {
		return domain.FinanceDisplay{}, err
	}
	for _, p := range pockets {
		if p.ShowOnTV && p.IsOpen() {
			display.Pockets = append(display.Pockets, domain.TVPocket{ID: p.ID, Name: p.Name, Kind: p.Kind, Flow: f.pocket(p.ID).WithMoves()})
		}
	}
	return display, nil
}
