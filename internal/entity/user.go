package entity

import (
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type (
	UserID uuid.UUID
	Email  string
)

type User struct {
	UserID    UserID     `json:"userId"`
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	NickName  string     `json:"nickName"`
	Gender    string     `json:"gender"`
	Age       int        `json:"age"`
	Email     Email      `json:"email"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}

type UserUpdate struct {
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	NickName  *string `json:"nickName"`
	Email     *Email  `json:"email"`
}

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

func (e *Email) Validate() error {
	emailStr := string(*e)
	_, err := mail.ParseAddress(emailStr)
	if err != nil {
		return fmt.Errorf("email validation error: %w", err)
	}

	return nil
}
