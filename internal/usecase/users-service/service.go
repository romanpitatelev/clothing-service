package usersservice

import (
	"context"

	"github.com/romanpitatelev/clothing-service/internal/entity"
)

type usersStore interface {
	CreateUser(ctx context.Context, user entity.User) (entity.User, error)
	GetUser(ctx context.Context, userID entity.UserID) (entity.User, error)
	UpdateUser(ctx context.Context, userID entity.UserID, updatedUser entity.UserUpdate) (entity.User, error)
	DeleteUser(ctx context.Context, userID entity.User) error
	GetUsers(ctx context.Context, request entity.GetUsersRequest) ([]entity.User, error)
}

type Service struct {
	usersStore usersStore
}

func New(usersStore usersStore) *Service {
	return &Service{
		usersStore: usersStore,
	}
}

func (s *Service) CreateUser(ctx context.Context, user entity.User) (entity.User, error)  {}
func (s *Service) GetUser(ctx context.Context, userID entity.UserID) (entity.User, error) {}
func (s *Service) UpdateUser(ctx context.Context, userID entity.UserID, updatedUser entity.UserUpdate) (entity.User, error) {
}
func (s *Service) DeleteUser(ctx context.Context, userID entity.User) error {}
func (s *Service) GetUsers(ctx context.Context, request entity.GetUsersRequest) ([]entity.User, error) {
}
