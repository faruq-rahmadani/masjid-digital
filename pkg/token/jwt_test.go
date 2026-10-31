package token

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "0123456789abcdef0123456789abcdef"

func newTestJWT(t *testing.T) *JWT {
	t.Helper()
	j, err := NewJWT(testSecret, time.Hour, "masjid-digital")
	if err != nil {
		t.Fatal(err)
	}
	return j
}

func TestIssueAndParse(t *testing.T) {
	j := newTestJWT(t)

	issued, err := j.Issue(42, "super_admin")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if issued.ExpiresIn != time.Hour {
		t.Errorf("ttl = %s, want 1h", issued.ExpiresIn)
	}
	if len(issued.ID) != 32 {
		t.Errorf("id = %q, want 32 hex characters", issued.ID)
	}

	claims, err := j.Parse(issued.Value)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.ID != issued.ID || claims.UserID != 42 || claims.Role != "super_admin" {
		t.Errorf("claims = %+v, want id %s, user 42 super_admin", claims, issued.ID)
	}

	other, _ := j.Issue(42, "super_admin")
	if other.ID == issued.ID {
		t.Error("two tokens got the same id")
	}
	if d := time.Until(claims.ExpiresAt); d <= 59*time.Minute || d > time.Hour {
		t.Errorf("expires in %s, want about 1h", d)
	}
}

func TestParseRejects(t *testing.T) {
	j := newTestJWT(t)
	valid, _ := j.Issue(42, "user")

	expired := newTestJWT(t)
	expired.now = func() time.Time { return time.Now().Add(-2 * time.Hour) }
	expiredToken, _ := expired.Issue(42, "user")

	otherSecret, _ := NewJWT(strings.Repeat("x", MinSecretLength), time.Hour, "masjid-digital")
	forged, _ := otherSecret.Issue(42, "super_admin")

	otherIssuer, _ := NewJWT(testSecret, time.Hour, "someone-else")
	wrongIssuer, _ := otherIssuer.Issue(42, "user")

	withoutID, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: "42", Issuer: "masjid-digital", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	}).SignedString([]byte(testSecret))

	none, _ := jwt.NewWithClaims(jwt.SigningMethodNone, jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: "42", Issuer: "masjid-digital", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)

	tests := map[string]string{
		"empty":        "",
		"garbage":      "not-a-token",
		"tampered":     valid.Value[:len(valid.Value)-2] + "xx",
		"expired":      expiredToken.Value,
		"other secret": forged.Value,
		"wrong issuer": wrongIssuer.Value,
		"alg none":     none,
		"without id":   withoutID,
	}

	for name, tokenString := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := j.Parse(tokenString); !errors.Is(err, ErrInvalidToken) {
				t.Errorf("Parse err = %v, want ErrInvalidToken", err)
			}
		})
	}
}

func TestNewJWTValidatesConfig(t *testing.T) {
	if _, err := NewJWT("too-short", time.Hour, "x"); err == nil {
		t.Error("short secret accepted")
	}
	if _, err := NewJWT(testSecret, 0, "x"); err == nil {
		t.Error("zero ttl accepted")
	}
}
