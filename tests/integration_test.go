package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/romanpitatelev/clothing-service/internal/controller/rest"
	clotheshandler "github.com/romanpitatelev/clothing-service/internal/controller/rest/clothes-handler"
	fileshandler "github.com/romanpitatelev/clothing-service/internal/controller/rest/files-handler"
	iamhandler "github.com/romanpitatelev/clothing-service/internal/controller/rest/iam-handler"
	usershandler "github.com/romanpitatelev/clothing-service/internal/controller/rest/users-handler"
	"github.com/romanpitatelev/clothing-service/internal/entity"
	filesrepo "github.com/romanpitatelev/clothing-service/internal/repository/objects-repo"
	clothesrepo "github.com/romanpitatelev/clothing-service/internal/repository/products-repo"
	smsregistrationrepo "github.com/romanpitatelev/clothing-service/internal/repository/sms-registration-repo"
	"github.com/romanpitatelev/clothing-service/internal/repository/store"
	usersrepo "github.com/romanpitatelev/clothing-service/internal/repository/users-repo"
	clothesservice "github.com/romanpitatelev/clothing-service/internal/usecase/clothes-service"
	filesservice "github.com/romanpitatelev/clothing-service/internal/usecase/files-service"
	"github.com/romanpitatelev/clothing-service/internal/usecase/token-service"
	usersservice "github.com/romanpitatelev/clothing-service/internal/usecase/users-service"
	"github.com/rs/zerolog/log"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/suite"
)

const (
	pgDSN       = "postgresql://postgres:my_pass@localhost:5432/clothing_db"
	port        = 5003
	userPath    = "/api/v1/users"
	clothesPath = "/api/v1/clothes"
	imagesPath  = "/api/v1/images"
)

type IntegrationTestSuite struct {
	suite.Suite
	cancelFunc     context.CancelFunc
	db             *store.DataStore
	usersRepo      *usersrepo.Repo
	clothesRepo    *clothesrepo.Repo
	filesRepo      *filesrepo.S3
	smsRepo        *smsregistrationrepo.SMSService
	tokenService   *tokenservice.Service
	usersService   *usersservice.Service
	clothesService *clothesservice.Service
	filesService   *filesservice.Service
	iamHandler     *iamhandler.Handler
	usersHandler   *usershandler.Handler
	clothesHandler *clotheshandler.Handler
	filesHandler   *fileshandler.Handler
	server         *rest.Server
	smsChan        chan otpResp
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
	s.smsRepo = smsregistrationrepo.New(smsregistrationrepo.Config{
		Host:   "localhost:" + strconv.Itoa(port+1),
		Schema: "http",
		Email:  "biba",
		ApiKey: "boba",
		Sender: "pasha",
	})
	s.usersRepo = usersrepo.New(s.db)
	s.clothesRepo = clothesrepo.New(s.db)
	s.filesRepo, err = filesrepo.New(filesrepo.S3Config{
		Address: "http://localhost:9000",
		Bucket:  "test.bucket",
		Access:  "access",
		Secret:  "truesecret",
		Region:  "us-east-1",
	})

	s.tokenService = tokenservice.New(tokenservice.Config{
		OTPLifetime: otpDuration,
		PrivateKey:  privateKey,
		PublicKey:   publicKey,
	}, s.usersRepo)
	s.usersService = usersservice.New(usersservice.Config{
		OTPMaxValue: 9999,
	}, s.usersRepo, s.smsRepo)
	// s.clothesService = clothesservice.New(s.clothesRepo)
	s.filesService = filesservice.New(s.filesRepo)

	s.usersHandler = usershandler.New(s.usersService)
	s.iamHandler = iamhandler.New(s.tokenService)
	s.clothesHandler = clotheshandler.New(s.clothesService)
	s.filesHandler = fileshandler.New(s.filesService)

	s.server = rest.New(rest.Config{Port: port}, s.usersHandler, s.iamHandler, s.clothesHandler, s.filesHandler)

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
	err := s.db.Truncate(context.Background(), "users", "variants", "products", "brands")
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

func (s *IntegrationTestSuite) getFile(method, path string, status int, entity any, user entity.User) ([]byte, string, string) {
	body, err := json.Marshal(entity)
	s.Require().NoError(err)

	req, err := http.NewRequestWithContext(context.Background(), method,
		fmt.Sprintf("http://localhost:%d%s", port, path), bytes.NewReader(body))
	s.Require().NoError(err)

	token := s.getToken(user)
	req.Header.Set("Authorization", "Bearer "+token)

	client := http.Client{}

	resp, err := client.Do(req)
	s.Require().NoError(err)

	defer func() {
		err = resp.Body.Close()
		s.Require().NoError(err)
	}()

	s.Require().Equal(status, resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	slice := strings.SplitN(resp.Header.Get("Content-Disposition"), "=", 2)

	return respBody, resp.Header.Get("Content-Type"), slice[len(slice)-1]
}

func (s *IntegrationTestSuite) getToken(user entity.User) string {
	claims := entity.Claims{
		UserID: user.ID,
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
