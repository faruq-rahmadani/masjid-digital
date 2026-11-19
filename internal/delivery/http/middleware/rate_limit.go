package middleware

import (
	"time"

	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

// RateLimit lets each key make burst requests at once, then one more every interval; past
// that it answers TOO_MANY_REQUESTS (429) through the server's error handler. Counts live in memory, per API instance, and
// restart from zero when the API restarts.
func RateLimit(interval time.Duration, burst int, key func(echo.Context) string) echo.MiddlewareFunc {
	tooMany := func(echo.Context, string, error) error {
		return apperror.New(apperror.CodeTooManyRequests, echomiddleware.ErrRateLimitExceeded, "terlalu banyak percobaan, silakan coba lagi nanti")
	}

	return echomiddleware.RateLimiterWithConfig(echomiddleware.RateLimiterConfig{
		Store: echomiddleware.NewRateLimiterMemoryStoreWithConfig(echomiddleware.RateLimiterMemoryStoreConfig{
			Rate:  rate.Every(interval),
			Burst: burst,
			// a key is forgotten only once its bucket would be full again, so waiting for the
			// cleanup never earns more than waiting for the refill
			ExpiresIn: interval * time.Duration(burst),
		}),
		IdentifierExtractor: func(c echo.Context) (string, error) { return key(c), nil },
		DenyHandler:         tooMany,
		// never called: the keys above cannot fail
		ErrorHandler: func(c echo.Context, err error) error { return tooMany(c, "", err) },
	})
}

// ByClientIP keys a rate limit by the caller's IP, see the server's IPExtractor
func ByClientIP(c echo.Context) string {
	return c.RealIP()
}

// Everyone shares one rate limit, capping a route however many IPs call it
func Everyone(echo.Context) string {
	return "everyone"
}
