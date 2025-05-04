package entity

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	phoneNumberLength      = 11
	phoneNumberLengthShort = 10
)

type User struct {
	ID            UserID     `json:"id" db:"id"`
	FirstName     *string    `json:"firstName" db:"first_name"`
	LastName      *string    `json:"lastName" db:"last_name"`
	NickName      string     `json:"nickName" db:"nick_name"`
	Gender        *string    `json:"gender" db:"gender"`
	BirthDate     *time.Time `json:"birthDate" db:"birth_date"`
	Email         *string    `json:"email" db:"email"`
	EmailVerified bool       `json:"emailVerified" db:"email_verified"`
	Phone         string     `json:"phone" db:"phone"`
	PhoneVerified bool       `json:"phoneVerified" db:"phone_verified"`
	CreatedAt     time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt     *time.Time `json:"updatedAt" db:"updated_at"`
	OTP           string     `json:"otp" db:"otp"`
	OTPCreatedAt  time.Time  `json:"-" db:"otp_created_at"`
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

type ValidateUserRequest struct {
	UserID UserID `json:"userId"`
	OTP    string `json:"otp"`
}

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidPhone        = errors.New("invalid phone number")
	ErrInvalidUserIDFormat = errors.New("invalid user id format")
	ErrInvalidOTP          = errors.New("invalid otp")
	ErrDuplicateContact    = errors.New("duplicate contact")
)

func validateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("email validation error: %w", err)
	}

	return nil
}

var phoneValidatorRE = regexp.MustCompile(`\D`)

func validatePhone(phone string) (string, error) {
	digits := phoneValidatorRE.ReplaceAllString(phone, "")

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

type UserID uuid.UUID

func (u UserID) String() string {
	return uuid.UUID(u).String()
}

func (u *UserID) UnmarshalText(data []byte) error {
	return (*uuid.UUID)(u).UnmarshalText(data)
}

func (u UserID) MarshalText() ([]byte, error) {
	return uuid.UUID(u).MarshalText()
}
