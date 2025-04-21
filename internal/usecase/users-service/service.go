package usersservice

import (
	"context"
	"fmt"

	"github.com/romanpitatelev/clothing-service/internal/entity"
)

type usersStore interface {
	CreateUser(ctx context.Context, user entity.User) (entity.User, error)
	GetUser(ctx context.Context, userID entity.UserID) (entity.User, error)
	UpdateUser(ctx context.Context, userID entity.UserID, updatedUser entity.UserUpdate) (entity.User, error)
	DeleteUser(ctx context.Context, userID entity.UserID) error
	GetUsers(ctx context.Context, request entity.GetRequestParams) ([]entity.User, error)
}

type Service struct {
	usersStore usersStore
}

func New(usersStore usersStore) *Service {
	return &Service{
		usersStore: usersStore,
	}
}

func (s *Service) CreateUser(ctx context.Context, user entity.User) (entity.User, error) {
}

func (s *Service) GetUser(ctx context.Context, userID entity.UserID) (entity.User, error) {
	user, err := s.usersStore.GetUser(ctx, userID)
	if err != nil {
		return entity.User{}, fmt.Errorf("failed to get wallet: %w", err)
	}

	return user, nil
}

func (s *Service) UpdateUser(ctx context.Context, userID entity.UserID, newInfoUser entity.UserUpdate) (entity.User, error) {
	dbUser, err := s.usersStore.GetUser(ctx, userID)
	if err != nil {
		return entity.User{}, fmt.Errorf("user not found: %w", err)
	}

	if newInfoUser.FirstName == "" {
		newInfoUser.FirstName = dbUser.FirstName
	}

	if newInfoUser.LastName == "" {
		newInfoUser.LastName = dbUser.LastName
	}

	if newInfoUser.NickName == "" {
		newInfoUser.NickName = dbUser.NickName
	}

	if string(newInfoUser.Email) == "" {
		newInfoUser.Email = dbUser.Email
	} else {
		err = newInfoUser.Email.Validate()
		if err != nil {
			return entity.User{}, fmt.Errorf("invalid email address: %w", err)
		}
	}

	updatedUser, err := s.usersStore.UpdateUser(ctx, userID, newInfoUser)
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

func (s *Service) GetUsers(ctx context.Context, request entity.GetRequestParams) ([]entity.User, error) {
	users, err := s.usersStore.GetUsers(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error getting users: %w", err)
	}

	return users, nil
}
