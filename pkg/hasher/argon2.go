package hasher

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2Params tunes argon2id. Higher values are slower to brute force but
// cost more CPU and memory on every login.
type Argon2Params struct {
	Memory      uint32 // KiB
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32 // bytes
	KeyLength   uint32 // bytes
}

// DefaultArgon2Params follows the OWASP minimum (19 MiB, 2 iterations), light
// enough for a mini PC or set-top box running a single mosque
var DefaultArgon2Params = Argon2Params{
	Memory:      19 * 1024,
	Iterations:  2,
	Parallelism: 1,
	SaltLength:  16,
	KeyLength:   32,
}

type argon2Hasher struct {
	params Argon2Params
}

// NewArgon2 returns an argon2id PasswordHasher. Hashes store their own
// parameters, so changing params later still verifies existing hashes.
func NewArgon2(params Argon2Params) PasswordHasher {
	return &argon2Hasher{params: params}
}

// Hash returns $argon2id$v=19$m=<memory>,t=<iterations>,p=<parallelism>$<salt>$<key>
func (h *argon2Hasher) Hash(plain string) (string, error) {
	salt := make([]byte, h.params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("error generating salt: %w", err)
	}

	key := argon2.IDKey([]byte(plain), salt, h.params.Iterations, h.params.Memory, h.params.Parallelism, h.params.KeyLength)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.params.Memory,
		h.params.Iterations,
		h.params.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func (h *argon2Hasher) Verify(plain, hashed string) (bool, error) {
	params, salt, key, err := decodeArgon2(hashed)
	if err != nil {
		return false, err
	}

	other := argon2.IDKey([]byte(plain), salt, params.Iterations, params.Memory, params.Parallelism, uint32(len(key)))

	// constant time, so the response time does not leak how much matched
	return subtle.ConstantTimeCompare(key, other) == 1, nil
}

func decodeArgon2(hashed string) (Argon2Params, []byte, []byte, error) {
	// "", "argon2id", "v=19", "m=..,t=..,p=..", salt, key
	parts := strings.Split(hashed, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return Argon2Params{}, nil, nil, ErrInvalidHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return Argon2Params{}, nil, nil, ErrInvalidHash
	}

	var params Argon2Params
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Iterations, &params.Parallelism); err != nil {
		return Argon2Params{}, nil, nil, ErrInvalidHash
	}
	if params.Memory == 0 || params.Iterations == 0 || params.Parallelism == 0 {
		return Argon2Params{}, nil, nil, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) == 0 {
		return Argon2Params{}, nil, nil, ErrInvalidHash
	}

	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(key) == 0 {
		return Argon2Params{}, nil, nil, ErrInvalidHash
	}

	return params, salt, key, nil
}
