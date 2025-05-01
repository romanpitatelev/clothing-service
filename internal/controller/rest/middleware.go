package rest

import (
	"context"
	"crypto/rsa"
	_ "embed"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/romanpitatelev/clothing-service/internal/entity"
)

const (
	tokenLength   = 3
	tokenDuration = 24 * time.Hour
)

//go:embed keys/public_key.pem
var publicKeyData []byte

func (s *Server) jwtAuth(next http.Handler) http.Handler {
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

		token, err := jwt.ParseWithClaims(headerParts[1], &entity.Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, entity.ErrInvalidSigningMethod
			}

			return s.key, nil
		})
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		claims, ok := token.Claims.(*entity.Claims)
		if !ok || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		if claims.ExpiresAt.Before(time.Now()) {
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

func NewClaims() *entity.Claims {
	tokenTime := time.Now().Add(tokenDuration)

	return &entity.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(tokenTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
}

func ReadPublicKey() (*rsa.PublicKey, error) {
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyData)
	if err != nil {
		return nil, fmt.Errorf("error parsing public key: %w", err)
	}

	return publicKey, nil
}

func GetPublicKey() *rsa.PublicKey {
	key, err := ReadPublicKey()
	if err != nil {
		return nil
	}

	return key
}
