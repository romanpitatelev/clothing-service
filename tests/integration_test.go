package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/romanpitatelev/clothing-service/internal/controller/rest"
	iamhandler "github.com/romanpitatelev/clothing-service/internal/controller/rest/iam-handler"
	usershandler "github.com/romanpitatelev/clothing-service/internal/controller/rest/users-handler"
	"github.com/romanpitatelev/clothing-service/internal/entity"
	smsregistrationrepo "github.com/romanpitatelev/clothing-service/internal/repository/sms-registration-repo"
	"github.com/romanpitatelev/clothing-service/internal/repository/store"
	usersrepo "github.com/romanpitatelev/clothing-service/internal/repository/users-repo"
	"github.com/romanpitatelev/clothing-service/internal/usecase/token-service"
	usersservice "github.com/romanpitatelev/clothing-service/internal/usecase/users-service"
	"github.com/rs/zerolog/log"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/suite"
)

const (
	pgDSN               = "postgresql://postgres:my_pass@localhost:5432/clothing_db"
	port                = 5003
	userPath            = "/api/v1/users"
	email               = "rpitatelev@gmail.com"
	sender              = "clothing-service"
	codeLength          = 4
	accessTokenDuration = 3 * time.Minute
)

type IntegrationTestSuite struct {
	suite.Suite
	cancelFunc   context.CancelFunc
	db           *store.DataStore
	usersRepo    *usersrepo.Repo
	tokenService *tokenservice.Service
	usersService *usersservice.Service
	iamHandler   *iamhandler.Handler
	usersHandler *usershandler.Handler
	server       *rest.Server
	smsRepo      *smsregistrationrepo.SMSService
	smsChan      chan otpResp
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

	privateKey, err := readPrivateKey()
	s.Require().NoError(err)

	publicKey := &privateKey.PublicKey
	s.usersRepo = usersrepo.New(s.db)
	s.smsRepo = smsregistrationrepo.New(smsregistrationrepo.Config{
		Host:   "localhost:" + strconv.Itoa(port+1),
		Schema: "http",
		Email:  "biba",
		ApiKey: "boba",
		Sender: "pasha",
	})

	s.tokenService = tokenservice.New(tokenservice.Config{
		OTPLifetime: otpDuration,
		PrivateKey:  privateKey,
		PublicKey:   publicKey,
	}, s.usersRepo)
	s.usersService = usersservice.New(usersservice.Config{
		OTPMaxValue: 9999,
	}, s.usersRepo, s.smsRepo)

	s.usersHandler = usershandler.New(s.usersService)
	s.iamHandler = iamhandler.New(s.tokenService)

	s.server = rest.New(rest.Config{Port: port}, s.usersHandler, s.iamHandler)

	log.Info().Msg("sms client is ready")

	go func() {
		s.smsChan = make(chan otpResp)
		s.runServer(ctx, ":"+strconv.Itoa(port+1))
	}()

	//nolint:testifylint
	go func() {
		err = s.server.Run(ctx)
		s.Require().NoError(err)
	}()

	time.Sleep(50 * time.Millisecond)
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
	s.Require().NoError(err, "failed to execute request")

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
