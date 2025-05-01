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
	CreateUser(ctx context.Context, user entity.User) error
	ValidateUser(ctx context.Context, user entity.User) (entity.Tokens, error)
	LoginUser(ctx context.Context, userID entity.UserID) (entity.Tokens, error)
	RefreshToken(ctx context.Context, tokens entity.Tokens) (entity.Tokens, error)
	GetUser(ctx context.Context, userID entity.UserID) (entity.User, error)
	UpdateUser(ctx context.Context, userID entity.UserID, updatedUser entity.UserUpdate) (entity.User, error)
	DeleteUser(ctx context.Context, userID entity.UserID) error
}

type Handler struct {
	usersService usersService
}

func New(usersService usersService) *Handler {
	return &Handler{
		usersService: usersService,
	}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user entity.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)

		return
	}

	ctx := r.Context()

	if err := h.usersService.CreateUser(ctx, user); err != nil {
		common.ErrorResponse(w, "error creating user", err)

		return
	}

	common.OkResponse(w, http.StatusOK, nil)
}

func (h *Handler) ValidateUser(w http.ResponseWriter, r *http.Request) {
	var user entity.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)

		return
	}

	ctx := r.Context()

	tokens, err := h.usersService.ValidateUser(ctx, user)
	if err != nil {
		common.ErrorResponse(w, "error validating user", err)

		return
	}

	common.OkResponse(w, http.StatusOK, tokens)
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var userID entity.UserID

	if err := json.NewDecoder(r.Body).Decode(&userID); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)

		return
	}

	ctx := r.Context()

	tokens, err := h.usersService.LoginUser(ctx, userID)
	if err != nil {
		common.ErrorResponse(w, "user login error", err)

		return
	}

	common.OkResponse(w, http.StatusOK, tokens)
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var tokens entity.Tokens

	if err := json.NewDecoder(r.Body).Decode(&tokens); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)

		return
	}

	ctx := r.Context()

	newTokens, err := h.usersService.RefreshToken(ctx, tokens)
	if err != nil {
		common.ErrorResponse(w, "error refreshing token", err)

		return
	}

	common.OkResponse(w, http.StatusOK, newTokens)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "failed to parse user id", http.StatusBadRequest)

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
	userIDStr := chi.URLParam(r, "id")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "failed to parse user id", http.StatusBadRequest)

		return
	}

	ctx := r.Context()

	var user entity.UserUpdate

	if err = json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)

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
	userIDStr := chi.URLParam(r, "id")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "failed to parse user id", http.StatusBadRequest)

		return
	}

	ctx := r.Context()

	if err = h.usersService.DeleteUser(ctx, entity.UserID(userID)); err != nil {
		common.ErrorResponse(w, "error deleting user", err)

		return
	}

	common.OkResponse(w, http.StatusNoContent, "user deleted successfully")
}
