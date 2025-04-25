package entity

import (
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

type (
	UserID uuid.UUID
)

type User struct {
	UserID     UserID     `json:"id" db:"id"`
	FirstName  string     `json:"firstName"`
	LastName   string     `json:"lastName"`
	NickName   string     `json:"nickName"`
	Gender     string     `json:"gender"`
	Age        int        `json:"age"`
	Email      string     `json:"email"`
	Phone      string     `json:"phone"`
	CreatedAt  time.Time  `json:"createdAt"`
	IsVerified bool       `json:"isVerified"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	DeletedAt  *time.Time `json:"deletedAt"`
}

type UserUpdate struct {
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	NickName  *string `json:"nickName"`
	Email     *string `json:"email"`
	Phone     *string `json:"phone"`
}

type ClothingItem struct{}

type GetRequestParams struct {
	Sorting    string `json:"sorting,omitempty"`
	Descending bool   `json:"descending,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Filter     string `json:"filter,omitempty"`
	Offset     int    `json:"offset,omitempty"`
}

var (
	ErrInvalidToken         = errors.New("invalid token")
	ErrInvalidSigningMethod = errors.New("invalid signing method")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidPhone         = errors.New("invalid phone number")
	ErrInvalidCode          = errors.New("invalid code")
)

type Claims struct {
	UserID UserID `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type UserInfo struct {
	UserID UserID `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

func validateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("email validation error: %w", err)
	}

	return nil
}

func validatePhone(phone string) (string, error) {
	re, err := regexp.Compile(`\D`)
	if err != nil {
		return "", fmt.Errorf("failed to compile regexp: %w", err)
	}

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

	err = validateEmail(u.Email)
	if err != nil {
		return User{}, fmt.Errorf("email validation error: %w", err)
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
