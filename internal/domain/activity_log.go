package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

// The activity log is who changed what. Usecases publish an ActivityLog after each successful
// change; it is saved in the background (see ActivityLogPublisherItf), so a failed save never
// fails or slows the change itself. Home shows the latest few. Rows are never edited; a user is
// never deleted, so the actor's name stays.
//
// Each request builds its own log (req.ToActivityLog, next to the request); a change without a
// request struct (a delete by ID) has a constructor such as SlideDeletedActivityLog.

const (
	DefaultActivityLogLimit = 10
	MaxActivityLogLimit     = 50
)

type ActivityAction string

const (
	ActivityUserCreated       ActivityAction = "user.created"
	ActivityUserUpdated       ActivityAction = "user.updated"
	ActivityUserPasswordReset ActivityAction = "user.password_reset"

	ActivityMosqueProfileUpdated ActivityAction = "mosque_profile.updated"
	ActivityMosqueLogoUpdated    ActivityAction = "mosque_logo.updated"
	ActivityMosqueLogoDeleted    ActivityAction = "mosque_logo.deleted"

	ActivityPrayerSettingsUpdated ActivityAction = "prayer_settings.updated"
	ActivityPrayerOverrideCreated ActivityAction = "prayer_override.created"
	ActivityPrayerOverrideDeleted ActivityAction = "prayer_override.deleted"

	ActivityIqamahSettingsUpdated ActivityAction = "iqamah_settings.updated"

	ActivityFridayDutySaved   ActivityAction = "friday_duty.saved"
	ActivityFridayDutyDeleted ActivityAction = "friday_duty.deleted"

	ActivityRunningTextCreated        ActivityAction = "running_text.created"
	ActivityRunningTextUpdated        ActivityAction = "running_text.updated"
	ActivityRunningTextDeleted        ActivityAction = "running_text.deleted"
	ActivityRunningTextsReordered     ActivityAction = "running_text.reordered"
	ActivityRunningTextSettingUpdated ActivityAction = "running_text_setting.updated"

	ActivitySlideCreated    ActivityAction = "slide.created"
	ActivitySlideUpdated    ActivityAction = "slide.updated"
	ActivitySlideDeleted    ActivityAction = "slide.deleted"
	ActivitySlidesReordered ActivityAction = "slide.reordered"

	ActivityScreenUpdated   ActivityAction = "screen.updated"
	ActivityScreenBlocked   ActivityAction = "screen.blocked"
	ActivityScreenUnblocked ActivityAction = "screen.unblocked"

	ActivityAccountCreated  ActivityAction = "account.created"
	ActivityAccountUpdated  ActivityAction = "account.updated"
	ActivityAccountArchived ActivityAction = "account.archived"
	ActivityAccountRestored ActivityAction = "account.restored"

	ActivityPocketCreated  ActivityAction = "pocket.created"
	ActivityPocketUpdated  ActivityAction = "pocket.updated"
	ActivityPocketArchived ActivityAction = "pocket.archived"
	ActivityPocketRestored ActivityAction = "pocket.restored"

	ActivityCategoryCreated  ActivityAction = "category.created"
	ActivityCategoryUpdated  ActivityAction = "category.updated"
	ActivityCategoryDeleted  ActivityAction = "category.deleted"
	ActivityCategoryArchived ActivityAction = "category.archived"
	ActivityCategoryRestored ActivityAction = "category.restored"

	ActivityBookkeepingStarted ActivityAction = "bookkeeping.started"

	ActivityTransactionCreated ActivityAction = "transaction.created"
	ActivityTransactionUpdated ActivityAction = "transaction.updated"
	ActivityTransactionDeleted ActivityAction = "transaction.deleted"
	ActivityFundsMoved         ActivityAction = "transaction.moved"
	ActivityReceiptUploaded    ActivityAction = "transaction.receipt_uploaded"
	ActivityPocketFinished     ActivityAction = "pocket.finished"
	ActivityMonthClosed        ActivityAction = "book_closing.closed"

	ActivityFinanceTVSettingUpdated ActivityAction = "finance_tv_setting.updated"
)

type ActivityLog struct {
	ID        int64          `json:"id"`
	ActorID   int64          `json:"actor_id"`   // 0 when nobody was logged in
	ActorName string         `json:"actor_name"` // "" when nobody was logged in; filled when read
	Action    ActivityAction `json:"action"`
	EntityID  int64          `json:"entity_id"` // the changed row, 0 for a setting or when unknown
	// Summary is what happened in Indonesian, shown after the actor's name ("Hasan mengubah
	// pengaturan iqamah"), so it starts with a lowercase verb
	Summary   string    `json:"summary"`
	CreatedAt time.Time `json:"created_at"` // when it happened, UTC
}

// ListActivityLogsRequest is GET /activity-logs?limit=N; Limit is the raw query parameter
type ListActivityLogsRequest struct {
	Limit string
}

// ValidateAndParse returns how many logs to list: DefaultActivityLogLimit when Limit is empty,
// else 1..MaxActivityLogLimit
func (r ListActivityLogsRequest) ValidateAndParse() (int, error) {
	if r.Limit == "" {
		return DefaultActivityLogLimit, nil
	}
	limit, err := strconv.Atoi(r.Limit)
	if err != nil || limit < 1 || limit > MaxActivityLogLimit {
		return 0, apperror.New(apperror.CodeInvalidInput, fmt.Errorf("invalid limit %q", r.Limit), fmt.Sprintf("limit harus angka 1-%d", MaxActivityLogLimit))
	}
	return limit, nil
}

// maxSummaryQuoteLength caps user text quoted in a summary (a running text is up to 200 characters)
const maxSummaryQuoteLength = 40

// quoteForSummary quotes user text for a summary, shortened to maxSummaryQuoteLength characters
func quoteForSummary(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if utf8.RuneCountInString(s) > maxSummaryQuoteLength {
		s = strings.TrimSpace(string([]rune(s)[:maxSummaryQuoteLength-1])) + "…"
	}
	return `"` + s + `"`
}

var indonesianMonths = [...]string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}

// dateForSummary turns a YYYY-MM-DD date into "6 Nov 2026"; anything else is returned as is
func dateForSummary(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.Format("2 ") + indonesianMonths[t.Month()-1] + t.Format(" 2006")
}
