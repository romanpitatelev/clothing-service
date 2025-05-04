package tests

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/romanpitatelev/clothing-service/internal/entity"
	"github.com/romanpitatelev/clothing-service/internal/utils"
)

const (
	otpDuration = 5 * time.Minute
)

func (s *IntegrationTestSuite) TestCreateUser() {
	phone := "79877342381"
	user := entity.User{
		Phone: phone,
	}

	var newUser entity.User

	s.Run("successful user creation", func() {
		s.sendRequest(http.MethodPost, userPath+"/register", http.StatusOK, user, &newUser, entity.User{})
		s.Require().Equal(user.Phone, newUser.Phone)
		<-s.smsChan
	})

	s.Run("existing phone", func() {
		s.sendRequest(http.MethodPost, userPath+"/register", http.StatusConflict, user, nil, entity.User{})
	})

	s.Run("invalid phone number", func() {
		invalidUser := entity.User{
			Phone: "85634",
		}

		s.sendRequest(http.MethodPost, userPath+"/register", http.StatusBadRequest, invalidUser, nil, entity.User{})
	})
}

func (s *IntegrationTestSuite) TestValidateLoginUser() {
	user := entity.User{
		Phone: "79877342381",
	}

	var req entity.ValidateUserRequest

	s.Run("successful validation", func() {
		s.sendRequest(http.MethodPost, userPath+"/register", http.StatusOK, user, &user, entity.User{})

		resp := <-s.smsChan
		s.Require().Equal("pasha", resp.sender)
		s.Require().Equal(user.Phone, resp.phone)

		req.OTP = resp.otp

		var tokens entity.Tokens

		s.sendRequest(http.MethodPost, userPath+"/"+user.ID.String()+"/otp", http.StatusOK, req, &tokens, entity.User{})

		s.Require().NotEmpty(tokens.AccessToken)
		s.Require().NotEmpty(tokens.RefreshToken)
		s.Require().NotEmpty(tokens.Timeout)
	})

	s.Run("login successfully", func() {
		s.sendRequest(http.MethodPost, userPath+"/"+user.ID.String()+"/login", http.StatusOK, nil, nil, entity.User{})

		resp := <-s.smsChan
		req.OTP = resp.otp

		var tokens entity.Tokens

		s.sendRequest(http.MethodPost, userPath+"/"+user.ID.String()+"/otp", http.StatusOK, req, &tokens, entity.User{})

		s.Require().NotEmpty(tokens.AccessToken)
		s.Require().NotEmpty(tokens.RefreshToken)
		s.Require().NotEmpty(tokens.Timeout)
	})

	s.Run("try to login with not existing user", func() {
		s.sendRequest(http.MethodPost, userPath+"/"+uuid.New().String()+"/login", http.StatusNotFound, nil, nil, entity.User{})
	})

	s.Run("failed validation with incorrect otp", func() {
		s.sendRequest(http.MethodPost, userPath+"/"+user.ID.String()+"/login", http.StatusOK, nil, nil, entity.User{})

		resp := <-s.smsChan
		req.OTP = resp.otp + "1"

		s.sendRequest(http.MethodPost, userPath+"/"+user.ID.String()+"/otp", http.StatusForbidden, req, nil, entity.User{})
	})

	s.Run("failed validation with correct otp after expiration", func() {
		s.sendRequest(http.MethodPost, userPath+"/"+user.ID.String()+"/login", http.StatusOK, nil, nil, entity.User{})

		resp := <-s.smsChan
		req.OTP = resp.otp

		_, err := s.db.Exec(s.T().Context(), `UPDATE users SET otp_created_at = otp_created_at - INTERVAL '1 DAY'`)
		s.Require().NoError(err)

		s.sendRequest(http.MethodPost, userPath+"/"+user.ID.String()+"/otp", http.StatusForbidden, req, nil, entity.User{})
	})
}

func (s *IntegrationTestSuite) TestGetUpdateDeleteUser() {
	user := entity.User{
		ID:            entity.UserID(uuid.New()),
		FirstName:     utils.Pointer("John"),
		LastName:      utils.Pointer("Ivanov"),
		BirthDate:     utils.Pointer(time.Now()),
		Phone:         "79031355530",
		PhoneVerified: true,
	}

	err := s.db.UpsertUser(context.Background(), user)
	s.Require().NoError(err)

	s.Run("user not found", func() {
		userIDPath := userPath + "/" + uuid.New().String()

		s.sendRequest(http.MethodGet, userIDPath, http.StatusNotFound, nil, nil, entity.User{})
	})

	var newUser entity.User

	s.Run("get user successfully", func() {
		userIDPath := userPath + "/" + user.ID.String()

		s.sendRequest(http.MethodGet, userIDPath, http.StatusOK, nil, &newUser, entity.User{})
		s.Require().Equal(user.ID, newUser.ID)
		s.Require().Equal(user.FirstName, newUser.FirstName)
		s.Require().Equal(user.LastName, newUser.LastName)
		s.Require().Equal(user.Phone, newUser.Phone)
		s.Require().Equal(user.PhoneVerified, newUser.PhoneVerified)
	})

	s.Run("update not exists", func() {
		s.sendRequest(http.MethodPatch, userPath+"/"+uuid.New().String(), http.StatusNotFound, nil, nil, entity.User{})
	})

	user2 := entity.User{
		ID:            entity.UserID(uuid.New()),
		FirstName:     utils.Pointer("John"),
		LastName:      utils.Pointer("Ivanov"),
		BirthDate:     utils.Pointer(time.Now()),
		Phone:         "79031355531",
		PhoneVerified: true,
	}

	err = s.db.UpsertUser(context.Background(), user2)
	s.Require().NoError(err)

	updateUser := entity.UserUpdate{
		FirstName: utils.Pointer("Biba"),
		LastName:  utils.Pointer("Boba"),
		BirthDate: utils.Pointer(time.Now().Add(-24 * time.Hour)),
		Phone:     &user.Phone,
	}

	s.Run("update conflict contact", func() {
		s.sendRequest(http.MethodPatch, userPath+"/"+user2.ID.String(), http.StatusConflict, updateUser, nil, entity.User{})
	})

	var updatedUser entity.User

	updateUser.Phone = nil

	s.Run("update user successfully", func() {
		s.sendRequest(http.MethodPatch, userPath+"/"+user2.ID.String(), http.StatusOK, updateUser, &updatedUser, entity.User{})
		s.Require().Equal(user2.ID, updatedUser.ID)
		s.Require().Equal(*updateUser.FirstName, *updatedUser.FirstName)
		s.Require().Equal(*updateUser.LastName, *updatedUser.LastName)
	})

	s.Run("delete non existent", func() {
		s.sendRequest(http.MethodDelete, userPath+"/"+uuid.New().String(), http.StatusNotFound, nil, nil, entity.User{})
	})

	s.Run("delete non existent", func() {
		s.sendRequest(http.MethodDelete, userPath+"/"+user.ID.String(), http.StatusNoContent, nil, nil, entity.User{})
	})

	s.Run("change phone for existing user to the value belonging to the deleted one", func() {
		updateUser.Phone = &user.Phone
		s.sendRequest(http.MethodPatch, userPath+"/"+user2.ID.String(), http.StatusOK, updateUser, nil, entity.User{})
	})
}
