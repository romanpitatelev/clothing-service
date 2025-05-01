package usersservice

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/romanpitatelev/clothing-service/internal/entity"
)

const (
	otpLength   = 4
	otpDuration = 5 * time.Minute
	ten         = 10
)

type usersStore interface {
	CreateUnverifiedUser(ctx context.Context, user entity.User, otp string, otpExpiresAt time.Time) error
	VerifyUserWithOTP(ctx context.Context, unverifiedUser entity.User) error
	GetUser(ctx context.Context, userID entity.UserID) (entity.User, error)
	UpdateUser(ctx context.Context, userID entity.UserID, updatedUser entity.UserUpdate) (entity.User, error)
	DeleteUser(ctx context.Context, userID entity.UserID) error
}

type smsRegistration interface {
	SendOTP(ctx context.Context, phone string, otp string) error
}

type tokenGenerator interface {
	GenerateTokens(user entity.User) (entity.Tokens, error)
	GenerateAccessToken(user entity.User) (string, error)
	GenerateAccessTokenTimeout() time.Time
	ParseRefreshToken(tokens entity.Tokens) (entity.UserID, error)
}

type Service struct {
	usersStore      usersStore
	smsRegistration smsRegistration
	tokenGenerator  tokenGenerator
}

func New(usersStore usersStore, smssmsRegistration smsRegistration, tokentokenGenerator tokenGenerator) *Service {
	return &Service{
		usersStore:      usersStore,
		smsRegistration: smssmsRegistration,
		tokenGenerator:  tokentokenGenerator,
	}
}

func (s *Service) CreateUser(ctx context.Context, user entity.User) error {
	validatedUser, err := user.Validate()
	if err != nil {
		return fmt.Errorf("user validation failed: %w", err)
	}

	otp := generateOTP()
	otpExpiresAt := time.Now().Add(otpDuration)

	if err = s.usersStore.CreateUnverifiedUser(ctx, validatedUser, otp, otpExpiresAt); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	if err = s.smsRegistration.SendOTP(ctx, *user.Phone, user.OTP); err != nil {
		return fmt.Errorf("failed to send otp: %w", err)
	}

	return nil
}

func (s *Service) ValidateUser(ctx context.Context, user entity.User) (entity.Tokens, error) {
	err := s.usersStore.VerifyUserWithOTP(ctx, user)
	if err != nil {
		return entity.Tokens{}, fmt.Errorf("verification failed: %w", err)
	}

	tokens, err := s.tokenGenerator.GenerateTokens(user)
	if err != nil {
		return entity.Tokens{}, fmt.Errorf("failed to generate tokens in ValidateUser(): %w", err)
	}

	return tokens, nil
}

func (s *Service) LoginUser(ctx context.Context, userID entity.UserID) (entity.Tokens, error) {
	user, err := s.usersStore.GetUser(ctx, userID)
	if err != nil {
		return entity.Tokens{}, fmt.Errorf("failed to get user in LoginUser(): %w", err)
	}

	var tokens entity.Tokens

	if user.IsVerified {
		tokens, err = s.tokenGenerator.GenerateTokens(user)
		if err != nil {
			return entity.Tokens{}, fmt.Errorf("failed to generate tokens in LoginUser(): %w", err)
		}
	}

	return tokens, nil
}

func (s *Service) RefreshToken(ctx context.Context, tokens entity.Tokens) (entity.Tokens, error) {
	if tokens.Timeout.Unix() > time.Now().Unix() {
		return entity.Tokens{}, entity.ErrAccessTokenExpired
	}

	userID, err := s.tokenGenerator.ParseRefreshToken(tokens)
	if err != nil {
		return entity.Tokens{}, fmt.Errorf("invalid refresh token: %w", err)
	}

	user, err := s.usersStore.GetUser(ctx, userID)
	if err != nil {
		return entity.Tokens{}, fmt.Errorf("user not found in RefreshToken(): %w", err)
	}

	if !user.IsVerified {
		return entity.Tokens{}, entity.ErrUserNotVerified
	}

	accessToken, err := s.tokenGenerator.GenerateAccessToken(user)
	if err != nil {
		return entity.Tokens{}, fmt.Errorf("error creating access token: %w", err)
	}

	return entity.Tokens{
		RefreshToken: tokens.RefreshToken,
		AccessToken:  accessToken,
		Timeout:      s.tokenGenerator.GenerateAccessTokenTimeout(),
	}, nil
}

func (s *Service) GetUser(ctx context.Context, userID entity.UserID) (entity.User, error) {
	user, err := s.usersStore.GetUser(ctx, userID)
	if err != nil {
		return entity.User{}, fmt.Errorf("failed to get wallet: %w", err)
	}

	return user, nil
}

func (s *Service) UpdateUser(ctx context.Context, userID entity.UserID, newInfoUser entity.UserUpdate) (entity.User, error) {
	_, err := s.usersStore.GetUser(ctx, userID)
	if err != nil {
		return entity.User{}, fmt.Errorf("user not found: %w", err)
	}

	newInfoUserValidated, err := newInfoUser.Validate()
	if err != nil {
		return entity.User{}, fmt.Errorf("new info validation failed: %w", err)
	}

	updatedUser, err := s.usersStore.UpdateUser(ctx, userID, newInfoUserValidated)
	if err != nil {
		return entity.User{}, fmt.Errorf("failed to update user info: %w", err)
	}

	return updatedUser, nil
}

func (s *Service) DeleteUser(ctx context.Context, userID entity.UserID) error {
	if err := s.usersStore.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func generateOTP() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	code := ""

	for range otpLength {
		code += strconv.Itoa(r.Intn(ten))
	}

	return code
}
