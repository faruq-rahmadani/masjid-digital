package http

import (
	"errors"
	"log"
	"net/http"

	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
	"github.com/labstack/echo/v4"
)

// httpStatuses maps each internal error code to the HTTP status returned to the client
var httpStatuses = map[apperror.Code]int{
	apperror.CodeInvalidInput:    http.StatusBadRequest,
	apperror.CodeUnauthenticated: http.StatusUnauthorized,
	apperror.CodeForbidden:       http.StatusForbidden,
	apperror.CodeNotFound:        http.StatusNotFound,
	apperror.CodeAlreadyExists:   http.StatusConflict,
	apperror.CodeBusinessRule:    http.StatusUnprocessableEntity,
	apperror.CodeInternal:        http.StatusInternalServerError,
}

type errorResponse struct {
	Code    apperror.Code `json:"code"`
	Message string        `json:"message"`
}

// httpStatus returns the HTTP status of an internal code, 500 when unmapped
func httpStatus(code apperror.Code) int {
	if status, ok := httpStatuses[code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

// codeFromStatus is the reverse of httpStatus, for errors raised by Echo itself.
// Unmapped statuses fall back to INVALID_INPUT (4xx) or INTERNAL (5xx).
func codeFromStatus(status int) apperror.Code {
	for code, s := range httpStatuses {
		if s == status {
			return code
		}
	}
	if status >= http.StatusInternalServerError {
		return apperror.CodeInternal
	}
	return apperror.CodeInvalidInput
}

// newErrorHandler renders every error as {code, message}:
//   - *apperror.AppError with the status mapped from its code
//   - Echo's own errors (unknown route, bad method, body too large, ...) with their status
//   - anything else as a 500, so internal details never reach the client
func newErrorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		status, appErr := toAppError(err)
		if status >= http.StatusInternalServerError {
			log.Printf("request %s %s failed: request_id=%s code=%s err=%v",
				c.Request().Method,
				c.Path(),
				c.Response().Header().Get(echo.HeaderXRequestID),
				appErr.Code,
				appErr,
			)
		}

		if c.Request().Method == http.MethodHead {
			err = c.NoContent(status)
		} else {
			err = c.JSON(status, errorResponse{Code: appErr.Code, Message: appErr.MessageError})
		}
		if err != nil {
			log.Printf("error writing error response: %v", err)
		}
	}
}

// toAppError returns the HTTP status and the AppError to render for err
func toAppError(err error) (int, *apperror.AppError) {
	// AppError first: it may wrap an *echo.HTTPError, e.g. from c.Bind
	if appErr, ok := errors.AsType[*apperror.AppError](err); ok {
		return httpStatus(appErr.Code), appErr
	}

	if httpErr, ok := errors.AsType[*echo.HTTPError](err); ok {
		message := http.StatusText(httpErr.Code)
		if text, ok := httpErr.Message.(string); ok && httpErr.Code < http.StatusInternalServerError {
			message = text
		}
		return httpErr.Code, apperror.New(codeFromStatus(httpErr.Code), err, message)
	}

	return http.StatusInternalServerError, apperror.New(apperror.CodeInternal, err, http.StatusText(http.StatusInternalServerError))
}
