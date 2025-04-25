package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/romanpitatelev/clothing-service/internal/controller/rest"
	usershandler "github.com/romanpitatelev/clothing-service/internal/controller/rest/users-handler"
	"github.com/romanpitatelev/clothing-service/internal/entity"
	"github.com/romanpitatelev/clothing-service/internal/repository/store"
	usersrepo "github.com/romanpitatelev/clothing-service/internal/repository/users-repo"
	smsregistration "github.com/romanpitatelev/clothing-service/internal/sms-registration"
	usersservice "github.com/romanpitatelev/clothing-service/internal/usecase/users-service"
	"github.com/rs/zerolog/log"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/suite"
)

const (
	pgDSN                = "postgresql://postgres:my_pass@localhost:5432/clothing_db"
	port                 = 5003
	userPath             = "api/v1/users"
	baseURL              = "https://direct.i-dgtl.ru"
	authToken            = "QWxhZGRpbjpvcGVuIHNlc2FtZQ=="
	senderName           = "sms_promo"
	codeLength           = 4
	codeValidityDuration = 3 * time.Minute
	cleanupPeriod        = 5 * time.Minute
)

type IntegrationTestSuite struct {
	suite.Suite
	cancelFunc   context.CancelFunc
	db           *store.DataStore
	usersrepo    *usersrepo.Repo
	usersservice *usersservice.Service
	usershandler *usershandler.Handler
	server       *rest.Server
	smsRepo      *smsregistration.SMSService
}

func (s *IntegrationTestSuite) SetupSuite() {
	log.Info().Msg("starting SetupSuite ...")

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel

	var err error

	s.db, err = store.New(ctx, store.Config{Dsn: pgDSN})
	s.Require().NoError(err)

	log.Info().Msg("starting new db ...")

	err = s.db.Migrate(migrate.Up)
	s.Require().NoError(err)

	log.Info().Msg("migrations are ready")

	s.usersrepo = usersrepo.New(s.db)

	s.smsRepo = smsregistration.New(smsregistration.Config{
		BaseURL:              fmt.Sprintf("%s/api/v1/message", baseURL),
		AuthToken:            authToken,
		SenderName:           senderName,
		CodeLength:           codeLength,
		CodeValidityDuration: codeValidityDuration,
		CleanupPeriod:        cleanupPeriod,
	})
	s.Require().NoError(err)

	log.Info().Msg("sms client is ready")

	s.usersservice = usersservice.New(s.usersrepo, s.smsRepo)

	s.usershandler = usershandler.New(s.usersservice)

	s.server = rest.New(rest.Config{Port: port}, s.usershandler, rest.GetPublicKey())

	go func() {
		err = s.server.Run(ctx)
		s.Require().NoError(err)
	}()

	time.Sleep(20 * time.Second)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.cancelFunc()
}

func (s *IntegrationTestSuite) TearDownTest() {
	err := s.db.Truncate(context.Background(), "users")
	s.Require().NoError(err)
}

func TestIntegrationSetupSuite(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) sendRequest(method, path string, status int, entity, result any, user entity.User) {
	body, err := json.Marshal(entity)
	s.Require().NoError(err)

	requestURL := fmt.Sprintf("http://localhost:%d%s", port, path)
	s.T().Logf("Sending request to %s", requestURL)

	request, err := http.NewRequestWithContext(context.Background(), method,
		fmt.Sprintf("http://localhost:%d%s", port, path), bytes.NewReader(body))
	s.Require().NoError(err, "failed to create request")

	token := s.getToken(user)
	request.Header.Set("Authorization", "Bearer "+token)

	client := http.Client{}

	response, err := client.Do(request)
	s.Require().NoError(err, "failed mto execute request")

	s.Require().NotNil(response, "response object is nil")

	defer func() {
		err = response.Body.Close()
		s.Require().NoError(err)
	}()

	s.T().Logf("Response Status Code: %d", response.StatusCode)

	if status != response.StatusCode {
		responseBody, err := io.ReadAll(response.Body)
		s.Require().NoError(err)

		s.T().Logf("Response Body: %s", string(responseBody))

		s.Require().Equal(status, response.StatusCode, "unexpected status code")

		return
	}

	if result == nil {
		return
	}

	err = json.NewDecoder(response.Body).Decode(result)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) getToken(user entity.User) string {
	claims := entity.Claims{
		UserID: user.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	privateKey, err := readPrivateKey()
	s.Require().NoError(err)

	token, err := generateToken(&claims, privateKey)
	s.Require().NoError(err)

	return token
}
