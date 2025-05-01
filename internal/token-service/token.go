package tokenservice

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/romanpitatelev/clothing-service/internal/entity"
)

const (
	accessTokenDuration  = 5 * time.Minute
	refreshTokenDuration = 24 * 30 * time.Hour
)

var (
	ErrInvalidUserIDFormat = errors.New("invalid user id format")
	ErrInvalidToken        = errors.New("invalid token")
)

type Generator struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func New(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) *Generator {
	return &Generator{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

func (g *Generator) GenerateTokens(user entity.User) (entity.Tokens, error) {
	accessToken, err := g.GenerateAccessToken(user)
	if err != nil {
		return entity.Tokens{}, err
	}

	refreshToken, err := g.generateRefreshToken(user)
	if err != nil {
		return entity.Tokens{}, err
	}

	return entity.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Timeout:      g.GenerateAccessTokenTimeout(),
	}, nil
}

func (g *Generator) GenerateAccessToken(user entity.User) (string, error) {
	claims := jwt.MapClaims{
		"userId": user.UserID,
		"phone":  *user.Phone,
		"exp":    g.GenerateAccessTokenTimeout(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenStr, err := token.SignedString(g.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenStr, nil
}

func (g *Generator) ParseRefreshToken(tokens entity.Tokens) (entity.UserID, error) {
	token, err := jwt.Parse(tokens.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, entity.ErrInvalidSigningMethod
		}

		return g.publicKey, nil
	})
	if err != nil {
		return entity.UserID(uuid.Nil), fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr, ok := claims["userId"].(string)
		if !ok {
			return entity.UserID(uuid.Nil), fmt.Errorf("invalid user ID token: %w", err)
		}

		userID, err := uuid.FromString(userIDStr)
		if err != nil {
			return entity.UserID(uuid.Nil), ErrInvalidUserIDFormat
		}

		return entity.UserID(userID), nil
	}

	return entity.UserID(uuid.Nil), ErrInvalidToken
}

func (g *Generator) GenerateAccessTokenTimeout() time.Time {
	return time.Now().Add(accessTokenDuration)
}

func (g *Generator) generateRefreshToken(user entity.User) (string, error) {
	claims := jwt.MapClaims{
		"userId": user.UserID,
		"exp":    time.Now().Add(refreshTokenDuration),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenStr, err := token.SignedString(g.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenStr, nil
}
