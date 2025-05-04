package entity

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	phoneNumberLength      = 11
	phoneNumberLengthShort = 10
)

type UserID uuid.UUID //nolint:recvcheck

func (u UserID) String() string {
	return uuid.UUID(u).String()
}

type User struct {
	UserID        UserID     `json:"id"`
	FirstName     *string    `json:"firstName"`
	LastName      *string    `json:"lastName"`
	NickName      string     `json:"nickName"`
	Gender        *string    `json:"gender"`
	BirthDate     *time.Time `json:"birthDate"`
	Email         *string    `json:"email"`
	EmailVerified bool       `json:"emailVerified"`
	Phone         string     `json:"phone"`
	PhoneVerified bool       `json:"phoneVerified"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     *time.Time `json:"updatedAt"`
	DeletedAt     *time.Time `json:"deletedAt"`
	OTP           string     `json:"otp"`
	OTPCreatedAt  time.Time  `json:"-"`
}

type UserUpdate struct {
	FirstName     *string    `json:"firstName"`
	LastName      *string    `json:"lastName"`
	NickName      *string    `json:"nickName"`
	Gender        *string    `json:"gender"`
	Email         *string    `json:"email"`
	BirthDate     *time.Time `json:"birthDate"`
	EmailVerified *bool      `json:"-"`
	Phone         *string    `json:"phone"`
	PhoneVerified *bool      `json:"-"`
	OTP           *string    `json:"-"`
	OTPCreatedAt  *time.Time `json:"-"`
}

type Claims struct {
	UserID UserID `json:"userId"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	jwt.RegisteredClaims
}

var (
	ErrInvalidSigningMethod = errors.New("invalid signing method")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidPhone         = errors.New("invalid phone number")
	ErrInvalidUserIDFormat  = errors.New("invalid user id format")
	ErrInvalidToken         = errors.New("invalid token")
	ErrTokenExpired         = errors.New("token is expired")
	ErrInvalidUUIDFormat    = errors.New("invalid uuid format")
	ErrInvalidOTP           = errors.New("invalid otp")
	ErrDuplicateContact     = errors.New("duplicate contact")
)

type UserInfo struct {
	UserID UserID `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type Tokens struct {
	UserID       UserID `json:"-"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	Timeout      int    `json:"timeout"`
}

func validateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("email validation error: %w", err)
	}

	return nil
}

func validatePhone(phone string) (string, error) {
	re := regexp.MustCompile(`\D`)

	digits := re.ReplaceAllString(phone, "")

	switch {
	case len(digits) == phoneNumberLength && strings.HasPrefix(digits, "7"):
	case len(digits) == phoneNumberLength && strings.HasPrefix(digits, "8"):
		digits = "7" + digits[1:]
	case len(digits) == phoneNumberLengthShort:
		digits = "7" + digits
	default:
		return "", ErrInvalidPhone
	}

	if len(digits) != phoneNumberLength {
		return "", ErrInvalidPhone
	}

	if !containsOnlyDigits(digits) {
		return "", ErrInvalidPhone
	}

	return digits, nil
}

func containsOnlyDigits(s string) bool {
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

func (u *User) Validate() (User, error) {
	formattedPhone, err := validatePhone(u.Phone)
	if err != nil {
		return User{}, fmt.Errorf("phone validation error: %w", err)
	}

	u.Phone = formattedPhone

	if u.Email != nil {
		err := validateEmail(*u.Email)
		if err != nil {
			return User{}, fmt.Errorf("email validation error: %w", err)
		}
	}

	return *u, nil
}

func (uu *UserUpdate) Validate() (UserUpdate, error) {
	if uu.Email != nil {
		err := validateEmail(*uu.Email)
		if err != nil {
			return UserUpdate{}, fmt.Errorf("invalid email address: %w", err)
		}
	}

	if uu.Phone != nil {
		formattedPhone, err := validatePhone(*uu.Phone)
		if err != nil {
			return UserUpdate{}, fmt.Errorf("invalid phone number: %w", err)
		}

		*uu.Phone = formattedPhone
	}

	return *uu, nil
}

func unmarshalUUID(id *uuid.UUID, data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("unmarshalling error: %w", err)
	}

	parsed, err := uuid.Parse(s)
	if err != nil {
		return ErrInvalidUUIDFormat
	}

	*id = parsed

	return nil
}

func (u *UserID) UnmarshalText(data []byte) error {
	return unmarshalUUID((*uuid.UUID)(u), data)
}

func (u *UserID) MarshalText() ([]byte, error) {
	data, err := json.Marshal(uuid.UUID(*u).String())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal UUID: %w", err)
	}

	return data, nil
}

type ValidateUserRequest struct {
	UserID UserID `json:"userId"`
	OTP    string `json:"otp"`
}
