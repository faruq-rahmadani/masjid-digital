package domain

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/faruq-rahmadani/masjid-digital/pkg/apperror"
)

// UserRole is what a user may do in the app
type UserRole string

const (
	UserRoleSuperAdmin UserRole = "super_admin" // manages the whole app, including users
	UserRoleTakmir     UserRole = "takmir"      // runs the mosque's operations: prayer times, slides, announcements
	UserRoleBendahara  UserRole = "bendahara"   // the treasurer: records kas income and expenses
)

// UserStatus tells whether a user can log in. Users are disabled, never
// deleted, because kas transactions keep referencing who recorded them.
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with the same phone or email already exists")
)

// User is a person who logs in to the admin panel with their phone number
type User struct {
	ID           int64
	Name         string
	Phone        string // E.164, e.g. +6281234567890
	Email        string // optional, empty when not set
	PasswordHash string
	Role         UserRole
	Status       UserStatus
	LastLoginAt  *time.Time
	CreatedBy    int64 // ID of the user who created this one, 0 when nobody did
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

var (
	ErrInvalidPhone = errors.New("invalid phone number")
	ErrInvalidEmail = errors.New("invalid email address")
)

// CreateUserRequest is the input for creating a user. Password is plain text;
// the usecase hashes it and never stores it.
type CreateUserRequest struct {
	Name     string
	Phone    string
	Email    string // optional
	Password string
	Role     UserRole // required, one of the UserRole constants

	// CreatedBy is the logged in user making the request. The controller sets it
	// from the session after binding, overwriting anything the client sent.
	CreatedBy int64 `json:"-"`
}

func (c *CreateUserRequest) ValidateAndNormalize() error {
	validations := []func() error{
		c.validateName,
		c.validatePhone,
		c.validateEmail,
		c.validatePassword,
		c.validateRole,
	}

	for _, fn := range validations {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func (c *CreateUserRequest) validatePhone() error {
	phone, err := validatePhone(c.Phone)
	c.Phone = phone
	return err
}

func (c *CreateUserRequest) validateEmail() error {
	email, err := validateEmail(c.Email)
	c.Email = email
	return err
}

func (c *CreateUserRequest) validatePassword() error {
	return validatePassword(c.Password)
}

func (c *CreateUserRequest) validateName() error {
	name, err := validateName(c.Name)
	c.Name = name
	return err
}

// validateRole rejects a missing or unknown role; there is no default, the admin picks one
func (c *CreateUserRequest) validateRole() error {
	return validateRole(c.Role)
}

func (c *CreateUserRequest) ToUser(passwordHash string) User {
	return User{
		Name:         c.Name,
		Phone:        c.Phone,
		Email:        c.Email,
		PasswordHash: passwordHash,
		Role:         c.Role,
		Status:       UserStatusActive,
		CreatedBy:    c.CreatedBy,
	}
}

// IsValid reports whether r is a known platform role
func (r UserRole) IsValid() bool {
	switch r {
	case UserRoleSuperAdmin, UserRoleTakmir, UserRoleBendahara:
		return true
	default:
		return false
	}
}

// IsValid reports whether s is a known user status
func (s UserStatus) IsValid() bool {
	return s == UserStatusActive || s == UserStatusDisabled
}

// UpdateUserRequest replaces the editable fields of a user (PUT): every field is
// sent, and an empty email clears it. Password changes have their own flow.
type UpdateUserRequest struct {
	// ID comes from the URL path; the controller sets it after binding
	ID int64 `json:"-"`

	Name   string     `json:"name"`
	Phone  string     `json:"phone"`
	Email  string     `json:"email"` // empty clears the email
	Role   UserRole   `json:"role"`
	Status UserStatus `json:"status"`
}

func (r *UpdateUserRequest) ValidateAndNormalize() error {
	var err error

	if r.Name, err = validateName(r.Name); err != nil {
		return err
	}
	if r.Phone, err = validatePhone(r.Phone); err != nil {
		return err
	}
	if r.Email, err = validateEmail(r.Email); err != nil {
		return err
	}
	if err := validateRole(r.Role); err != nil {
		return err
	}
	if !r.Status.IsValid() {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("unknown status %q", r.Status), "status tidak valid")
	}

	return nil
}

// ApplyTo returns user with the request's fields replacing its editable ones
func (r *UpdateUserRequest) ApplyTo(user User) User {
	user.Name = r.Name
	user.Phone = r.Phone
	user.Email = r.Email
	user.Role = r.Role
	user.Status = r.Status
	return user
}

// ChangePasswordRequest is a logged in user changing their own password
type ChangePasswordRequest struct {
	// UserID is the logged in user; the controller sets it from the session after binding
	UserID int64 `json:"-"`

	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (r *ChangePasswordRequest) Validate() error {
	if r.CurrentPassword == "" {
		return apperror.New(apperror.CodeInvalidInput, errors.New("missing current password"), "password lama wajib di isi")
	}

	return validatePassword(r.NewPassword)
}

// ResetPasswordRequest is an admin setting another user's password, without knowing the old one
type ResetPasswordRequest struct {
	// UserID is the user whose password is reset; the controller sets it from the URL path
	UserID int64 `json:"-"`

	NewPassword string `json:"new_password"`
}

func (r *ResetPasswordRequest) Validate() error {
	return validatePassword(r.NewPassword)
}

// validateName trims name and checks it is present and not too long
func validateName(name string) (string, error) {
	const maxNameLength = 100

	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return name, apperror.New(apperror.CodeInvalidInput, fmt.Errorf("missing name (%s)", name), "nama wajib di isi")
	}

	if utf8.RuneCountInString(trimmed) > maxNameLength {
		return name, apperror.New(apperror.CodeInvalidInput, fmt.Errorf("name must be less than %d characters long", maxNameLength), fmt.Sprintf("nama maksimal %d karakter", maxNameLength))
	}

	return trimmed, nil
}

// validatePhone normalizes phone to E.164
func validatePhone(phone string) (string, error) {
	normalized, err := NormalizePhone(phone)
	if err != nil {
		return phone, apperror.New(apperror.CodeInvalidInput, err, "nomor HP tidak valid")
	}
	return normalized, nil
}

// validateEmail normalizes an optional email to lowercase
func validateEmail(email string) (string, error) {
	normalized, err := NormalizeEmail(email)
	if err != nil {
		return email, apperror.New(apperror.CodeInvalidInput, err, "email tidak valid")
	}
	return normalized, nil
}

// validatePassword checks a new password is between minPasswordLength and maxPasswordLength
func validatePassword(password string) error {
	if len(password) < minPasswordLength {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("password must be at least %d characters long", minPasswordLength), fmt.Sprintf("password minimal %d karakter", minPasswordLength))
	}

	if len(password) > maxPasswordLength {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("password must be at most %d characters long", maxPasswordLength), fmt.Sprintf("password maksimal %d karakter", maxPasswordLength))
	}

	return nil
}

// validateRole rejects a missing or unknown role
func validateRole(role UserRole) error {
	if role == "" {
		return apperror.New(apperror.CodeInvalidInput, errors.New("missing role"), "role wajib di isi")
	}

	if !role.IsValid() {
		return apperror.New(apperror.CodeInvalidInput, fmt.Errorf("unknown role %q", role), "role tidak valid")
	}
	return nil
}

var (
	// phoneSeparators are stripped from a phone number before validating it
	phoneSeparators = strings.NewReplacer(" ", "", "-", "", ".", "", "(", "", ")", "")
	// e164 is a + followed by a country code and up to 15 digits in total
	e164 = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
	// indonesianMobile is +62 8xx with 9 to 12 digits after the country code
	indonesianMobile = regexp.MustCompile(`^\+628\d{8,11}$`)
)

// NormalizePhone converts a phone number to E.164. Numbers without a + are
// read as Indonesian: 0812-3456-7890 and 6281234567890 both become +6281234567890.
func NormalizePhone(phone string) (string, error) {
	p := phoneSeparators.Replace(strings.TrimSpace(phone))

	switch {
	case strings.HasPrefix(p, "+"):
	case strings.HasPrefix(p, "0"):
		p = "+62" + p[1:]
	case strings.HasPrefix(p, "62"):
		p = "+" + p
	default:
		return "", ErrInvalidPhone
	}

	if !e164.MatchString(p) {
		return "", ErrInvalidPhone
	}
	if strings.HasPrefix(p, "+62") && !indonesianMobile.MatchString(p) {
		return "", ErrInvalidPhone
	}

	return p, nil
}

// NormalizeEmail trims and lowercases an email, returning "" for an empty one
func NormalizeEmail(email string) (string, error) {
	e := strings.ToLower(strings.TrimSpace(email))
	if e == "" {
		return "", nil
	}

	addr, err := mail.ParseAddress(e)
	if err != nil || addr.Address != e {
		return "", ErrInvalidEmail
	}

	return e, nil
}
