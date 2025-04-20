package usershandler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/romanpitatelev/clothing-service/internal/controller/rest/common"
	"github.com/romanpitatelev/clothing-service/internal/entity"
)

type usersService interface {
	CreateUser(ctx context.Context, user entity.User) (entity.User, error)
	GetUser(ctx context.Context, userID entity.UserID) (entity.User, error)
	UpdateUser(ctx context.Context, userID entity.UserID, updatedUser entity.UserUpdate) (entity.User, error)
	DeleteUser(ctx context.Context, userID entity.UserID) error
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

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		common.ErrorResponse(w, "error parsing uuid", err)

		return
	}

	ctx := r.Context()
	userInfo, err := h.usersService.GetUser(ctx, entity.UserID(userID))
	if err != nil {
		common.ErrorResponse(w, "error getting user", err)

		return
	}

	common.OkResponse(w, http.StatusOK, userInfo)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		common.ErrorResponse(w, "error parsing uuid", err)

		return
	}

	ctx := r.Context()

	var user entity.UserUpdate

	if err = json.NewDecoder(r.Body).Decode(&user); err != nil {
		common.ErrorResponse(w, "error decoding request body", err)

		return
	}

	updatedUser, err := h.usersService.UpdateUser(ctx, entity.UserID(userID), user)
	if err != nil {
		common.ErrorResponse(w, "error updating user", err)

		return
	}

	common.OkResponse(w, http.StatusOK, updatedUser)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		common.ErrorResponse(w, "error parsing uuid", err)

		return
	}

	ctx := r.Context()

	if err = h.usersService.DeleteUser(ctx, entity.UserID(userID)); err != nil {
		common.ErrorResponse(w, "error deleting user", err)

		return
	}

	common.OkResponse(w, http.StatusNoContent, "user deleted successfully")

}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {}
