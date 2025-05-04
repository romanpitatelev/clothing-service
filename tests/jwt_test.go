package tests

import (
	"context"
	"crypto/rsa"
	_ "embed"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/romanpitatelev/clothing-service/internal/entity"
	"github.com/romanpitatelev/clothing-service/internal/utils"
)

func (s *IntegrationTestSuite) TestJWT() {
	userID, err := uuid.Parse("5131f6fe-1447-4945-a637-5b33a233e47e")
	s.Require().NoError(err)

	user := entity.User{
		ID:            entity.UserID(userID),
		FirstName:     utils.Pointer("John"),
		LastName:      utils.Pointer("Ivanov"),
		BirthDate:     utils.Pointer(time.Now()),
		Phone:         "79031355531",
		PhoneVerified: true,
	}

	err = s.db.UpsertUser(context.Background(), user)
	s.Require().NoError(err)

	var tokens entity.Tokens

	s.Run("refresh token invalid token", func() {
		s.sendRequest(http.MethodPost, userPath+"/"+user.ID.String()+"/refresh", http.StatusUnauthorized, tokens, nil, user)
	})

	tokens.RefreshToken = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiJcIjE0M2Y3M2NhLTA5M2YtNDdmNi1iMzcyLTNlZGNhYzAyNTFiOFwiIiwiZW1haWwiOiIiLCJwaG9uZSI6IiIsImV4cCI6MTc0NjIyMTU0NCwiaWF0IjoxNzQ2MzA3OTQ0fQ.WqSV7qz0AtZ1KwkQLtwsATvIAvUjNGDo_JRt1faqvrtDZ6TO39RdEzJWAhl0qmhrAqoM3bQeaB1vxSfv9EMGvmWJMo5KP0DrSpxRBL0KsycEL3EnK5ov81iDFq3txXdn8my5iShYQAOdASdDZ-IGRMxuGVKRjheJf3bMdyTrvtr6jopKWd1rEPlDDcEUuaXrIjIhyNSLwlT4P_LG3g1KHqeC_ZLNZC8Zqo31t0jelS35klTFx8DypHqUqxSFDuLBG6KqiB7h1jlNaonE_i7znkzWOofEouXYQqKrHoaaZYkhyZqGh1A_5YQW4XiRIbKmH7-TKHTAQtaN1s7jnTxQ1w" //nolint:lll

	s.Run("token expired", func() {
		s.sendRequest(http.MethodPost, userPath+"/"+uuid.NewString()+"/refresh", http.StatusUnauthorized, tokens, nil, user)
	})

	tokens.RefreshToken = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjQzMzg1NTY0NDAsInVzZXJJZCI6IjUxMzFmNmZlLTE0NDctNDk0NS1hNjM3LTViMzNhMjMzZTQ3ZSJ9.DXBNDdeFHpRxZ9u0ZvvKQXf3LcjzT_dduQHCqxVY-IIoIFBAASAcL13FGJgv8jW1g3tcoOsno40sWyFCcpBsngRzAt8HDiiiPbz8MVSSvCcfNe8jsupaV5DnxPKyzPd9jexdOrswD8b_UVU75zrFhU4sgnU0fTXHWj0FA6KmmNLyVPf1oWz4hHlRb42iAruZv4782hcaDe5khu3ZVHO5gO9Qzj42EyGEbm5Hk03v8_cEgUcHn9APQtPCaE6iH9TY-yD6TuH1w0CAZl8qElJSbvTLvVXKTOeufmSi5u_35DUT6R724zdn2ZD8bXwzZfF7HgzW1NlOGH_3TDDZW0rrVQ" //nolint:lll

	s.Run("refresh token for non existent user", func() {
		s.sendRequest(http.MethodPost, userPath+"/"+uuid.NewString()+"/refresh", http.StatusOK, tokens, nil, user)
	})

	var newTokens entity.Tokens

	s.Run("refresh token", func() {
		s.sendRequest(http.MethodPost, userPath+"/"+user.ID.String()+"/refresh", http.StatusOK, tokens, &newTokens, user)
		s.Require().NotEqual(tokens.RefreshToken, newTokens.RefreshToken)
	})
}

//go:embed keys/private_key.pem
var privateKeyData []byte

func readPrivateKey() (*rsa.PrivateKey, error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %w", err)
	}

	return privateKey, nil
}

func generateToken(claims *entity.Claims, secret *rsa.PrivateKey) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenStr, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token^ %w", err)
	}

	return tokenStr, nil
}
