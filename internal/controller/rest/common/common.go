package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/romanpitatelev/clothing-service/internal/entity"
	"github.com/rs/zerolog/log"
)

const DefaultLimit = 25

func ErrorResponse(w http.ResponseWriter, errorText string, err error) {
	statusCode := getStatusCode(err)

	errResp := fmt.Errorf("%s: %w", errorText, err).Error()
	if statusCode == http.StatusInternalServerError {
		errResp = http.StatusText(http.StatusInternalServerError)

		log.Warn().Err(err)
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
	case errors.Is(err, entity.ErrUserNotFound):
		return http.StatusNotFound
	case errors.Is(err, entity.ErrInvalidToken):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

func ParseGetRequest(r *http.Request) entity.GetRequestParams {
	queryParams := r.URL.Query()

	parameters := entity.GetRequestParams{
		Sorting: queryParams.Get("sorting"),
		Filter:  queryParams.Get("filter"),
	}

	var (
		limit  int64
		offset int64
	)

	if d := queryParams.Get("descending"); d != "" {
		parameters.Descending, _ = strconv.ParseBool(d)
	}

	if l := queryParams.Get("limit"); l != "" {
		if limit, _ = strconv.ParseInt(l, 0, 64); limit == 0 {
			limit = DefaultLimit
		}
	}

	if o := queryParams.Get("offset"); o != "" {
		offset, _ = strconv.ParseInt(o, 0, 64)
	}

	parameters.Limit = int(limit)
	parameters.Offset = int(offset)

	return parameters
}
