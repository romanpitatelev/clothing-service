package iamhandler

import (
	"context"
	_ "embed"
	"net/http"
	"strings"

	"github.com/romanpitatelev/clothing-service/internal/entity"
)

const (
	tokenLength = 3
)

func (h *Handler) JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		headerParts := strings.Split(header, " ")

		if headerParts[0] != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		encodedToken := strings.Split(headerParts[1], ".")
		if len(encodedToken) != tokenLength {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		claims, err := h.tokenService.ParseToken(headerParts[1])
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		userInfo := entity.UserInfo{
			UserID: claims.UserID,
			Email:  claims.Email,
			Role:   claims.Phone,
		}

		r = r.WithContext(context.WithValue(r.Context(), entity.UserInfo{}, userInfo))

		next.ServeHTTP(w, r)
	})
}
