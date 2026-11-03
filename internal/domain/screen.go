package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// Screens are the TVs that opened /tv (E9). A TV makes a random device ID once and keeps it, then
// holds one live connection (SSE) to the API: being connected is being online, and commands come
// down that connection. The database keeps one row per device, written when it first connects
// and when an admin changes it; last seen and the app version live in the cache.

var (
	ErrScreenNotFound         = errors.New("screen not found")
	ErrScreenPresenceNotFound = errors.New("screen not seen yet")
	ErrScreenBlocked          = errors.New("screen blocked")
)

const (
	// MaxScreens caps the rows a public connect can create; blocked ones count too
	MaxScreens = 50
	// ScreenOfflineAfter is how long a screen may be disconnected and still show online, so a
	// reconnect after a Wi-Fi blip doesn't flash Offline
	ScreenOfflineAfter = 2 * time.Minute

	MaxScreenNameLength    = 50
	maxScreenUserAgentLen  = 300
	maxScreenAppVersionLen = 30
)

// deviceIDPattern is what a TV may use as its ID: a UUID or similar random string
var deviceIDPattern = regexp.MustCompile(`^[A-Za-z0-9-]{8,64}$`)

// ScreenMode is what a screen shows: its own schedule (auto) or what an admin forced
type ScreenMode string

const (
	ScreenModeAuto   ScreenMode = "auto"
	ScreenModeNormal ScreenMode = "normal" // the normal layout, even at prayer time
	ScreenModeJumat  ScreenMode = "jumat"  // the Jumat screen
	ScreenModeDark   ScreenMode = "dark"   // a black screen
)

func (m ScreenMode) IsValid() bool {
	switch m {
	case ScreenModeAuto, ScreenModeNormal, ScreenModeJumat, ScreenModeDark:
		return true
	default:
		return false
	}
}

type ScreenStatus string

const (
	ScreenStatusOnline  ScreenStatus = "online"
	ScreenStatusOffline ScreenStatus = "offline"
)

// Screen is one TV. The repository fills the stored fields; the usecase adds the presence ones
// (Status, LastSeenAt, AppVersion) from the cache and the live connections.
type Screen struct {
	ID          int64      `json:"id"`
	DeviceID    string     `json:"device_id"`
	Name        string     `json:"name"`      // "" until an admin names it
	NameHint    string     `json:"name_hint"` // from /tv?layar=…
	UserAgent   string     `json:"user_agent"`
	ForcedMode  ScreenMode `json:"forced_mode"`
	ForcedUntil *time.Time `json:"forced_until"`
	BlockedAt   *time.Time `json:"blocked_at"`
	FirstSeenAt time.Time  `json:"first_seen_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	Status     ScreenStatus `json:"status"`
	LastSeenAt *time.Time   `json:"last_seen_at"` // nil when the cache has nothing, e.g. after a cache flush
	AppVersion string       `json:"app_version"`
}

// IsNew is a screen nobody has named yet (BARU)
func (s *Screen) IsNew() bool {
	return s.Name == ""
}

func (s *Screen) IsBlocked() bool {
	return s.BlockedAt != nil
}

// ModeAt is the mode in force at now: a forced mode ends at ForcedUntil
func (s *Screen) ModeAt(now time.Time) ScreenMode {
	if s.ForcedUntil == nil || !now.Before(*s.ForcedUntil) {
		return ScreenModeAuto
	}
	return s.ForcedMode
}

// ScreenPresence is what the cache keeps per device
type ScreenPresence struct {
	LastSeenAt time.Time `json:"last_seen_at"`
	AppVersion string    `json:"app_version"`
}

// ConnectScreenRequest is a TV opening its connection; everything comes from the TV, unchecked
type ConnectScreenRequest struct {
	DeviceID   string
	NameHint   string
	AppVersion string
	UserAgent  string
}

func (r *ConnectScreenRequest) ValidateAndNormalize() error {
	if !deviceIDPattern.MatchString(r.DeviceID) {
		return invalid("bad device id", "id layar tidak valid")
	}
	r.NameHint = truncateRunes(strings.TrimSpace(r.NameHint), MaxScreenNameLength)
	r.AppVersion = truncateRunes(strings.TrimSpace(r.AppVersion), maxScreenAppVersionLen)
	r.UserAgent = truncateRunes(strings.TrimSpace(r.UserAgent), maxScreenUserAgentLen)
	return nil
}

// UpdateScreenRequest names a screen and sets its mode (PUT)
type UpdateScreenRequest struct {
	// ID comes from the path; the controller sets it
	ID int64 `json:"-"`

	Name       string     `json:"name"`
	ForcedMode ScreenMode `json:"forced_mode"`
}

func (r *UpdateScreenRequest) ValidateAndNormalize() error {
	r.Name = strings.TrimSpace(r.Name)
	if r.Name == "" {
		return invalid("empty name", "nama layar wajib diisi")
	}
	if utf8.RuneCountInString(r.Name) > MaxScreenNameLength {
		return invalid("name too long", fmt.Sprintf("nama layar maksimal %d karakter", MaxScreenNameLength))
	}
	if !r.ForcedMode.IsValid() {
		return invalid("bad mode "+string(r.ForcedMode), "mode tampilan tidak valid")
	}
	return nil
}

// ScreenCommand is a one-off order to a TV
type ScreenCommand string

const (
	ScreenCommandReload    ScreenCommand = "reload"
	ScreenCommandTestSound ScreenCommand = "test_sound"
)

func (c ScreenCommand) IsValid() bool {
	return c == ScreenCommandReload || c == ScreenCommandTestSound
}

// SendScreenCommandRequest sends a command to one screen, or to every connected one when
// ScreenID is 0
type SendScreenCommandRequest struct {
	ScreenID int64         `json:"-"`
	Command  ScreenCommand `json:"command"`
}

func (r *SendScreenCommandRequest) Validate() error {
	if !r.Command.IsValid() {
		return invalid("bad command "+string(r.Command), "perintah tidak valid")
	}
	return nil
}

// ScreenEventType names what a TV receives on its connection
type ScreenEventType string

const (
	ScreenEventState     ScreenEventType = "state"   // its name and mode; sent on connect and on change
	ScreenEventRefresh   ScreenEventType = "refresh" // content changed: reload the data now
	ScreenEventReload    ScreenEventType = "reload"  // reload the page
	ScreenEventTestSound ScreenEventType = "test_sound"
	ScreenEventBlocked   ScreenEventType = "blocked" // an admin blocked it; the connection ends
)

type ScreenEvent struct {
	Type ScreenEventType
	Data any
}

// ScreenState is what a TV needs to know about itself
type ScreenState struct {
	Name        string     `json:"name"`
	ForcedMode  ScreenMode `json:"forced_mode"`
	ForcedUntil *time.Time `json:"forced_until"`
}

func (s *Screen) State() ScreenState {
	return ScreenState{Name: s.Name, ForcedMode: s.ForcedMode, ForcedUntil: s.ForcedUntil}
}

// ScreenConnection is an open TV connection: its first state and the events that follow. Close
// must be called when the TV goes away.
type ScreenConnection struct {
	State  ScreenState
	Events <-chan ScreenEvent
	Close  func()
}

// Label names the screen to people: its name, else the name it suggested, else "baru"
func (s *Screen) Label() string {
	switch {
	case s.Name != "":
		return s.Name
	case s.NameHint != "":
		return s.NameHint
	default:
		return "baru"
	}
}

func (s *Screen) BlockedActivityLog() ActivityLog {
	return ActivityLog{Action: ActivityScreenBlocked, EntityID: s.ID, Summary: "memblokir layar " + s.Label()}
}

func (s *Screen) UnblockedActivityLog() ActivityLog {
	return ActivityLog{Action: ActivityScreenUnblocked, EntityID: s.ID, Summary: "membuka blokir layar " + s.Label()}
}

// ToActivityLog is the log of this update; call after ValidateAndNormalize (the name is required)
func (r *UpdateScreenRequest) ToActivityLog() ActivityLog {
	return ActivityLog{Action: ActivityScreenUpdated, EntityID: r.ID, Summary: "mengubah layar " + r.Name}
}
