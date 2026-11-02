package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// NotifyScreens calls notify after every successful change (not a GET) to a route under one of
// paths, so TVs reload what they show within seconds instead of at their next 5-minute refresh.
// paths are route prefixes as registered, e.g. "/api/v1/slides".
func NotifyScreens(notify func(), paths ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil || c.Request().Method == http.MethodGet || c.Response().Status >= http.StatusMultipleChoices {
				return err
			}
			for _, p := range paths {
				if strings.HasPrefix(c.Path(), p) {
					notify()
					return nil
				}
			}
			return nil
		}
	}
}
