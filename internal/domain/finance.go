package domain

import (
	"errors"
	"time"
)

// Keuangan (E7): single-entry, cash-basis bookkeeping on two axes. An Account says where the
// money is (Tunai, a bank); a Pocket ("kas" in the UI) says whose money it is and what it may be
// spent on. Every Transaction has both, so every account balance and every pocket balance can be
// computed, and the two always add up to the same total.
//
// Rules live here, per pocket kind, not in the database: zakat and fidyah are closed both ways,
// earmarked (terikat) money moves only with the donor's consent, and nothing goes below Rp 0.
// Balances are never stored; the usecase sums the transactions after the last BookClosing.

var (
	ErrAccountNotFound     = errors.New("account not found")
	ErrPocketNotFound      = errors.New("pocket not found")
	ErrCategoryNotFound    = errors.New("category not found")
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrBookClosingNotFound = errors.New("book closing not found")
)

const (
	// MaxTransactionAmount is the largest amount one transaction may carry, in Rupiah
	MaxTransactionAmount int64 = 10_000_000_000
	// ConfirmAmountAbove is where the form asks to confirm the amount (a mistyped zero costs a lot)
	ConfirmAmountAbove int64 = 10_000_000
)

// AccountKind is what an account physically is
type AccountKind string

const (
	AccountKindCash AccountKind = "cash" // the box (Tunai)
	AccountKindBank AccountKind = "bank"
)

func (k AccountKind) IsValid() bool {
	return k == AccountKindCash || k == AccountKindBank
}

// Account is where money physically sits
type Account struct {
	ID         int64       `json:"id"`
	Name       string      `json:"name"`
	Kind       AccountKind `json:"kind"`
	Number     string      `json:"number"` // bank account number; "" for cash
	ArchivedAt *time.Time  `json:"archived_at"`
	CreatedBy  int64       `json:"created_by"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// PocketKind decides what a pocket's money may be used for. The four built-in kinds are seeded
// once and carry fixed rules; users only create terikat pockets.
type PocketKind string

const (
	PocketKindUmum        PocketKind = "umum"         // Kas Umum: infaq without a stated purpose
	PocketKindZakatMal    PocketKind = "zakat_mal"    // only from muzakki, only to the 8 asnaf
	PocketKindZakatFitrah PocketKind = "zakat_fitrah" // like zakat mal; paid out before Idul Fitri
	PocketKindFidyah      PocketKind = "fidyah"       // only to fakir and miskin
	PocketKindTerikat     PocketKind = "terikat"      // earmarked by donors for one purpose
)

func (k PocketKind) IsValid() bool {
	_, ok := pocketCategoryScopes[k]
	return ok
}

// IsBuiltIn is a kind seeded once and never added, renamed or deleted by users
func (k PocketKind) IsBuiltIn() bool {
	return k.IsValid() && k != PocketKindTerikat
}

// IsZakat is a zakat or fidyah pocket: its money is closed both ways (QS At-Taubah 60)
func (k PocketKind) IsZakat() bool {
	return k == PocketKindZakatMal || k == PocketKindZakatFitrah || k == PocketKindFidyah
}

// CanMoveFunds is whether Pindah dana may take money out of, or bring money into, this kind of
// pocket from another pocket. Moving between accounts within one pocket (setor tunai) is always
// allowed.
func (k PocketKind) CanMoveFunds() bool {
	return k == PocketKindUmum || k == PocketKindTerikat
}

// NeedsDonorConsent is whether moving money out of this kind to another pocket needs the donors'
// consent: earmarked infaq may only be redirected when they agree
func (k PocketKind) NeedsDonorConsent() bool {
	return k == PocketKindTerikat
}

// CategoryScope is which categories this kind of pocket uses
func (k PocketKind) CategoryScope() CategoryScope {
	return pocketCategoryScopes[k]
}

var pocketCategoryScopes = map[PocketKind]CategoryScope{
	PocketKindUmum:        CategoryScopeUmumTerikat,
	PocketKindTerikat:     CategoryScopeUmumTerikat,
	PocketKindZakatMal:    CategoryScopeZakat,
	PocketKindZakatFitrah: CategoryScopeZakat,
	PocketKindFidyah:      CategoryScopeFidyah,
}

// PocketStatus is where a pocket is in its life
type PocketStatus string

const (
	PocketStatusActive        PocketStatus = "active"
	PocketStatusTargetReached PocketStatus = "target_reached"
	PocketStatusDone          PocketStatus = "done"
	PocketStatusArchived      PocketStatus = "archived"
)

// SurplusPlan is what happens to a terikat pocket's remainder once its program is done, stated
// when the pocket is created so donors agree before they give
type SurplusPlan string

const (
	SurplusPlanToUmum   SurplusPlan = "to_umum"   // "Dialihkan ke Kas Umum"
	SurplusPlanAskDonor SurplusPlan = "ask_donor" // "Minta persetujuan donatur dulu"
)

func (p SurplusPlan) IsValid() bool {
	return p == SurplusPlanToUmum || p == SurplusPlanAskDonor
}

// Pocket is a kas: whose money it is and what it may be spent on
type Pocket struct {
	ID           int64        `json:"id"`
	Kind         PocketKind   `json:"kind"`
	Name         string       `json:"name"`
	Purpose      string       `json:"purpose"`       // peruntukan
	TargetAmount *int64       `json:"target_amount"` // terikat only; nil = no target
	SurplusPlan  *SurplusPlan `json:"surplus_plan"`  // terikat only
	ShowOnTV     bool         `json:"show_on_tv"`
	Status       PocketStatus `json:"status"`
	CreatedBy    int64        `json:"created_by"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// IsOpen is a pocket that still takes transactions
func (p *Pocket) IsOpen() bool {
	return p.Status == PocketStatusActive || p.Status == PocketStatusTargetReached
}

// CategoryType is the direction of a category and of the transactions in it
type CategoryType string

const (
	CategoryTypeIn  CategoryType = "in"  // pemasukan
	CategoryTypeOut CategoryType = "out" // pengeluaran
)

func (t CategoryType) IsValid() bool {
	return t == CategoryTypeIn || t == CategoryTypeOut
}

// CategoryScope is which pockets a category belongs to
type CategoryScope string

const (
	CategoryScopeUmumTerikat CategoryScope = "umum_terikat"
	CategoryScopeZakat       CategoryScope = "zakat"
	CategoryScopeFidyah      CategoryScope = "fidyah"
	CategoryScopeAll         CategoryScope = "all" // Pindah dana only
)

// System category codes. Their rows are seeded by the migration; the code is how the rules find
// them, so their names can never drive behaviour.
const (
	CategoryCodePindahIn    = "pindah_in"
	CategoryCodePindahOut   = "pindah_out"
	CategoryCodeZakatMal    = "zakat_mal"
	CategoryCodeZakatFitrah = "zakat_fitrah"
	CategoryCodeAmil        = "asnaf_amil"
	CategoryCodeFidyah      = "fidyah"
)

// categoryCodePockets ties an income category to the one pocket kind it may go into: zakat mal
// into Zakat Mal, zakat fitrah into Zakat Fitrah
var categoryCodePockets = map[string]PocketKind{
	CategoryCodeZakatMal:    PocketKindZakatMal,
	CategoryCodeZakatFitrah: PocketKindZakatFitrah,
}

// Category groups transactions by purpose (Infaq Jumat, Listrik, Fakir)
type Category struct {
	ID         int64         `json:"id"`
	Name       string        `json:"name"`
	Type       CategoryType  `json:"type"`
	Scope      CategoryScope `json:"scope"`
	Code       *string       `json:"code"` // system categories only
	ArchivedAt *time.Time    `json:"archived_at"`
	CreatedBy  int64         `json:"created_by"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

// IsSystem is a category seeded with rules attached: never renamed or deleted
func (c *Category) IsSystem() bool {
	return c.Code != nil
}

// IsPindah is one side of a Pindah dana pair, left out of the pemasukan/pengeluaran totals
func (c *Category) IsPindah() bool {
	return c.Code != nil && (*c.Code == CategoryCodePindahIn || *c.Code == CategoryCodePindahOut)
}

// FitsPocket is whether a transaction in this category may go into a pocket of the kind
func (c *Category) FitsPocket(kind PocketKind) bool {
	if c.Scope == CategoryScopeAll {
		return true
	}
	if c.Scope != kind.CategoryScope() {
		return false
	}
	if c.Code == nil {
		return true
	}
	only, ok := categoryCodePockets[*c.Code]
	return !ok || only == kind
}

// Transaction is one movement of money in one account and one pocket. There are only two kinds:
// a Pindah dana, a split income or a "Selesaikan program" is several rows sharing a GroupID.
type Transaction struct {
	ID           int64        `json:"id"`
	Kind         CategoryType `json:"kind"`
	Amount       int64        `json:"amount"` // integer Rupiah, > 0
	Date         string       `json:"date"`   // YYYY-MM-DD, the mosque's date
	AccountID    int64        `json:"account_id"`
	PocketID     int64        `json:"pocket_id"`
	CategoryID   int64        `json:"category_id"`
	GroupID      *string      `json:"group_id"`
	DonorConsent bool         `json:"donor_consent"`
	CorrectsID   *int64       `json:"corrects_id"`
	Note         string       `json:"note"`
	DonorNote    string       `json:"donor_note"`
	HasPhoto     bool         `json:"has_photo"`
	CreatedBy    int64        `json:"created_by"`
	UpdatedBy    int64        `json:"updated_by"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	DeletedAt    *time.Time   `json:"deleted_at"`
	DeletedBy    int64        `json:"deleted_by"`
}

// IsEdited is shown as "Diedit": changed after it was first saved
func (t *Transaction) IsEdited() bool {
	return t.UpdatedAt.After(t.CreatedAt)
}

// Signed is the amount as it moves a balance: + for in, − for out
func (t *Transaction) Signed() int64 {
	if t.Kind == CategoryTypeOut {
		return -t.Amount
	}
	return t.Amount
}

// TransactionPhoto is the receipt of a transaction
type TransactionPhoto struct {
	ContentType string
	Data        []byte
}

// RevisionAction is what happened to a transaction
type RevisionAction string

const (
	RevisionEdit   RevisionAction = "edit"
	RevisionDelete RevisionAction = "delete"
)

// TransactionRevision keeps a transaction as it was before an edit or a delete, with the reason
type TransactionRevision struct {
	ID            int64          `json:"id"`
	TransactionID int64          `json:"transaction_id"`
	Action        RevisionAction `json:"action"`
	Previous      Transaction    `json:"previous"`
	Reason        string         `json:"reason"`
	UserID        int64          `json:"user_id"`
	CreatedAt     time.Time      `json:"created_at"`
}

// BookClosingKind tells the opening (Mulai pembukuan) from a month's closing
type BookClosingKind string

const (
	BookClosingOpening BookClosingKind = "opening"
	BookClosingClosing BookClosingKind = "closing"
)

// BookClosing is a month's end: its totals and every account's and pocket's balance, which the
// next balances start from. The opening is a closing of the month before pembukuan starts.
type BookClosing struct {
	ID             int64                `json:"id"`
	Month          string               `json:"month"` // YYYY-MM
	Kind           BookClosingKind      `json:"kind"`
	TotalIn        int64                `json:"total_in"`
	TotalOut       int64                `json:"total_out"`
	ClosingBalance int64                `json:"closing_balance"`
	Accounts       []BookClosingAccount `json:"accounts"`
	Pockets        []BookClosingPocket  `json:"pockets"`
	ClosedBy       int64                `json:"closed_by"`
	ClosedAt       time.Time            `json:"closed_at"`
	ReopenedAt     *time.Time           `json:"reopened_at"`
	ReopenReason   string               `json:"reopen_reason"`
}

type BookClosingAccount struct {
	AccountID      int64  `json:"account_id"`
	ClosingBalance int64  `json:"closing_balance"` // computed
	CountedBalance *int64 `json:"counted_balance"` // entered; nil for the opening
}

type BookClosingPocket struct {
	PocketID       int64 `json:"pocket_id"`
	ClosingBalance int64 `json:"closing_balance"`
}

// TransactionFilter selects transactions; zero fields select everything
type TransactionFilter struct {
	From           string // YYYY-MM-DD, inclusive
	To             string // YYYY-MM-DD, inclusive
	PocketID       int64
	AccountID      int64
	GroupID        string
	IncludeDeleted bool
}
