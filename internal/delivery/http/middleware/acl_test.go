package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
)

// serveAs runs a handler behind RequireRole(roles...) for a user with role,
// reporting whether the handler ran
func serveAs(role domain.UserRole, roles ...domain.UserRole) (bool, error) {
	c := echo.New().NewContext(httptest.NewRequest(http.MethodGet, "/protected", nil), httptest.NewRecorder())
	c.Set(sessionContextKey, &domain.AuthSession{ID: "abc", UserID: 7, Role: role})

	ran := false
	err := RequireRole(roles...)(func(echo.Context) error {
		ran = true
		return nil
	})(c)
	return ran, err
}

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name    string
		role    domain.UserRole
		allowed []domain.UserRole
		wantRun bool
	}{
		{name: "super admin on super admin route", role: domain.UserRoleSuperAdmin, allowed: []domain.UserRole{domain.UserRoleSuperAdmin}, wantRun: true},
		{name: "takmir on super admin route", role: domain.UserRoleTakmir, allowed: []domain.UserRole{domain.UserRoleSuperAdmin}},
		{name: "bendahara on super admin route", role: domain.UserRoleBendahara, allowed: []domain.UserRole{domain.UserRoleSuperAdmin}},
		{name: "bendahara on kas route", role: domain.UserRoleBendahara, allowed: []domain.UserRole{domain.UserRoleSuperAdmin, domain.UserRoleBendahara}, wantRun: true},
		{name: "takmir on kas route", role: domain.UserRoleTakmir, allowed: []domain.UserRole{domain.UserRoleSuperAdmin, domain.UserRoleBendahara}},
		{name: "role removed from the app", role: "user", allowed: []domain.UserRole{domain.UserRoleSuperAdmin, domain.UserRoleTakmir, domain.UserRoleBendahara}},
		{name: "no role allowed", role: domain.UserRoleSuperAdmin, allowed: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ran, err := serveAs(tt.role, tt.allowed...)

			if ran != tt.wantRun {
				t.Errorf("handler ran = %v, want %v", ran, tt.wantRun)
			}
			if tt.wantRun && err != nil {
				t.Errorf("err = %v, want nil", err)
			}
			if !tt.wantRun {
				assertCode(t, err, apperror.CodeForbidden)
			}
		})
	}
}

func TestRequireRoleWithoutRequireAuth(t *testing.T) {
	c := echo.New().NewContext(httptest.NewRequest(http.MethodGet, "/protected", nil), httptest.NewRecorder())

	ran := false
	err := RequireRole(domain.UserRoleSuperAdmin)(func(echo.Context) error {
		ran = true
		return nil
	})(c)

	// a route wired without RequireAuth is a bug, not a client error
	assertCode(t, err, apperror.CodeInternal)
	if ran {
		t.Error("handler ran without a session")
	}
}
