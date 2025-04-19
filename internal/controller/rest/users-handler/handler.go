package usershandler

import (
	"context"
	"net/http"

	"github.com/romanpitatelev/clothing-service/internal/entity"
)

type usersService interface {
	CreateUser(ctx context.Context, user entity.User) (entity.User, error)
	GetUser(ctx context.Context, userID entity.UserID) (entity.User, error)
	UpdateUser(ctx context.Context, userID entity.UserID, updatedUser entity.UserUpdate) (entity.User, error)
	DeleteUser(ctx context.Context, userID entity.User) error
	GetUsers(ctx context.Context, request entity.GetUsersRequest) ([]entity.User, error)
}

type Handler struct {
	usersService usersService
}

func New(usersService usersService) *Handler {
	return &Handler{
		usersService: usersService,
	}
}

func (h *Handler) CreateUser(w http.ResponseWriter, t *http.Request) {}
func (h *Handler) GetUser(w http.ResponseWriter, t *http.Request)    {}
func (h *Handler) UpdateUser(w http.ResponseWriter, t *http.Request) {}
func (h *Handler) DeleteUser(w http.ResponseWriter, t *http.Request) {}
func (h *Handler) GetUsers(w http.ResponseWriter, t *http.Request)   {}
