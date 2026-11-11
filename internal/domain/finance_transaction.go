package domain

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

// Requests of the transaction endpoints (E7-02, E7-15, E7-14, E7-05). Shape checks live here;
// the rules that need stored data (kas kinds, balances, closed months) live in the finance
// usecase.

const (
	MaxTransactionNoteLength = 200
	MaxReasonLength          = 200
	// MaxReceiptPhotoSize is the largest receipt photo accepted; the browser resizes to fit
	MaxReceiptPhotoSize = 1 << 20
)

var receiptPhotoTypes = map[string]bool{"image/png": true, "image/jpeg": true, "image/webp": true}

// TransactionLine is one kas's share of a transaction: one line for a plain transaction,
// several for a split income ("Bagi ke beberapa kas")
type TransactionLine struct {
	PocketID   int64 `json:"pocket_id"`
	CategoryID int64 `json:"category_id"`
	Amount     int64 `json:"amount"`
}

// SaveTransactionRequest records a pemasukan or pengeluaran, or edits one (ID set). Editing a
// split edits the whole group, line by line in the same order.
type SaveTransactionRequest struct {
	UserID int64 `json:"-"`
	ID     int64 `json:"-"` // 0 creates

	Kind       CategoryType      `json:"kind"`
	Date       string            `json:"date"` // YYYY-MM-DD
	AccountID  int64             `json:"account_id"`
	Lines      []TransactionLine `json:"lines"`
	Note       string            `json:"note"`
	DonorNote  string            `json:"donor_note"`
	CorrectsID *int64            `json:"corrects_id"` // a transaction in a closed month this one corrects
	Reason     string            `json:"reason"`      // required when editing
}

func (r *SaveTransactionRequest) ValidateAndNormalize() error {
	if !r.Kind.IsValid() {
		return invalidInput(fmt.Errorf("unknown kind %q", r.Kind), "jenis transaksi harus pemasukan atau pengeluaran")
	}
	if err := validateDate(r.Date); err != nil {
		return err
	}
	if r.AccountID <= 0 {
		return invalidInput(errors.New("no account"), "pilih akun")
	}
	if len(r.Lines) == 0 {
		return invalidInput(errors.New("no lines"), "pilih kas dan nominal")
	}
	if len(r.Lines) > 1 && r.Kind != CategoryTypeIn {
		return invalidInput(errors.New("split expense"), "hanya pemasukan yang bisa dibagi ke beberapa kas")
	}
	seen := map[int64]bool{}
	for _, l := range r.Lines {
		if l.PocketID <= 0 || l.CategoryID <= 0 {
			return invalidInput(errors.New("line without kas or category"), "pilih kas dan kategori")
		}
		if seen[l.PocketID] {
			return invalidInput(fmt.Errorf("pocket %d repeated", l.PocketID), "satu kas hanya boleh muncul sekali")
		}
		seen[l.PocketID] = true
		if err := validateAmount(l.Amount); err != nil {
			return err
		}
	}
	var err error
	if r.Note, err = validateText(r.Note, MaxTransactionNoteLength, "keterangan"); err != nil {
		return err
	}
	if r.DonorNote, err = validateText(r.DonorNote, MaxTransactionNoteLength, "catatan donatur"); err != nil {
		return err
	}
	return r.validateReason()
}

func (r *SaveTransactionRequest) validateReason() error {
	var err error
	if r.Reason, err = validateText(r.Reason, MaxReasonLength, "alasan"); err != nil {
		return err
	}
	if r.ID != 0 && r.Reason == "" {
		return invalidInput(errors.New("edit without reason"), "alasan perubahan wajib diisi")
	}
	return nil
}

// Total is the amount of every line together
func (r *SaveTransactionRequest) Total() int64 {
	var total int64
	for _, l := range r.Lines {
		total += l.Amount
	}
	return total
}

// Rows are the transactions the request saves, one per line, sharing groupID when split
func (r *SaveTransactionRequest) Rows(groupID *string) []Transaction {
	rows := make([]Transaction, 0, len(r.Lines))
	for _, l := range r.Lines {
		rows = append(rows, Transaction{
			Kind: r.Kind, Amount: l.Amount, Date: r.Date, AccountID: r.AccountID, PocketID: l.PocketID,
			CategoryID: l.CategoryID, GroupID: groupID, CorrectsID: r.CorrectsID, Note: r.Note,
			DonorNote: r.DonorNote, CreatedBy: r.UserID, UpdatedBy: r.UserID,
		})
	}
	return rows
}

func (r *SaveTransactionRequest) ToActivityLog(id int64) ActivityLog {
	verb, what := "mencatat", "pemasukan"
	if r.ID != 0 {
		verb = "mengubah"
	}
	if r.Kind == CategoryTypeOut {
		what = "pengeluaran"
	}
	return ActivityLog{Action: transactionAction(r.ID), EntityID: id, Summary: fmt.Sprintf("%s %s %s", verb, what, rupiahForSummary(r.Total()))}
}

func transactionAction(id int64) ActivityAction {
	if id != 0 {
		return ActivityTransactionUpdated
	}
	return ActivityTransactionCreated
}

// MoveFundsRequest is Pindah dana: an out row on the source and an in row on the destination.
// The same kas on both sides with two accounts is setor/tarik tunai.
type MoveFundsRequest struct {
	UserID  int64  `json:"-"`
	GroupID string `json:"-"` // set when editing a pair

	FromPocketID  int64  `json:"from_pocket_id"`
	FromAccountID int64  `json:"from_account_id"`
	ToPocketID    int64  `json:"to_pocket_id"`
	ToAccountID   int64  `json:"to_account_id"`
	Amount        int64  `json:"amount"`
	Date          string `json:"date"`
	Note          string `json:"note"` // the reason it moves, shown on both rows
	DonorConsent  bool   `json:"donor_consent"`
	Reason        string `json:"reason"` // required when editing
}

func (r *MoveFundsRequest) ValidateAndNormalize() error {
	if r.FromPocketID <= 0 || r.ToPocketID <= 0 || r.FromAccountID <= 0 || r.ToAccountID <= 0 {
		return invalidInput(errors.New("missing side"), "pilih kas dan akun asal serta tujuan")
	}
	if r.FromPocketID == r.ToPocketID && r.FromAccountID == r.ToAccountID {
		return invalidInput(errors.New("same side"), "asal dan tujuan tidak boleh sama")
	}
	if err := validateAmount(r.Amount); err != nil {
		return err
	}
	if err := validateDate(r.Date); err != nil {
		return err
	}
	var err error
	if r.Note, err = validateText(r.Note, MaxTransactionNoteLength, "alasan"); err != nil {
		return err
	}
	if r.Reason, err = validateText(r.Reason, MaxReasonLength, "alasan perubahan"); err != nil {
		return err
	}
	if r.GroupID != "" && r.Reason == "" {
		return invalidInput(errors.New("edit without reason"), "alasan perubahan wajib diisi")
	}
	return nil
}

// Rows are the out and in rows of the pair, given the two Pindah dana categories
func (r *MoveFundsRequest) Rows(groupID string, outCategory, inCategory int64) []Transaction {
	g := groupID
	base := Transaction{Amount: r.Amount, Date: r.Date, GroupID: &g, DonorConsent: r.DonorConsent, Note: r.Note, CreatedBy: r.UserID, UpdatedBy: r.UserID}
	out, in := base, base
	out.Kind, out.AccountID, out.PocketID, out.CategoryID = CategoryTypeOut, r.FromAccountID, r.FromPocketID, outCategory
	in.Kind, in.AccountID, in.PocketID, in.CategoryID = CategoryTypeIn, r.ToAccountID, r.ToPocketID, inCategory
	return []Transaction{out, in}
}

func (r *MoveFundsRequest) ToActivityLog(id int64) ActivityLog {
	verb, action := "memindahkan dana", ActivityFundsMoved
	if r.GroupID != "" {
		verb, action = "mengubah pindah dana", ActivityTransactionUpdated
	}
	return ActivityLog{Action: action, EntityID: id, Summary: verb + " " + rupiahForSummary(r.Amount)}
}

// DeleteTransactionRequest soft-deletes a transaction and every row of its group
type DeleteTransactionRequest struct {
	UserID int64 `json:"-"`
	ID     int64 `json:"-"`

	Reason string `json:"reason"`
}

func (r *DeleteTransactionRequest) ValidateAndNormalize() error {
	var err error
	if r.Reason, err = validateText(r.Reason, MaxReasonLength, "alasan"); err != nil {
		return err
	}
	if r.Reason == "" {
		return invalidInput(errors.New("delete without reason"), "alasan penghapusan wajib diisi")
	}
	return nil
}

func TransactionDeletedActivityLog(rows []Transaction) ActivityLog {
	var total int64
	for _, r := range rows {
		if r.Kind == rows[0].Kind {
			total += r.Amount
		}
	}
	return ActivityLog{Action: ActivityTransactionDeleted, EntityID: rows[0].ID, Summary: "menghapus transaksi " + rupiahForSummary(total)}
}

// ListTransactionsRequest is GET /finance/transactions: a month, optionally one kas
type ListTransactionsRequest struct {
	Month          string `query:"month"` // YYYY-MM
	PocketID       int64  `query:"pocket_id"`
	IncludeDeleted bool   `query:"include_deleted"`
}

func (r *ListTransactionsRequest) Filter() (TransactionFilter, error) {
	start, err := time.Parse(monthLayout, r.Month)
	if err != nil {
		return TransactionFilter{}, invalidInput(fmt.Errorf("invalid month %q", r.Month), "bulan tidak valid")
	}
	end := start.AddDate(0, 1, -1)
	return TransactionFilter{From: start.Format(dateLayout), To: end.Format(dateLayout), PocketID: r.PocketID, IncludeDeleted: r.IncludeDeleted}, nil
}

// ReceiptPhoto validates an uploaded receipt
func ReceiptPhoto(data []byte) (TransactionPhoto, error) {
	if len(data) == 0 {
		return TransactionPhoto{}, invalidInput(errors.New("empty photo"), "foto bukti wajib diunggah")
	}
	if len(data) > MaxReceiptPhotoSize {
		return TransactionPhoto{}, invalidInput(fmt.Errorf("photo is %d bytes", len(data)), "ukuran foto maksimal 1 MB")
	}
	contentType := http.DetectContentType(data)
	if !receiptPhotoTypes[contentType] {
		return TransactionPhoto{}, invalidInput(fmt.Errorf("photo type %q", contentType), "format foto harus PNG, JPG, atau WebP")
	}
	return TransactionPhoto{ContentType: contentType, Data: data}, nil
}

func validateAmount(amount int64) error {
	if amount <= 0 || amount > MaxTransactionAmount {
		return invalidInput(fmt.Errorf("amount %d out of range", amount), "nominal harus antara Rp 1 dan Rp 10.000.000.000")
	}
	return nil
}

func validateDate(date string) error {
	if _, err := time.Parse(dateLayout, date); err != nil {
		return invalidInput(fmt.Errorf("invalid date %q", date), "tanggal tidak valid")
	}
	return nil
}

func validateText(s string, max int, field string) (string, error) {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) > max {
		return "", invalidInput(fmt.Errorf("%s longer than %d", field, max), fmt.Sprintf("%s maksimal %d karakter", field, max))
	}
	return s, nil
}

func invalidInput(err error, message string) *apperror.AppError {
	return apperror.New(apperror.CodeInvalidInput, err, message)
}

// rupiahForSummary turns 1500000 into "Rp 1.500.000"
func rupiahForSummary(amount int64) string {
	digits := fmt.Sprint(amount)
	var b strings.Builder
	for i, d := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(d)
	}
	return "Rp " + b.String()
}

// RupiahForMessage is rupiahForSummary for usecase messages
func RupiahForMessage(amount int64) string {
	if amount < 0 {
		return "-" + rupiahForSummary(-amount)
	}
	return rupiahForSummary(amount)
}
