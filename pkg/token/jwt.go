// Package token issues and parses signed access tokens (JWT, HS256).
// It knows nothing about users; callers pass the subject and role.
package token

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// MinSecretLength is the shortest accepted signing secret, in bytes
const MinSecretLength = 32

// ErrInvalidToken is returned by Parse for a malformed, expired or forged token
var ErrInvalidToken = errors.New("token: invalid or expired token")

// Issued is a freshly signed token
type Issued struct {
	Value     string        // the signed token sent to the client
	ID        string        // unique token ID (jti), e.g. to track it as a session
	ExpiresIn time.Duration // lifetime of the token
}

// Claims is what a token says about its holder
type Claims struct {
	ID        string // unique token ID (jti)
	UserID    int64
	Role      string
	ExpiresAt time.Time
}

type jwtClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type JWT struct {
	secret []byte
	ttl    time.Duration
	issuer string
	now    func() time.Time
}

// NewJWT returns a JWT issuer signing with secret. Tokens expire after ttl.
func NewJWT(secret string, ttl time.Duration, issuer string) (*JWT, error) {
	if len(secret) < MinSecretLength {
		return nil, fmt.Errorf("jwt secret must be at least %d bytes, got %d", MinSecretLength, len(secret))
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("token ttl must be positive, got %s", ttl)
	}

	return &JWT{secret: []byte(secret), ttl: ttl, issuer: issuer, now: time.Now}, nil
}

// Issue signs a token with a new random ID for the user
func (j *JWT) Issue(userID int64, role string) (Issued, error) {
	id, err := newTokenID()
	if err != nil {
		return Issued{}, err
	}

	now := j.now()

	claims := jwtClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        id,
			Subject:   strconv.FormatInt(userID, 10),
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.ttl)),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(j.secret)
	if err != nil {
		return Issued{}, fmt.Errorf("error signing token: %w", err)
	}

	return Issued{Value: signed, ID: id, ExpiresIn: j.ttl}, nil
}

// Parse verifies the signature, issuer and expiry of a token and returns its claims
func (j *JWT) Parse(tokenString string) (Claims, error) {
	var claims jwtClaims

	_, err := jwt.ParseWithClaims(tokenString, &claims,
		func(*jwt.Token) (any, error) { return j.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), // rejects "none" and algorithm swaps
		jwt.WithIssuer(j.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(j.now),
	)
	if err != nil {
		return Claims{}, errors.Join(ErrInvalidToken, err)
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return Claims{}, errors.Join(ErrInvalidToken, fmt.Errorf("invalid subject %q", claims.Subject))
	}

	if claims.ID == "" {
		return Claims{}, errors.Join(ErrInvalidToken, errors.New("missing token id"))
	}

	return Claims{ID: claims.ID, UserID: userID, Role: claims.Role, ExpiresAt: claims.ExpiresAt.Time}, nil
}

// newTokenID returns 128 random bits as hex
func newTokenID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("error generating token id: %w", err)
	}
	return hex.EncodeToString(b), nil
}
