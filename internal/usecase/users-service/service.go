package usersservice

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/romanpitatelev/clothing-service/internal/entity"
	"github.com/romanpitatelev/clothing-service/internal/utils"
)

type usersStore interface {
	CreateUnverifiedUser(ctx context.Context, user entity.User, otp string) (entity.User, error)
	GetUser(ctx context.Context, userID entity.UserID) (entity.User, error)
	UpdateUser(ctx context.Context, userID entity.UserID, updatedUser entity.UserUpdate) (entity.User, error)
	DeleteUser(ctx context.Context, userID entity.UserID) error
}

type smsRegistration interface {
	SendOTP(ctx context.Context, phone string, otp string) error
}

type Config struct {
	OTPMaxValue int
}

type Service struct {
	cfg             Config
	usersStore      usersStore
	smsRegistration smsRegistration
}

func New(cfg Config, usersStore usersStore, smsRegistration smsRegistration) *Service {
	return &Service{
		cfg:             cfg,
		usersStore:      usersStore,
		smsRegistration: smsRegistration,
	}
}

func (s *Service) CreateUser(ctx context.Context, user entity.User) (entity.User, error) {
	validatedUser, err := user.Validate()
	if err != nil {
		return entity.User{}, fmt.Errorf("user validation failed: %w", err)
	}

	otp := s.generateOTP()
	validatedUser.ID = entity.UserID(uuid.New())

	if user, err = s.usersStore.CreateUnverifiedUser(ctx, validatedUser, otp); err != nil {
		return entity.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	if err = s.smsRegistration.SendOTP(ctx, user.Phone, otp); err != nil {
		return entity.User{}, fmt.Errorf("failed to send otp: %w", err)
	}

	return user, nil
}

func (s *Service) LoginUser(ctx context.Context, userID entity.UserID) error {
	otp := s.generateOTP()

	user, err := s.usersStore.UpdateUser(ctx, userID, entity.UserUpdate{
		OTP:          &otp,
		OTPCreatedAt: utils.Pointer(time.Now()),
	})
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if err = s.smsRegistration.SendOTP(ctx, user.Phone, otp); err != nil {
		return fmt.Errorf("failed to send otp: %w", err)
	}

	return nil
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

func (s *Service) generateOTP() string {
	randomInt, _ := rand.Int(rand.Reader, big.NewInt(int64(s.cfg.OTPMaxValue)))

	res := randomInt.String()

	for {
		if len(res) == len(strconv.Itoa(s.cfg.OTPMaxValue)) {
			return res
		}

		res = "0" + res
	}
}
