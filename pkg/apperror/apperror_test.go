package apperror

import (
	"errors"
	"fmt"
	"testing"
)

var errCause = errors.New("phone already taken")

func TestError(t *testing.T) {
	tests := []struct {
		name string
		err  *AppError
		want string
	}{
		{name: "with internal error", err: New(CodeAlreadyExists, errCause, "nomor HP sudah terdaftar"), want: "phone already taken"},
		{name: "without internal error", err: New(CodeInvalidInput, nil, "nomor HP wajib diisi"), want: "nomor HP wajib diisi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUnwrap(t *testing.T) {
	err := fmt.Errorf("creating user: %w", New(CodeAlreadyExists, errCause, "nomor HP sudah terdaftar"))

	if !errors.Is(err, errCause) {
		t.Error("errors.Is does not reach the internal error")
	}

	appErr, ok := errors.AsType[*AppError](err)
	if !ok || appErr.Code != CodeAlreadyExists {
		t.Errorf("errors.AsType = %v, want an ALREADY_EXISTS AppError", appErr)
	}
}
