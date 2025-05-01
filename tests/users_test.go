package tests

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/romanpitatelev/clothing-service/internal/entity"
)

const (
	otpDuration = 5 * time.Second
)

func (s *IntegrationTestSuite) TestCreateUser() {
	phone := "89877342381"
	user := entity.User{
		UserID:       entity.UserID(uuid.New()),
		Phone:        &phone,
		OTP:          "1234",
		OTPExpiresAt: time.Now().Add(otpDuration),
	}

	s.Run("successful user creation", func() {
		s.sendRequest(http.MethodPost, userPath+"/register", http.StatusOK, &user, nil, entity.User{})
	})

	s.Run("invalid phone number", func() {
		invalidPhone := "85634"
		invalidUser := entity.User{
			UserID:       entity.UserID(uuid.New()),
			Phone:        &invalidPhone,
			OTP:          "1234",
			OTPExpiresAt: time.Now().Add(otpDuration),
		}

		s.sendRequest(http.MethodPost, userPath+"/register", http.StatusBadRequest, &invalidUser, nil, entity.User{})
	})
}

func (s *IntegrationTestSuite) TestValidateUser() {
	phone := "79877342381"
	validOTP := "1234"
	wrongOTP := "0000"
	userID := entity.UserID(uuid.New())

	s.Run("successful validation", func() {
		user := entity.User{
			UserID:       userID,
			Phone:        &phone,
			OTP:          validOTP,
			OTPExpiresAt: time.Now().Add(otpDuration),
		}

		s.sendRequest(http.MethodPost, userPath+"/register", http.StatusOK, &user, nil, entity.User{})

		var tokens entity.Tokens

		s.sendRequest(http.MethodPatch, userPath+"/register/otp", http.StatusOK, &user, &tokens, entity.User{})

		s.Require().NotEmpty(tokens.AccessToken)
		s.Require().NotEmpty(tokens.RefreshToken)
		s.Require().NotEmpty(tokens.Timeout)
	})

	s.Run("failed validation with correct otp after expiration", func() {
		shortOTPExpirationUser := entity.User{
			UserID:       userID,
			Phone:        &phone,
			OTP:          validOTP,
			OTPExpiresAt: time.Now().Add(-1 * time.Minute),
		}

		s.sendRequest(http.MethodPost, userPath+"/register", http.StatusOK, &shortOTPExpirationUser, nil, entity.User{})

		<-s.smsService.sendOTPChan

		time.Sleep(2 * time.Second)

		s.sendRequest(http.MethodPatch, userPath+"/register/otp", http.StatusGone, &shortOTPExpirationUser, nil, entity.User{})
	})

	s.Run("failed validation with incorrect otp", func() {
		incorrectOTPUser := entity.User{
			UserID:       userID,
			Phone:        &phone,
			OTP:          validOTP,
			OTPExpiresAt: time.Now().Add(5 * time.Minute),
		}

		s.sendRequest(http.MethodPost, userPath+"/register", http.StatusOK, &incorrectOTPUser, nil, entity.User{})

		<-s.smsService.sendOTPChan

		invalidReq := entity.User{
			UserID: userID,
			OTP:    wrongOTP,
		}

		s.sendRequest(http.MethodPatch, userPath+"/register/otp", http.StatusForbidden, &invalidReq, nil, entity.User{})
	})
}

func (s *IntegrationTestSuite) TestGetUser() {
	firstName := "John"
	lastName := "Ivanov"
	age := 20
	user := entity.User{
		UserID:     entity.UserID(uuid.New()),
		FirstName:  &firstName,
		LastName:   &lastName,
		Age:        &age,
		IsVerified: true,
	}

	err := s.db.UpsertUser(context.Background(), user)
	s.Require().NoError(err)

	s.Run("user not found", func() {
		uuidString := uuid.New().String()
		userIDPath := userPath + "/" + uuidString

		s.sendRequest(http.MethodGet, userIDPath, http.StatusNotFound, nil, nil, entity.User{})
	})

	s.Run("get user successfully", func() {
		uuidString := uuid.UUID(user.UserID).String()
		userIDPath := userPath + "/" + uuidString

		s.sendRequest(http.MethodGet, userIDPath, http.StatusOK, nil, &user, entity.User{})
	})
}
