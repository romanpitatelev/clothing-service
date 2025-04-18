package entity

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type (
	UserID uuid.UUID
)

type User struct {
	UserID                UserID     `json:"userId"`
	FirstName             string     `json:"firstName"`
	LastName              string     `json:"lastName"`
	NickName              string     `json:"nickName"`
	Gender                string     `json:"gender"`
	Age                   int        `json:"age"`
	Email                 string     `json:"email"`
	Country               string     `json:"country"`
	LowerBodyClothingSize int        `json:"lowerBodyClothingSize"`
	UpperBodyClothingSize int        `json:"upperBodyClothingSize"`
	FootwearSize          int        `json:"footwearSize"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             time.Time  `json:"updatedAt"`
	DeletedAt             *time.Time `json:"deletedAt"`
	Active                bool       `json:"active"`
}

var (
	ErrInvalidToken         = errors.New("invalid token")
	ErrInvalidSigningMethod = errors.New("invalid signing method")
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
