package iamhandler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/romanpitatelev/clothing-service/internal/controller/rest/common"
	"github.com/romanpitatelev/clothing-service/internal/entity"
)

type tokenService interface {
	ValidateUser(ctx context.Context, user entity.ValidateUserRequest) (entity.Tokens, error)
	RefreshToken(ctx context.Context, tokens entity.Tokens) (entity.Tokens, error)
	ParseToken(tokenStr string) (entity.Claims, error)
}

type Handler struct {
	tokenService tokenService
}

func New(tokenService tokenService) *Handler {
	return &Handler{
		tokenService: tokenService,
	}
}

func (h *Handler) ValidateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	var validateUserRequest entity.ValidateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&validateUserRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	validateUserRequest.UserID = entity.UserID(userID)

	ctx := r.Context()

	tokens, err := h.tokenService.ValidateUser(ctx, validateUserRequest)
	if err != nil {
		common.ErrorResponse(w, "error validating user", err)

		return
	}

	common.OkResponse(w, http.StatusOK, tokens)
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var tokens entity.Tokens

	if err := json.NewDecoder(r.Body).Decode(&tokens); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	ctx := r.Context()

	newTokens, err := h.tokenService.RefreshToken(ctx, tokens)
	if err != nil {
		common.ErrorResponse(w, "error refreshing token", err)

		return
	}

	common.OkResponse(w, http.StatusOK, newTokens)
}
