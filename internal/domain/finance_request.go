package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

// Requests and responses of the Kas, Akun, Kategori and Mulai pembukuan endpoints (E7-13, E7-09,
// E7-01). Shape checks live here; rules that need stored data (a kas's kind, a balance) live in
// the finance usecase.

const (
	MaxFinanceNameLength   = 60
	MaxPocketPurposeLength = 200
	MaxAccountNumberLength = 40

	monthLayout = "2006-01"
)

// AccountWithBalance is an account with its balance computed from the last closing and the
// transactions after it
type AccountWithBalance struct {
	Account
	Balance int64 `json:"balance"`
}

// PocketWithBalance is a kas with its balance
type PocketWithBalance struct {
	Pocket
	Balance int64 `json:"balance"`
}

// CategoryWithCount is a category with how many transactions (not deleted) use it
type CategoryWithCount struct {
	Category
	TransactionCount int `json:"transaction_count"`
}

// CanArchive is a system category that may be hidden: an asnaf or a fidyah recipient. Income
// system categories and Pindah dana carry the rules, so they always stay.
func (c *Category) CanArchive() bool {
	return !c.IsSystem() || (c.Type == CategoryTypeOut && !c.IsPindah())
}

type CreateAccountRequest struct {
	CreatedBy int64 `json:"-"`

	Name   string      `json:"name"`
	Kind   AccountKind `json:"kind"`
	Number string      `json:"number"` // bank only; ignored for cash
}

func (r *CreateAccountRequest) ValidateAndNormalize() error {
	if !r.Kind.IsValid() {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("unknown account kind %q", r.Kind), "jenis akun harus tunai atau bank")
	}
	var err error
	if r.Name, err = validateFinanceName(r.Name, "nama akun"); err != nil {
		return err
	}
	r.Number, err = normalizeAccountNumber(r.Kind, r.Number)
	return err
}

func (r *CreateAccountRequest) ToActivityLog(id int64) ActivityLog {
	return ActivityLog{Action: ActivityAccountCreated, EntityID: id, Summary: "menambah akun " + quoteForSummary(r.Name)}
}

type UpdateAccountRequest struct {
	ID int64 `json:"-"`

	Name   string `json:"name"`
	Number string `json:"number"`
}

// ValidateAndNormalize checks the request against the account's kind, which never changes
func (r *UpdateAccountRequest) ValidateAndNormalize(kind AccountKind) error {
	var err error
	if r.Name, err = validateFinanceName(r.Name, "nama akun"); err != nil {
		return err
	}
	r.Number, err = normalizeAccountNumber(kind, r.Number)
	return err
}

func (r *UpdateAccountRequest) ToActivityLog() ActivityLog {
	return ActivityLog{Action: ActivityAccountUpdated, EntityID: r.ID, Summary: "mengubah akun " + quoteForSummary(r.Name)}
}

// AccountArchiveActivityLog is the log of archiving (archived true) or restoring an account
func AccountArchiveActivityLog(account Account, archived bool) ActivityLog {
	if archived {
		return ActivityLog{Action: ActivityAccountArchived, EntityID: account.ID, Summary: "mengarsipkan akun " + quoteForSummary(account.Name)}
	}
	return ActivityLog{Action: ActivityAccountRestored, EntityID: account.ID, Summary: "memulihkan akun " + quoteForSummary(account.Name)}
}

// TerikatPocketFields are what pengurus set on a terikat kas, on create and on edit
type TerikatPocketFields struct {
	Name         string      `json:"name"`
	Purpose      string      `json:"purpose"`
	TargetAmount *int64      `json:"target_amount"` // nil or 0 = no target
	SurplusPlan  SurplusPlan `json:"surplus_plan"`
	ShowOnTV     bool        `json:"show_on_tv"`
}

func (f *TerikatPocketFields) ValidateAndNormalize() error {
	var err error
	if f.Name, err = validateFinanceName(f.Name, "nama kas"); err != nil {
		return err
	}
	f.Purpose = strings.TrimSpace(f.Purpose)
	if f.Purpose == "" {
		return apperror.New(apperror.CodeInvalidInput, errors.New("empty purpose"), "peruntukan wajib diisi")
	}
	if utf8.RuneCountInString(f.Purpose) > MaxPocketPurposeLength {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("purpose longer than %d characters", MaxPocketPurposeLength), fmt.Sprintf("peruntukan maksimal %d karakter", MaxPocketPurposeLength))
	}
	if f.TargetAmount != nil && *f.TargetAmount == 0 {
		f.TargetAmount = nil
	}
	if f.TargetAmount != nil && (*f.TargetAmount < 0 || *f.TargetAmount > MaxTransactionAmount) {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("target %d out of range", *f.TargetAmount), "target harus antara Rp 1 dan Rp 10.000.000.000")
	}
	if !f.SurplusPlan.IsValid() {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("unknown surplus plan %q", f.SurplusPlan), "pilih rencana sisa dana")
	}
	return nil
}

// apply copies the fields onto a terikat pocket
func (f *TerikatPocketFields) apply(p *Pocket) {
	plan := f.SurplusPlan
	p.Name, p.Purpose, p.TargetAmount, p.SurplusPlan, p.ShowOnTV = f.Name, f.Purpose, f.TargetAmount, &plan, f.ShowOnTV
}

// CreatePocketRequest adds a terikat kas; the built-in ones are seeded
type CreatePocketRequest struct {
	CreatedBy int64 `json:"-"`

	TerikatPocketFields
}

func (r *CreatePocketRequest) ToPocket() Pocket {
	p := Pocket{Kind: PocketKindTerikat, Status: PocketStatusActive, CreatedBy: r.CreatedBy}
	r.apply(&p)
	return p
}

func (r *CreatePocketRequest) ToActivityLog(id int64) ActivityLog {
	return ActivityLog{Action: ActivityPocketCreated, EntityID: id, Summary: "menambah kas " + quoteForSummary(r.Name)}
}

// UpdatePocketRequest edits a kas. A built-in kas is locked except for ShowOnTV; its other
// fields are ignored.
type UpdatePocketRequest struct {
	ID int64 `json:"-"`

	TerikatPocketFields
}

// Apply returns the pocket as saved after the request: a built-in kas only changes ShowOnTV
func (r *UpdatePocketRequest) Apply(p Pocket) (Pocket, error) {
	if p.Kind.IsBuiltIn() {
		p.ShowOnTV = r.ShowOnTV
		return p, nil
	}
	if err := r.ValidateAndNormalize(); err != nil {
		return Pocket{}, err
	}
	r.apply(&p)
	return p, nil
}

func PocketUpdatedActivityLog(p Pocket) ActivityLog {
	return ActivityLog{Action: ActivityPocketUpdated, EntityID: p.ID, Summary: "mengubah kas " + quoteForSummary(p.Name)}
}

// PocketArchiveActivityLog is the log of archiving (archived true) or restoring a kas
func PocketArchiveActivityLog(p Pocket, archived bool) ActivityLog {
	if archived {
		return ActivityLog{Action: ActivityPocketArchived, EntityID: p.ID, Summary: "mengarsipkan kas " + quoteForSummary(p.Name)}
	}
	return ActivityLog{Action: ActivityPocketRestored, EntityID: p.ID, Summary: "memulihkan kas " + quoteForSummary(p.Name)}
}

// CreateCategoryRequest adds an own category, which is always for Kas Umum and terikat kas
type CreateCategoryRequest struct {
	CreatedBy int64 `json:"-"`

	Name string       `json:"name"`
	Type CategoryType `json:"type"`
}

func (r *CreateCategoryRequest) ValidateAndNormalize() error {
	if !r.Type.IsValid() {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("unknown category type %q", r.Type), "jenis kategori harus pemasukan atau pengeluaran")
	}
	var err error
	r.Name, err = validateFinanceName(r.Name, "nama kategori")
	return err
}

func (r *CreateCategoryRequest) ToCategory() Category {
	return Category{Name: r.Name, Type: r.Type, Scope: CategoryScopeUmumTerikat, CreatedBy: r.CreatedBy}
}

func (r *CreateCategoryRequest) ToActivityLog(id int64) ActivityLog {
	return ActivityLog{Action: ActivityCategoryCreated, EntityID: id, Summary: "menambah kategori " + quoteForSummary(r.Name)}
}

// UpdateCategoryRequest renames an own category; the type is fixed and system ones keep their name
type UpdateCategoryRequest struct {
	ID int64 `json:"-"`

	Name string `json:"name"`
}

func (r *UpdateCategoryRequest) ValidateAndNormalize() error {
	var err error
	r.Name, err = validateFinanceName(r.Name, "nama kategori")
	return err
}

func (r *UpdateCategoryRequest) ToActivityLog() ActivityLog {
	return ActivityLog{Action: ActivityCategoryUpdated, EntityID: r.ID, Summary: "mengubah kategori " + quoteForSummary(r.Name)}
}

func CategoryDeletedActivityLog(c Category) ActivityLog {
	return ActivityLog{Action: ActivityCategoryDeleted, EntityID: c.ID, Summary: "menghapus kategori " + quoteForSummary(c.Name)}
}

// CategoryArchiveActivityLog is the log of archiving (archived true) or restoring a category
func CategoryArchiveActivityLog(c Category, archived bool) ActivityLog {
	if archived {
		return ActivityLog{Action: ActivityCategoryArchived, EntityID: c.ID, Summary: "mengarsipkan kategori " + quoteForSummary(c.Name)}
	}
	return ActivityLog{Action: ActivityCategoryRestored, EntityID: c.ID, Summary: "memulihkan kategori " + quoteForSummary(c.Name)}
}

// OpeningBalance is what one account or kas holds when pembukuan starts
type OpeningBalance struct {
	ID     int64 `json:"id"` // account or pocket ID
	Amount int64 `json:"amount"`
}

// StartBookkeepingRequest is "Mulai pembukuan": the first month recorded and what each account
// and each kas holds at its start. Accounts and kas left out start at Rp 0. Both lists must add
// up to the same total, since every rupiah sits in one account and belongs to one kas.
type StartBookkeepingRequest struct {
	ClosedBy int64 `json:"-"`

	StartMonth string           `json:"start_month"` // YYYY-MM
	Accounts   []OpeningBalance `json:"accounts"`
	Pockets    []OpeningBalance `json:"pockets"`
}

// ValidateAndNormalize checks the shape; currentMonth (YYYY-MM, the mosque's clock) is the
// latest month pembukuan may start
func (r *StartBookkeepingRequest) ValidateAndNormalize(currentMonth string) error {
	r.StartMonth = strings.TrimSpace(r.StartMonth)
	if _, err := time.Parse(monthLayout, r.StartMonth); err != nil {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("invalid start month %q", r.StartMonth), "bulan mulai tidak valid")
	}
	if r.StartMonth > currentMonth {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("start month %s after %s", r.StartMonth, currentMonth), "bulan mulai tidak boleh setelah bulan ini")
	}
	accountTotal, err := openingTotal(r.Accounts, "akun")
	if err != nil {
		return err
	}
	pocketTotal, err := openingTotal(r.Pockets, "kas")
	if err != nil {
		return err
	}
	if accountTotal != pocketTotal {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("accounts total %d, pockets total %d", accountTotal, pocketTotal), "total saldo akun harus sama dengan total saldo kas")
	}
	return nil
}

// OpeningClosing is the closing the opening is saved as: the month before StartMonth, holding
// every account and kas given, those not in the request at Rp 0
func (r *StartBookkeepingRequest) OpeningClosing(accountIDs, pocketIDs []int64) BookClosing {
	start, _ := time.Parse(monthLayout, r.StartMonth)
	closing := BookClosing{Month: start.AddDate(0, -1, 0).Format(monthLayout), Kind: BookClosingOpening, ClosedBy: r.ClosedBy}

	accounts := amounts(r.Accounts)
	for _, id := range accountIDs {
		closing.Accounts = append(closing.Accounts, BookClosingAccount{AccountID: id, ClosingBalance: accounts[id]})
		closing.ClosingBalance += accounts[id]
	}
	pockets := amounts(r.Pockets)
	for _, id := range pocketIDs {
		closing.Pockets = append(closing.Pockets, BookClosingPocket{PocketID: id, ClosingBalance: pockets[id]})
	}
	return closing
}

func (r *StartBookkeepingRequest) ToActivityLog(id int64) ActivityLog {
	return ActivityLog{Action: ActivityBookkeepingStarted, EntityID: id, Summary: "memulai pembukuan dari " + monthForSummary(r.StartMonth)}
}

// Bookkeeping is where pembukuan stands: not started yet, or started with the opening balances
type Bookkeeping struct {
	Started    bool         `json:"started"`
	StartMonth string       `json:"start_month"` // YYYY-MM; "" before it starts
	Opening    *BookClosing `json:"opening"`
}

// NewBookkeeping describes pembukuan from its opening closing
func NewBookkeeping(opening BookClosing) Bookkeeping {
	return Bookkeeping{Started: true, StartMonth: MonthAfter(opening.Month), Opening: &opening}
}

// MonthAfter is the month after month (YYYY-MM); an invalid month is returned as is
func MonthAfter(month string) string {
	t, err := time.Parse(monthLayout, month)
	if err != nil {
		return month
	}
	return t.AddDate(0, 1, 0).Format(monthLayout)
}

// FirstDayAfter is the first date (YYYY-MM-DD) after month (YYYY-MM)
func FirstDayAfter(month string) string {
	return MonthAfter(month) + "-01"
}

func openingTotal(balances []OpeningBalance, what string) (int64, error) {
	seen := make(map[int64]bool, len(balances))
	var total int64
	for _, b := range balances {
		if b.ID <= 0 || seen[b.ID] {
			return 0, apperror.New(apperror.CodeInvalidInput, fmt.Errorf("%s id %d invalid or repeated", what, b.ID), what+" tidak valid atau terulang")
		}
		seen[b.ID] = true
		if b.Amount < 0 || b.Amount > MaxTransactionAmount {
			return 0, apperror.New(apperror.CodeInvalidInput, fmt.Errorf("%s %d amount %d out of range", what, b.ID, b.Amount), "saldo awal harus antara Rp 0 dan Rp 10.000.000.000")
		}
		total += b.Amount
	}
	return total, nil
}

func amounts(balances []OpeningBalance) map[int64]int64 {
	m := make(map[int64]int64, len(balances))
	for _, b := range balances {
		m[b.ID] = b.Amount
	}
	return m
}

func validateFinanceName(s, field string) (string, error) {
	trimmed := strings.Join(strings.Fields(s), " ")
	if trimmed == "" {
		return "", apperror.New(apperror.CodeInvalidInput, fmt.Errorf("empty %s", field), field+" wajib diisi")
	}
	if utf8.RuneCountInString(trimmed) > MaxFinanceNameLength {
		return "", apperror.New(apperror.CodeInvalidInput, fmt.Errorf("%s longer than %d characters", field, MaxFinanceNameLength), fmt.Sprintf("%s maksimal %d karakter", field, MaxFinanceNameLength))
	}
	return trimmed, nil
}

func normalizeAccountNumber(kind AccountKind, number string) (string, error) {
	if kind == AccountKindCash {
		return "", nil
	}
	number = strings.TrimSpace(number)
	if utf8.RuneCountInString(number) > MaxAccountNumberLength {
		return "", apperror.New(apperror.CodeInvalidInput, fmt.Errorf("account number longer than %d characters", MaxAccountNumberLength), fmt.Sprintf("nomor rekening maksimal %d karakter", MaxAccountNumberLength))
	}
	return number, nil
}

// monthForSummary turns YYYY-MM into "Nov 2026"; anything else is returned as is
func monthForSummary(month string) string {
	t, err := time.Parse(monthLayout, month)
	if err != nil {
		return month
	}
	return indonesianMonths[t.Month()-1] + t.Format(" 2006")
}
