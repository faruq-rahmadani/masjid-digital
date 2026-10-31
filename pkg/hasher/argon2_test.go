package hasher

import (
	"errors"
	"strings"
	"testing"
)

// fastParams keeps tests quick; production uses DefaultArgon2Params
var fastParams = Argon2Params{Memory: 1024, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32}

func TestHashAndVerify(t *testing.T) {
	h := NewArgon2(fastParams)

	hashed, err := h.Hash("bismillah123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !strings.HasPrefix(hashed, "$argon2id$v=19$m=1024,t=1,p=1$") {
		t.Errorf("hash = %q, want argon2id format with its params", hashed)
	}

	tests := []struct {
		name  string
		plain string
		want  bool
	}{
		{name: "same password", plain: "bismillah123", want: true},
		{name: "wrong password", plain: "bismillah124", want: false},
		{name: "empty password", plain: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := h.Verify(tt.plain, hashed)
			if err != nil {
				t.Fatalf("verify: %v", err)
			}
			if got != tt.want {
				t.Errorf("Verify(%q) = %v, want %v", tt.plain, got, tt.want)
			}
		})
	}
}

func TestHashIsSalted(t *testing.T) {
	h := NewArgon2(fastParams)

	first, _ := h.Hash("bismillah123")
	second, _ := h.Hash("bismillah123")
	if first == second {
		t.Error("same password hashed twice gave the same hash, want a random salt")
	}
}

func TestVerifyWithChangedParams(t *testing.T) {
	hashed, err := NewArgon2(fastParams).Hash("bismillah123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}

	stronger := fastParams
	stronger.Memory, stronger.Iterations = 2048, 2

	ok, err := NewArgon2(stronger).Verify("bismillah123", hashed)
	if err != nil || !ok {
		t.Errorf("verify old hash with new params = %v, %v; want true, nil", ok, err)
	}
}

func TestVerifyInvalidHash(t *testing.T) {
	h := NewArgon2(fastParams)

	for _, hashed := range []string{
		"",
		"plaintext",
		"$2a$10$abcdefghijklmnopqrstuv", // bcrypt
		"$argon2i$v=19$m=1024,t=1,p=1$c2FsdA$a2V5",
		"$argon2id$v=18$m=1024,t=1,p=1$c2FsdA$a2V5",
		"$argon2id$v=19$m=0,t=1,p=1$c2FsdA$a2V5",
		"$argon2id$v=19$m=1024,t=1,p=1$!!!$a2V5",
		"$argon2id$v=19$m=1024,t=1,p=1$c2FsdA$",
	} {
		if _, err := h.Verify("bismillah123", hashed); !errors.Is(err, ErrInvalidHash) {
			t.Errorf("Verify(%q) err = %v, want ErrInvalidHash", hashed, err)
		}
	}
}

func BenchmarkHashDefault(b *testing.B) {
	h := NewArgon2(DefaultArgon2Params)
	for b.Loop() {
		_, _ = h.Hash("bismillah123")
	}
}
