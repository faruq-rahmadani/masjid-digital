package users

import (
	"reflect"
	"testing"
	"time"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
)

func TestUserModelRoundTrip(t *testing.T) {
	wib := time.FixedZone("WIB", 7*60*60)
	lastLoginAt := time.Date(2026, 10, 31, 14, 0, 0, 0, wib)

	tests := []struct {
		name string
		user domain.User
	}{
		{
			name: "all fields",
			user: domain.User{
				ID:           1,
				Name:         "Ahmad",
				Phone:        "+6281200000001",
				Email:        "ahmad@example.com",
				PasswordHash: "hash",
				Role:         domain.UserRoleSuperAdmin,
				Status:       domain.UserStatusActive,
				LastLoginAt:  &lastLoginAt,
				CreatedBy:    3,
				CreatedAt:    lastLoginAt,
				UpdatedAt:    lastLoginAt,
			},
		},
		{
			name: "optional fields empty",
			user: domain.User{
				ID:     2,
				Phone:  "+6281200000002",
				Role:   domain.UserRoleTakmir,
				Status: domain.UserStatusDisabled,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newUserModel(tt.user)

			if tt.user.Email == "" && m.Email != nil {
				t.Errorf("empty email stored as %q, want NULL", *m.Email)
			}
			if tt.user.CreatedBy == 0 && m.CreatedBy != nil {
				t.Errorf("created_by 0 stored as %d, want NULL", *m.CreatedBy)
			}
			if m.CreatedAt.Location() != time.UTC {
				t.Errorf("created_at stored in %s, want UTC", m.CreatedAt.Location())
			}

			got := m.toDomain()
			if !got.CreatedAt.Equal(tt.user.CreatedAt) || !got.UpdatedAt.Equal(tt.user.UpdatedAt) {
				t.Errorf("timestamps changed: got %v/%v, want %v/%v", got.CreatedAt, got.UpdatedAt, tt.user.CreatedAt, tt.user.UpdatedAt)
			}
			if (got.LastLoginAt == nil) != (tt.user.LastLoginAt == nil) ||
				(got.LastLoginAt != nil && !got.LastLoginAt.Equal(*tt.user.LastLoginAt)) {
				t.Errorf("last_login_at = %v, want %v", got.LastLoginAt, tt.user.LastLoginAt)
			}

			// compare the rest, ignoring time zones
			got.CreatedAt, got.UpdatedAt, got.LastLoginAt = tt.user.CreatedAt, tt.user.UpdatedAt, tt.user.LastLoginAt
			if !reflect.DeepEqual(got, tt.user) {
				t.Errorf("round trip = %+v, want %+v", got, tt.user)
			}
		})
	}
}

func TestToDomainReturnsUTC(t *testing.T) {
	// what pgx hands back for TIMESTAMPTZ on a server running in WIB
	wib := time.FixedZone("WIB", 7*60*60)
	at := time.Date(2026, 10, 31, 10, 20, 54, 0, wib)

	got := userModel{CreatedAt: at, UpdatedAt: at, LastLoginAt: &at}.toDomain()

	for name, value := range map[string]time.Time{"created_at": got.CreatedAt, "updated_at": got.UpdatedAt, "last_login_at": *got.LastLoginAt} {
		if value.Location() != time.UTC || !value.Equal(at) {
			t.Errorf("%s = %v, want %v in UTC", name, value, at.UTC())
		}
	}
}
