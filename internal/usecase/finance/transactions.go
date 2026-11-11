package finance

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

// ledger is what every write is checked against: the last closing, the first date still open,
// today, every balance now and the accounts, kas and categories by ID
type ledger struct {
	openFrom   string // YYYY-MM-DD; earlier dates belong to a closed month
	today      string
	balances   balances
	accounts   map[int64]domain.Account
	pockets    map[int64]domain.Pocket
	categories map[int64]domain.Category
}

func (u *UseCase) ledger(ctx context.Context) (ledger, error) {
	closing, err := u.BookClosings.FindLatest(ctx)
	if errors.Is(err, domain.ErrBookClosingNotFound) {
		return ledger{}, errRule(err, "mulai pembukuan dulu di menu Akun")
	}
	if err != nil {
		return ledger{}, err
	}
	today, err := u.today(ctx)
	if err != nil {
		return ledger{}, err
	}
	b, err := u.balancesSince(ctx, closing)
	if err != nil {
		return ledger{}, err
	}
	l := ledger{openFrom: domain.FirstDayAfter(closing.Month), today: today.Format("2006-01-02"), balances: b,
		accounts: map[int64]domain.Account{}, pockets: map[int64]domain.Pocket{}, categories: map[int64]domain.Category{}}

	accounts, err := u.Accounts.FindAll(ctx)
	if err != nil {
		return ledger{}, err
	}
	for _, a := range accounts {
		l.accounts[a.ID] = a
	}
	pockets, err := u.Pockets.FindAll(ctx)
	if err != nil {
		return ledger{}, err
	}
	for _, p := range pockets {
		l.pockets[p.ID] = p
	}
	categories, err := u.Categories.FindAll(ctx)
	if err != nil {
		return ledger{}, err
	}
	for _, c := range categories {
		l.categories[c.ID] = c
	}
	return l, nil
}

// check applies the change (old rows replaced by rows) to the ledger and refuses it when it
// touches a closed month, a date after today, an archived account or kas, a category that does
// not fit, or takes any account or kas below Rp 0
func (l *ledger) check(old, rows []domain.Transaction, pindah bool) error {
	for _, o := range old {
		if o.Date < l.openFrom {
			return errRule(fmt.Errorf("transaction %d is in a closed month", o.ID), "bulan ini sudah tutup buku, perbaiki dengan transaksi koreksi")
		}
	}
	for _, r := range rows {
		if err := l.checkRow(r, old, pindah); err != nil {
			return err
		}
	}

	accounts, pockets := map[int64]int64{}, map[int64]int64{}
	for _, o := range old {
		accounts[o.AccountID] -= o.Signed()
		pockets[o.PocketID] -= o.Signed()
	}
	for _, r := range rows {
		accounts[r.AccountID] += r.Signed()
		pockets[r.PocketID] += r.Signed()
	}
	for id, delta := range accounts {
		if after := l.balances.accounts[id] + delta; after < 0 {
			return errRule(fmt.Errorf("account %d would hold %d", id, after), fmt.Sprintf("saldo akun %s tidak cukup (saldo %s)", l.accounts[id].Name, domain.RupiahForMessage(l.balances.accounts[id])))
		}
	}
	for id, delta := range pockets {
		if after := l.balances.pockets[id] + delta; after < 0 {
			return errRule(fmt.Errorf("pocket %d would hold %d", id, after), fmt.Sprintf("saldo kas %s tidak cukup (saldo %s)", l.pockets[id].Name, domain.RupiahForMessage(l.balances.pockets[id])))
		}
	}
	return nil
}

func (l *ledger) checkRow(r domain.Transaction, old []domain.Transaction, pindah bool) error {
	if r.Date < l.openFrom {
		return errRule(fmt.Errorf("date %s before %s", r.Date, l.openFrom), "tanggal itu sudah tutup buku atau sebelum pembukuan dimulai")
	}
	if r.Date > l.today {
		return invalidInput(fmt.Errorf("date %s after %s", r.Date, l.today), "tanggal tidak boleh setelah hari ini")
	}
	unchanged := func(match func(domain.Transaction) bool) bool {
		for _, o := range old {
			if match(o) {
				return true
			}
		}
		return false
	}

	account, ok := l.accounts[r.AccountID]
	if !ok {
		return apperror.New(apperror.CodeNotFound, domain.ErrAccountNotFound, "akun tidak ditemukan")
	}
	if account.ArchivedAt != nil && !unchanged(func(o domain.Transaction) bool { return o.AccountID == r.AccountID }) {
		return errRule(fmt.Errorf("account %d archived", r.AccountID), "akun "+account.Name+" sudah diarsipkan")
	}
	pocket, ok := l.pockets[r.PocketID]
	if !ok {
		return apperror.New(apperror.CodeNotFound, domain.ErrPocketNotFound, "kas tidak ditemukan")
	}
	if !pocket.IsOpen() {
		return errRule(fmt.Errorf("pocket %d closed", r.PocketID), "kas "+pocket.Name+" sudah diarsipkan")
	}
	category, ok := l.categories[r.CategoryID]
	if !ok {
		return apperror.New(apperror.CodeNotFound, domain.ErrCategoryNotFound, "kategori tidak ditemukan")
	}
	if category.ArchivedAt != nil && !unchanged(func(o domain.Transaction) bool { return o.CategoryID == r.CategoryID }) {
		return errRule(fmt.Errorf("category %d archived", r.CategoryID), "kategori "+category.Name+" sudah diarsipkan")
	}
	if category.Type != r.Kind || category.IsPindah() != pindah || !category.FitsPocket(pocket.Kind) {
		return errRule(fmt.Errorf("category %d does not fit %s into pocket %d", r.CategoryID, r.Kind, r.PocketID), "kategori "+category.Name+" tidak bisa dipakai di kas "+pocket.Name)
	}
	return nil
}

// ListTransactions returns a month's transactions, oldest first
func (u *UseCase) ListTransactions(ctx context.Context, req domain.ListTransactionsRequest) ([]domain.Transaction, error) {
	filter, err := req.Filter()
	if err != nil {
		return nil, err
	}
	return u.Transactions.Find(ctx, filter)
}

// CreateTransaction records a pemasukan or pengeluaran; several lines make a split income
func (u *UseCase) CreateTransaction(ctx context.Context, req domain.SaveTransactionRequest) ([]domain.Transaction, error) {
	req.ID = 0
	if err := req.ValidateAndNormalize(); err != nil {
		return nil, err
	}
	l, err := u.ledger(ctx)
	if err != nil {
		return nil, err
	}
	if err := u.checkCorrection(ctx, l, req.CorrectsID); err != nil {
		return nil, err
	}

	var groupID *string
	if len(req.Lines) > 1 {
		id := newGroupID()
		groupID = &id
	}
	rows := req.Rows(groupID)
	if err := l.check(nil, rows, false); err != nil {
		return nil, err
	}

	saved, err := u.write(ctx, rows)
	if err != nil {
		return nil, err
	}
	u.activityLog.Publish(ctx, req.ToActivityLog(saved[0].ID))
	return saved, nil
}

// UpdateTransaction edits a pemasukan or pengeluaran in an open month, with a reason. A split is
// edited as a whole, its lines in the order they were saved; the kind never changes.
func (u *UseCase) UpdateTransaction(ctx context.Context, req domain.SaveTransactionRequest) ([]domain.Transaction, error) {
	if err := req.ValidateAndNormalize(); err != nil {
		return nil, err
	}
	l, err := u.ledger(ctx)
	if err != nil {
		return nil, err
	}
	old, err := u.group(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if c := l.categories[old[0].CategoryID]; c.IsPindah() {
		return nil, errRule(errors.New("pindah edited as a transaction"), "ubah lewat Pindah dana")
	}
	if old[0].Kind != req.Kind {
		return nil, errRule(errors.New("kind changed"), "jenis transaksi tidak bisa diubah, hapus lalu catat ulang")
	}
	if len(old) != len(req.Lines) {
		return nil, errRule(errors.New("split lines changed"), "jumlah kas pada transaksi bagi tidak bisa diubah, hapus lalu catat ulang")
	}

	rows := make([]domain.Transaction, len(old))
	for i, o := range old {
		line := req.Lines[i]
		o.Amount, o.Date, o.AccountID, o.PocketID, o.CategoryID = line.Amount, req.Date, req.AccountID, line.PocketID, line.CategoryID
		o.Note, o.DonorNote, o.UpdatedBy = req.Note, req.DonorNote, req.UserID
		rows[i] = o
	}
	if err := l.check(old, rows, false); err != nil {
		return nil, err
	}

	saved, err := u.writeEdit(ctx, old, rows, req.Reason, req.UserID)
	if err != nil {
		return nil, err
	}
	u.activityLog.Publish(ctx, req.ToActivityLog(req.ID))
	return saved, nil
}

// MoveFunds is Pindah dana: an out row and an in row, never into or out of zakat or fidyah, out
// of a terikat kas only with a reason and the donors' consent
func (u *UseCase) MoveFunds(ctx context.Context, req domain.MoveFundsRequest) ([]domain.Transaction, error) {
	if err := req.ValidateAndNormalize(); err != nil {
		return nil, err
	}
	l, err := u.ledger(ctx)
	if err != nil {
		return nil, err
	}
	if err := l.checkMove(req); err != nil {
		return nil, err
	}
	outCategory, inCategory, err := u.pindahCategories(ctx)
	if err != nil {
		return nil, err
	}

	var old []domain.Transaction
	groupID := newGroupID()
	if req.GroupID != "" {
		if old, err = u.Transactions.Find(ctx, domain.TransactionFilter{GroupID: req.GroupID}); err != nil {
			return nil, err
		}
		if !isPair(old, l.categories) {
			return nil, apperror.New(apperror.CodeNotFound, domain.ErrTransactionNotFound, "pindah dana tidak ditemukan")
		}
		groupID = req.GroupID
	}

	rows := req.Rows(groupID, outCategory, inCategory)
	if err := l.check(old, rows, true); err != nil {
		return nil, err
	}

	var saved []domain.Transaction
	if len(old) == 0 {
		saved, err = u.write(ctx, rows)
	} else {
		saved, err = u.writeEdit(ctx, pairOrder(old), withIDs(rows, pairOrder(old)), req.Reason, req.UserID)
	}
	if err != nil {
		return nil, err
	}
	u.activityLog.Publish(ctx, req.ToActivityLog(saved[0].ID))
	return saved, nil
}

func (l *ledger) checkMove(req domain.MoveFundsRequest) error {
	from, ok := l.pockets[req.FromPocketID]
	to, ok2 := l.pockets[req.ToPocketID]
	if !ok || !ok2 {
		return apperror.New(apperror.CodeNotFound, domain.ErrPocketNotFound, "kas tidak ditemukan")
	}
	if from.ID == to.ID {
		return nil // setor/tarik tunai: the money stays in its kas
	}
	if !from.Kind.CanMoveFunds() || !to.Kind.CanMoveFunds() {
		return errRule(errors.New("zakat moved"), "dana zakat dan fidyah tidak bisa dipindah ke atau dari kas lain")
	}
	if from.Kind.NeedsDonorConsent() && (!req.DonorConsent || req.Note == "") {
		return errRule(errors.New("terikat moved without consent"), "dana kas terikat hanya bisa dipindah dengan alasan dan persetujuan donatur")
	}
	return nil
}

// DeleteTransaction soft-deletes a transaction and the rest of its group, with a reason
func (u *UseCase) DeleteTransaction(ctx context.Context, req domain.DeleteTransactionRequest) error {
	if err := req.ValidateAndNormalize(); err != nil {
		return err
	}
	l, err := u.ledger(ctx)
	if err != nil {
		return err
	}
	old, err := u.group(ctx, req.ID)
	if err != nil {
		return err
	}
	if err := l.check(old, nil, false); err != nil {
		return err
	}

	err = u.Transactor.WithTx(ctx, func(ctx context.Context) error {
		for _, o := range old {
			if err := u.Transactions.AddRevision(ctx, domain.TransactionRevision{TransactionID: o.ID, Action: domain.RevisionDelete, Previous: o, Reason: req.Reason, UserID: req.UserID}); err != nil {
				return err
			}
			if err := u.Transactions.SoftDelete(ctx, o.ID, req.UserID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	u.activityLog.Publish(ctx, domain.TransactionDeletedActivityLog(old))
	return nil
}

// ListRevisions returns what a transaction looked like before each edit or delete
func (u *UseCase) ListRevisions(ctx context.Context, id int64) ([]domain.TransactionRevision, error) {
	if _, err := u.Transactions.FindByID(ctx, id); err != nil {
		return nil, err
	}
	return u.Transactions.FindRevisions(ctx, id)
}

// SaveReceipt attaches or replaces the receipt photo of a transaction
func (u *UseCase) SaveReceipt(ctx context.Context, id int64, data []byte) error {
	photo, err := domain.ReceiptPhoto(data)
	if err != nil {
		return err
	}
	tx, err := u.Transactions.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if tx.DeletedAt != nil {
		return errRule(errors.New("deleted"), "transaksi sudah dihapus")
	}
	if err := u.Transactions.SavePhoto(ctx, id, photo); err != nil {
		return err
	}
	u.activityLog.Publish(ctx, domain.ActivityLog{Action: domain.ActivityReceiptUploaded, EntityID: id, Summary: "mengunggah foto bukti transaksi"})
	return nil
}

func (u *UseCase) GetReceipt(ctx context.Context, id int64) (domain.TransactionPhoto, error) {
	return u.Transactions.GetPhoto(ctx, id)
}

// group returns a live transaction with the rest of its group, in saved order
func (u *UseCase) group(ctx context.Context, id int64) ([]domain.Transaction, error) {
	tx, err := u.Transactions.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if tx.DeletedAt != nil {
		return nil, apperror.New(apperror.CodeNotFound, domain.ErrTransactionNotFound, "transaksi sudah dihapus")
	}
	if tx.GroupID == nil {
		return []domain.Transaction{tx}, nil
	}
	return u.Transactions.Find(ctx, domain.TransactionFilter{GroupID: *tx.GroupID})
}

func (u *UseCase) checkCorrection(ctx context.Context, l ledger, correctsID *int64) error {
	if correctsID == nil {
		return nil
	}
	original, err := u.Transactions.FindByID(ctx, *correctsID)
	if err != nil {
		return err
	}
	if original.Date >= l.openFrom {
		return errRule(errors.New("correcting an open month"), "transaksi di bulan yang belum ditutup cukup diubah, tidak perlu koreksi")
	}
	return nil
}

func (u *UseCase) pindahCategories(ctx context.Context) (int64, int64, error) {
	out, err := u.Categories.FindByCode(ctx, domain.CategoryCodePindahOut)
	if err != nil {
		return 0, 0, err
	}
	in, err := u.Categories.FindByCode(ctx, domain.CategoryCodePindahIn)
	if err != nil {
		return 0, 0, err
	}
	return out.ID, in.ID, nil
}

// write creates rows in one database transaction
func (u *UseCase) write(ctx context.Context, rows []domain.Transaction) ([]domain.Transaction, error) {
	saved := make([]domain.Transaction, 0, len(rows))
	err := u.Transactor.WithTx(ctx, func(ctx context.Context) error {
		for _, r := range rows {
			s, err := u.Transactions.Create(ctx, r)
			if err != nil {
				return err
			}
			saved = append(saved, s)
		}
		return nil
	})
	return saved, err
}

// writeEdit keeps each old row as a revision, then saves its new values
func (u *UseCase) writeEdit(ctx context.Context, old, rows []domain.Transaction, reason string, userID int64) ([]domain.Transaction, error) {
	saved := make([]domain.Transaction, 0, len(rows))
	err := u.Transactor.WithTx(ctx, func(ctx context.Context) error {
		for i, r := range rows {
			if err := u.Transactions.AddRevision(ctx, domain.TransactionRevision{TransactionID: old[i].ID, Action: domain.RevisionEdit, Previous: old[i], Reason: reason, UserID: userID}); err != nil {
				return err
			}
			r.ID, r.UpdatedBy = old[i].ID, userID
			s, err := u.Transactions.Update(ctx, r)
			if err != nil {
				return err
			}
			saved = append(saved, s)
		}
		return nil
	})
	return saved, err
}

// pairOrder puts a pindah pair as out, in
func pairOrder(rows []domain.Transaction) []domain.Transaction {
	if rows[0].Kind == domain.CategoryTypeOut {
		return rows
	}
	return []domain.Transaction{rows[1], rows[0]}
}

func withIDs(rows, old []domain.Transaction) []domain.Transaction {
	for i := range rows {
		rows[i].ID, rows[i].CreatedBy = old[i].ID, old[i].CreatedBy
	}
	return rows
}

func newGroupID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func invalidInput(err error, message string) *apperror.AppError {
	return apperror.New(apperror.CodeInvalidInput, err, message)
}

// isPair is a live pindah dana pair
func isPair(rows []domain.Transaction, categories map[int64]domain.Category) bool {
	if len(rows) != 2 {
		return false
	}
	c := categories[rows[0].CategoryID]
	return c.IsPindah()
}
