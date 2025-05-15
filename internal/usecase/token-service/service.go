package tokenservice

import (
	"context"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"

	"github.com/romanpitatelev/clothing-service/internal/entity"
)

const (
	accessTokenDurationSeconds = 300
	refreshTokenDuration       = 24 * 30 * time.Hour
)

type Config struct {
	OTPLifetime time.Duration
	PrivateKey  *rsa.PrivateKey
	PublicKey   *rsa.PublicKey
}

type usersStore interface {
	GetUser(ctx context.Context, userID entity.UserID) (entity.User, error)
	VerifyUserWithOTP(ctx context.Context, validateUserRequest entity.ValidateUserRequest, otpLifetime time.Duration) (entity.User, error)
}

type Service struct {
	cfg        Config
	usersStore usersStore
}

func New(cfg Config, usersStore usersStore) *Service {
	return &Service{
		cfg:        cfg,
		usersStore: usersStore,
	}
}

func (s *Service) GenerateAccessToken(user entity.User) (string, error) {
	claims := jwt.MapClaims{
		"userId": user.ID,
		"phone":  user.Phone,
		"exp":    time.Now().Add(accessTokenDurationSeconds * time.Second).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenStr, err := token.SignedString(s.cfg.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenStr, nil
}

func (s *Service) ParseRefreshToken(tokens entity.Tokens) (entity.UserID, error) {
	token, err := jwt.Parse(tokens.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, entity.ErrInvalidSigningMethod
		}

		return s.cfg.PublicKey, nil
	})
	if err != nil {
		return entity.UserID(uuid.Nil), fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr, ok := claims["userId"].(string)
		if !ok {
			return entity.UserID(uuid.Nil), fmt.Errorf("%w: userId is %v instead of string", err, claims["userId"])
		}

		userID, err := uuid.FromString(userIDStr)
		if err != nil {
			return entity.UserID(uuid.Nil), entity.ErrInvalidUserIDFormat
		}

		return entity.UserID(userID), nil
	}

	return entity.UserID(uuid.Nil), entity.ErrInvalidToken
}

func (s *Service) ParseToken(tokenStr string) (entity.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &entity.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, entity.ErrInvalidSigningMethod
		}

		return s.cfg.PublicKey, nil
	})
	if err != nil {
		return entity.Claims{}, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*entity.Claims)
	if !ok || !token.Valid {
		return entity.Claims{}, entity.ErrInvalidToken
	}

	if claims.ExpiresAt.Before(time.Now()) {
		return entity.Claims{}, entity.ErrTokenExpired
	}

	return *claims, nil
}

func (s *Service) ValidateUser(ctx context.Context, validateUserRequest entity.ValidateUserRequest) (entity.Tokens, error) {
	user, err := s.usersStore.VerifyUserWithOTP(ctx, validateUserRequest, s.cfg.OTPLifetime)
	if err != nil {
		return entity.Tokens{}, fmt.Errorf("verification failed: %w", err)
	}

	tokens, err := s.generateTokens(user)
	if err != nil {
		return entity.Tokens{}, fmt.Errorf("failed to generate tokens in ValidateUser(): %w", err)
	}

	return tokens, nil
}

func (s *Service) RefreshToken(ctx context.Context, tokens entity.Tokens) (entity.Tokens, error) {
	userID, err := s.ParseRefreshToken(tokens)
	if err != nil {
		return entity.Tokens{}, fmt.Errorf("%w: %w", entity.ErrInvalidToken, err)
	}

	user, err := s.usersStore.GetUser(ctx, userID)
	if err != nil {
		return entity.Tokens{}, fmt.Errorf("user not found in RefreshToken(): %w", err)
	}

	tokens, err = s.generateTokens(user)
	if err != nil {
		return entity.Tokens{}, fmt.Errorf("error creating access token: %w", err)
	}

	return tokens, nil
}

func (s *Service) generateTokens(user entity.User) (entity.Tokens, error) {
	accessToken, err := s.GenerateAccessToken(user)
	if err != nil {
		return entity.Tokens{}, err
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return entity.Tokens{}, err
	}

	return entity.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Timeout:      accessTokenDurationSeconds,
	}, nil
}

func (s *Service) generateRefreshToken(user entity.User) (string, error) {
	claims := jwt.MapClaims{
		"userId": user.ID,
		"exp":    time.Now().Add(refreshTokenDuration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenStr, err := token.SignedString(s.cfg.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenStr, nil
}
