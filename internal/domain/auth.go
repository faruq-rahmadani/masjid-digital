package domain

import (
	"errors"
	"strings"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 128 // bounds the work of hashing a huge input
)

// IdentifierMethod is the method used to identify a user, e.g. phone number or email
type IdentifierMethod string

const (
	IdentifierMethodPhoneNumber IdentifierMethod = "phone"
	IdentifierMethodEmail       IdentifierMethod = "email"
)

type LoginRequest struct {
	Identifier string `json:"identifier"` // phone number or email
	Password   string `json:"password"`
}

func (r *LoginRequest) IsPasswordLengthValid() bool {
	return len(r.Password) >= minPasswordLength && len(r.Password) <= maxPasswordLength
}

// IdentifierMethod will return the method used to identify the user
func (r *LoginRequest) IdentifierMethod() IdentifierMethod {
	if strings.Contains(r.Identifier, "@") {
		return IdentifierMethodEmail
	}

	return IdentifierMethodPhoneNumber
}

type LoginResponse struct {
	Token    string `json:"token"`
	ExpireIn int64  `json:"expire_in"` // in seconds
}

// ErrSessionNotFound means the session was logged out or has expired
var ErrSessionNotFound = errors.New("session not found")

// AuthSession is the logged in user behind a request
type AuthSession struct {
	ID     string // session ID, the token's jti
	UserID int64
	Role   UserRole
}
