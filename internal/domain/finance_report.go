package domain

import (
	"errors"
	"fmt"
	"time"
)

// Summaries, tutup buku, "Selesaikan program", syariah warnings and the TV payload (E7-04,
// E7-03, E7-16, E7-20). Every number here is computed by the finance usecase.

var ErrFinanceTVSettingNotFound = errors.New("finance tv setting not saved yet")

// Flow is how one kas or account moved over a period. Pindah dana is kept apart from pemasukan
// and pengeluaran, so Opening + In − Out + Moved = Closing.
type Flow struct {
	Opening int64 `json:"opening"`
	In      int64 `json:"in"`
	Out     int64 `json:"out"`
	Moved   int64 `json:"moved"` // pindah dana in minus out
	Closing int64 `json:"closing"`

	movedIn, movedOut int64
}

func (f *Flow) Add(tx Transaction, pindah bool) {
	f.Closing += tx.Signed()
	if pindah {
		f.Moved += tx.Signed()
		if tx.Kind == CategoryTypeIn {
			f.movedIn += tx.Amount
			return
		}
		f.movedOut += tx.Amount
		return
	}
	if tx.Kind == CategoryTypeIn {
		f.In += tx.Amount
		return
	}
	f.Out += tx.Amount
}

// WithMoves counts Pindah dana as this kas's own pemasukan and pengeluaran, so one kas's four
// columns on the TV add up (Opening + In − Out = Closing). The mosque's totals never do this.
func (f Flow) WithMoves() Flow {
	f.In += f.movedIn
	f.Out += f.movedOut
	f.Moved = 0
	return f
}

type PocketFlow struct {
	PocketID int64 `json:"pocket_id"`
	Flow
	TodayIn  int64 `json:"today_in"`
	TodayOut int64 `json:"today_out"`
}

type AccountFlow struct {
	AccountID int64 `json:"account_id"`
	Flow
}

// WarningKind names a syariah warning
type WarningKind string

const (
	WarningAmilCap        WarningKind = "amil_cap"        // amil above 1/8 of the zakat income this Hijri year
	WarningFitrahDeadline WarningKind = "fitrah_deadline" // Zakat Fitrah still holding money after 1 Syawal
)

type FinanceWarning struct {
	Kind     WarningKind `json:"kind"`
	PocketID int64       `json:"pocket_id"`
	Message  string      `json:"message"`
}

// FinanceSummary is one month of Keuangan: per kas and per account, the month's totals (Pindah
// dana left out) and the warnings that apply today
type FinanceSummary struct {
	Month    string           `json:"month"` // YYYY-MM
	Today    string           `json:"today"` // YYYY-MM-DD on the mosque's clock
	Closed   bool             `json:"closed"`
	TotalIn  int64            `json:"total_in"`
	TotalOut int64            `json:"total_out"`
	TodayIn  int64            `json:"today_in"`
	TodayOut int64            `json:"today_out"`
	Balance  int64            `json:"balance"` // every kas together at the end of the month (now for the current one)
	Pockets  []PocketFlow     `json:"pockets"`
	Accounts []AccountFlow    `json:"accounts"`
	Warnings []FinanceWarning `json:"warnings"`
}

// PocketProgress is how far a terikat kas is toward its target: everything that came in,
// Pindah dana from Kas Umum included
type PocketProgress struct {
	Collected     int64 `json:"collected"`
	TargetReached bool  `json:"target_reached"`
}

// FinishPocketRequest is "Selesaikan program": the remainder moves to Kas Umum, then the kas is
// archived
type FinishPocketRequest struct {
	UserID int64 `json:"-"`
	ID     int64 `json:"-"`

	AccountID    int64  `json:"account_id"` // where the remainder sits; the move does not change it
	DonorConsent bool   `json:"donor_consent"`
	Note         string `json:"note"`
}

func (r *FinishPocketRequest) ValidateAndNormalize(plan SurplusPlan) error {
	var err error
	if r.Note, err = validateText(r.Note, MaxTransactionNoteLength, "catatan"); err != nil {
		return err
	}
	if plan == SurplusPlanAskDonor && (!r.DonorConsent || r.Note == "") {
		return invalidInput(errors.New("ask_donor without consent"), "centang persetujuan donatur dan isi catatannya")
	}
	return nil
}

func PocketFinishedActivityLog(p Pocket, remainder int64) ActivityLog {
	return ActivityLog{Action: ActivityPocketFinished, EntityID: p.ID, Summary: fmt.Sprintf("menyelesaikan program %s, sisa %s ke Kas Umum", quoteForSummary(p.Name), rupiahForSummary(remainder))}
}

// CountedBalance is what the bendahara counted in one account at the end of the month
type CountedBalance struct {
	AccountID int64 `json:"account_id"`
	Counted   int64 `json:"counted"`
}

// CloseMonthRequest is Tutup buku of one month
type CloseMonthRequest struct {
	UserID int64 `json:"-"`

	Month    string           `json:"month"` // YYYY-MM
	Accounts []CountedBalance `json:"accounts"`
}

func (r *CloseMonthRequest) ValidateAndNormalize() error {
	if _, err := time.Parse(monthLayout, r.Month); err != nil {
		return invalidInput(fmt.Errorf("invalid month %q", r.Month), "bulan tidak valid")
	}
	seen := map[int64]bool{}
	for _, a := range r.Accounts {
		if a.AccountID <= 0 || seen[a.AccountID] {
			return invalidInput(fmt.Errorf("account %d invalid or repeated", a.AccountID), "akun tidak valid atau terulang")
		}
		seen[a.AccountID] = true
		if a.Counted < 0 {
			return invalidInput(fmt.Errorf("counted %d", a.Counted), "saldo hitung tidak boleh negatif")
		}
	}
	return nil
}

func (r *CloseMonthRequest) ToActivityLog(id int64) ActivityLog {
	return ActivityLog{Action: ActivityMonthClosed, EntityID: id, Summary: "tutup buku " + monthForSummary(r.Month)}
}

// TVPeriod is what the TV kas header covers
type TVPeriod string

const (
	TVPeriodToday TVPeriod = "today"
	TVPeriodWeek  TVPeriod = "week"
	TVPeriodMonth TVPeriod = "month"
)

func (p TVPeriod) IsValid() bool {
	return p == TVPeriodToday || p == TVPeriodWeek || p == TVPeriodMonth
}

// From is the first date (YYYY-MM-DD) of the period that holds today; a week starts on Monday
func (p TVPeriod) From(today time.Time) string {
	if p == TVPeriodToday {
		return today.Format(dateLayout)
	}
	if p == TVPeriodWeek {
		back := (int(today.Weekday()) + 6) % 7
		return today.AddDate(0, 0, -back).Format(dateLayout)
	}
	return today.Format(monthLayout) + "-01"
}

type FinanceTVSetting struct {
	Period TVPeriod `json:"period"`
}

func DefaultFinanceTVSetting() FinanceTVSetting {
	return FinanceTVSetting{Period: TVPeriodMonth}
}

type UpdateFinanceTVSettingRequest struct {
	UpdatedBy int64 `json:"-"`

	FinanceTVSetting
}

func (r *UpdateFinanceTVSettingRequest) ValidateAndNormalize() error {
	if !r.Period.IsValid() {
		return invalidInput(fmt.Errorf("unknown period %q", r.Period), "periode harus hari ini, minggu ini, atau bulan ini")
	}
	return nil
}

func (r *UpdateFinanceTVSettingRequest) ToActivityLog() ActivityLog {
	return ActivityLog{Action: ActivityFinanceTVSettingUpdated, Summary: "mengubah periode kas di layar TV"}
}

// TVPocket is one kas in the TV header: four static columns, one kas at a time
type TVPocket struct {
	ID   int64      `json:"id"`
	Name string     `json:"name"`
	Kind PocketKind `json:"kind"`
	Flow
}

// FinanceDisplay is what the TV shows of Keuangan
type FinanceDisplay struct {
	Period  TVPeriod   `json:"period"`
	From    string     `json:"from"`
	To      string     `json:"to"`
	Pockets []TVPocket `json:"pockets"`
}
