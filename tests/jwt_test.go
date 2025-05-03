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
	userID, err := uuid.Parse("143f73ca-093f-47f6-b372-3edcac0251b8")
	s.Require().NoError(err)

	user := entity.User{
		UserID:        entity.UserID(userID),
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
		s.sendRequest(http.MethodPost, userPath+"/"+user.UserID.String()+"/refresh", http.StatusUnauthorized, tokens, nil, user)
	})

	tokens.RefreshToken = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiJcIjE0M2Y3M2NhLTA5M2YtNDdmNi1iMzcyLTNlZGNhYzAyNTFiOFwiIiwiZW1haWwiOiIiLCJwaG9uZSI6IiIsImV4cCI6MTc0NjIyMTU0NCwiaWF0IjoxNzQ2MzA3OTQ0fQ.WqSV7qz0AtZ1KwkQLtwsATvIAvUjNGDo_JRt1faqvrtDZ6TO39RdEzJWAhl0qmhrAqoM3bQeaB1vxSfv9EMGvmWJMo5KP0DrSpxRBL0KsycEL3EnK5ov81iDFq3txXdn8my5iShYQAOdASdDZ-IGRMxuGVKRjheJf3bMdyTrvtr6jopKWd1rEPlDDcEUuaXrIjIhyNSLwlT4P_LG3g1KHqeC_ZLNZC8Zqo31t0jelS35klTFx8DypHqUqxSFDuLBG6KqiB7h1jlNaonE_i7znkzWOofEouXYQqKrHoaaZYkhyZqGh1A_5YQW4XiRIbKmH7-TKHTAQtaN1s7jnTxQ1w" //nolint:lll

	s.Run("token expired", func() {
		s.sendRequest(http.MethodPost, userPath+"/"+uuid.NewString()+"/refresh", http.StatusUnauthorized, tokens, nil, user)
	})

	tokens.RefreshToken = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiJcIjE0M2Y3M2NhLTA5M2YtNDdmNi1iMzcyLTNlZGNhYzAyNTFiOFwiIiwiZW1haWwiOiIiLCJwaG9uZSI6IiIsImV4cCI6MTc0NjM5MzU4MiwiaWF0IjoxNzQ2MzA3MTgyfQ.RD_rVym4t_1P5lGksZQ5LcYAkL01AuiOBmjWAFyEFQxMy8_HlWgvqXy3u2JXPTqWbfVUOIER1covti8ccASZ21E2Eh0XsnKCYZ_QIvvBmPzh3o-AgRs7YpMU6borEQvZHlYmRP0GuZw_fqlVTjOhDcJTcSh_XQKrBtcN-la2Wh-6fjL4MEVO8pPtKe8tzkaLAz11D2CzxoJpf8nQQq0J7jz5EvrvqGZvoWr_8n6Pu0OIKuJvxqsqCsW48seatfeJn1qSqu6DMfFf0rk-TaYMAxSeeuskxCpzIl9y-Dy6KjLxz_wWMnqqVL4k2yjHuG49fgD33i9mNwRRWdXv9x6zEA" //nolint:lll

	s.Run("refresh token for non existent user", func() {
		s.sendRequest(http.MethodPost, userPath+"/"+uuid.NewString()+"/refresh", http.StatusOK, tokens, nil, user)
	})

	var newTokens entity.Tokens

	s.Run("refresh token", func() {
		s.sendRequest(http.MethodPost, userPath+"/"+user.UserID.String()+"/refresh", http.StatusOK, tokens, &newTokens, user)
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
