package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/romanpitatelev/clothing-service/internal/entity"
	"github.com/rs/zerolog/log"
)

func ErrorResponse(w http.ResponseWriter, errorText string, err error) {
	statusCode := getStatusCode(err)

	errResp := fmt.Errorf("%s: %w", errorText, err).Error()
	if statusCode == http.StatusInternalServerError {
		errResp = http.StatusText(http.StatusInternalServerError)
	}

	response, err := json.Marshal(errResp)
	if err != nil {
		log.Warn().Msgf("error marshalling response: %v", err)
	}

	w.WriteHeader(statusCode)

	if _, err := w.Write(response); err != nil {
		log.Warn().Msgf("error writing response: %v", err)
	}
}

func OkResponse(w http.ResponseWriter, status int, response any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Warn().Msgf("error encoding response: %v", err)
	}
}

func getStatusCode(err error) int {
	switch {
	case errors.Is(err, entity.ErrUserNotFound) ||
		errors.Is(err, entity.ErrClothingNotFound):
		return http.StatusNotFound
	case errors.Is(err, entity.ErrInvalidPhone):
		return http.StatusBadRequest
	case errors.Is(err, entity.ErrInvalidOTP):
		return http.StatusForbidden
	case errors.Is(err, entity.ErrDuplicateContact):
		return http.StatusConflict
	case errors.Is(err, entity.ErrTokenExpired) ||
		errors.Is(err, entity.ErrInvalidToken) ||
		errors.Is(err, entity.ErrInvalidUUIDFormat) ||
		errors.Is(err, entity.ErrInvalidSigningMethod):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
