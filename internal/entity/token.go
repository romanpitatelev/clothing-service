package entity

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidSigningMethod = errors.New("invalid signing method")
	ErrInvalidToken         = errors.New("invalid token")
	ErrTokenExpired         = errors.New("token is expired")
	ErrInvalidUUIDFormat    = errors.New("invalid uuid format")
)

type Claims struct {
	UserID UserID `json:"userId"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	jwt.RegisteredClaims
}

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
