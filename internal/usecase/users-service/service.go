package usersservice

import (
	"context"
	"fmt"
	"time"

	"github.com/romanpitatelev/clothing-service/internal/entity"
	smsregistration "github.com/romanpitatelev/clothing-service/internal/sms-registration"
)

type usersStore interface {
	CreateUnverifiedUser(ctx context.Context, user entity.User) (entity.User, error)
	VerifyUser(ctx context.Context, unverifiedUserID entity.UserID) (entity.User, error)
	GetUser(ctx context.Context, userID entity.UserID) (entity.User, error)
	UpdateUser(ctx context.Context, userID entity.UserID, updatedUser entity.UserUpdate) (entity.User, error)
	DeleteUser(ctx context.Context, userID entity.UserID) error
}

type smsRegistration interface {
	SendVerificationCode(ctx context.Context, phoneNumber string) (*smsregistration.VerificationResponse, error)
	CheckVerificationCode(ctx context.Context, check smsregistration.VerificationCheck) (*smsregistration.VerificationResult, error)
}

type Service struct {
	usersStore      usersStore
	smsRegistration smsRegistration
}

func New(usersStore usersStore, smssmsRegistration smsRegistration) *Service {
	return &Service{
		usersStore:      usersStore,
		smsRegistration: smssmsRegistration,
	}
}

func (s *Service) CreateUser(ctx context.Context, user entity.User) (entity.User, error) {
	verificationResp, err := s.smsRegistration.SendVerificationCode(ctx, user.Phone)
	if err != nil {
		return entity.User{}, fmt.Errorf("failed to send verification code: %w", err)
	}

	validatedUser, err := user.Validate()
	if err != nil {
		return entity.User{}, fmt.Errorf("user validation failed: %w", err)
	}

	validatedUser.CreatedAt = time.Now()

	unverifiedUser, err := s.usersStore.CreateUnverifiedUser(ctx, validatedUser)
	if err != nil {
		return entity.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	result, err := s.smsRegistration.CheckVerificationCode(ctx, smsregistration.VerificationCheck{
		Phone: unverifiedUser.Phone,
		Code:  verificationResp.Code,
	})
	if err != nil {
		return entity.User{}, fmt.Errorf("verification failed: %w", err)
	}

	if !result.Success {
		return entity.User{}, fmt.Errorf("invalid verification code: %w", err)
	}

	verifiedUser, err := s.usersStore.VerifyUser(ctx, unverifiedUser.UserID)
	if err != nil {
		return entity.User{}, fmt.Errorf("failed to verify user: %w", err)
	}

	return verifiedUser, nil
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
