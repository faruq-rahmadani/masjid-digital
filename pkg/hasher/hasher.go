// Package hasher hashes passwords and checks a password against a stored hash.
// It knows nothing about users or storage; callers keep the hash wherever they like.
package hasher

//go:generate go tool mockgen -source=hasher.go -destination=../../mock/hasher.go -package=mock

import "errors"

// ErrInvalidHash is returned by Verify when the stored hash is malformed or
// was made by another algorithm
var ErrInvalidHash = errors.New("hasher: invalid hash format")

type PasswordHasher interface {
	// Hash returns a self-describing hash of plain, salted with random bytes
	Hash(plain string) (string, error)
	// Verify reports whether plain matches hashed. A mismatch is (false, nil);
	// an error means hashed could not be read.
	Verify(plain, hashed string) (bool, error)
}
