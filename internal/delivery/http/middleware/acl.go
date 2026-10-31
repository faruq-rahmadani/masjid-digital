package middleware

import (
	"fmt"
	"slices"

	"github.com/faruq-rahmadani/masjid-digital/internal/domain"
	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
)

// RequireRole lets a request through only when the logged in user has one of
// roles, and answers 403 otherwise. It reads the session stored by
// RequireAuth, so it must run after it, e.g. as a route middleware on a group
// that uses RequireAuth.
//
// The role comes from the access token. A role change logs the user out (see
// user.RevokeSessionOnRoleChangeHook), so a token never carries a stale role.
func RequireRole(roles ...domain.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			session, err := Session(c)
			if err != nil {
				return err
			}

			if !slices.Contains(roles, session.Role) {
				return apperror.New(apperror.CodeForbidden,
					fmt.Errorf("user %d with role %q needs one of %v for %s %s", session.UserID, session.Role, roles, c.Request().Method, c.Path()),
					"anda tidak memiliki akses",
				)
			}

			return next(c)
		}
	}
}
