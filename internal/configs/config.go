package configs

import (
	"crypto/rsa"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog/log"
)

const (
	JWTPrivateKeyPath = "./private_key.pem"
	JWTPublicKeyPath  = "./internal/configs/public_key.pem"
)

type Config struct {
	AppPort int `env:"APP_PORT" env-default:"8081" env-description:"Application port"`

	PostgresDSN string `env:"POSTGRES_DSN" env-default:"postgresql://postgres:my_pass@localhost:5432/clothing_db" env-description:"PostgreSQL DSN"`

	SMSAPIHost        string        `env:"SMS_API_HOST" env-default:"gate.smsaero.ru"`
	SMSAPISchema      string        `env:"SMS_API_SCHEMA" env-default:"https"`
	SMSEmail          string        `env:"SMS_EMAIL" env-default:"rpitatelev@gmail.com"`
	SMSAPIKey         string        `snf:"SMS_API_KEY" env-default:"o7KDkzhEcTFceryZLZ2xZcs3muTWgi-P"`
	SMSSenderName     string        `env:"SMS_SENDER_NAME" env-default:"Lookaround"`
	SMSSenderTestMode bool          `env:"SMS_SENDER_TEST_MODE" env-default:"true"`
	OTPMaxValue       int           `env:"OTP_MAX_VALUE" env-default:"9999"`
	OTPLifetime       time.Duration `env:"OTP_LIFETIME" env-default:"5m"`

	S3Address string `env:"S3_ADDRESS" env-default:"http://localhost:9000"`
	S3Bucket  string `env:"S3_BUCKET" env-default:"test.bucket"`
	S3Access  string `env:"S3_ACCESS_KEY" env-default:"access"`
	S3Secret  string `env:"S3_SECRET_KEY" env-default:"truesecret"`
	S3Region  string `env:"S3_REGION" env-default:"us-east-1"`

	JWTPrivateKey *rsa.PrivateKey
	JWTPublicKey  *rsa.PublicKey
}

func (e *Config) getHelpString() (string, error) {
	baseHeader := "Environment variables that can be set with env: "

	helpString, err := cleanenv.GetDescription(e, &baseHeader)
	if err != nil {
		return "", fmt.Errorf("failed to get help string: %w", err)
	}

	return helpString, nil
}

func New(withKeys bool) *Config {
	cfg := &Config{}

	helpString, err := cfg.getHelpString()
	if err != nil {
		log.Panic().Err(err).Msg("failed to get help string")
	}

	log.Info().Msg(helpString)

	if err := cleanenv.ReadEnv(cfg); err != nil {
		log.Panic().Err(err).Msg("failed to read config from envs")
	}

	if err = cleanenv.ReadConfig(".env", cfg); err != nil && !os.IsNotExist(err) {
		log.Panic().Err(err).Msg("failed to read config from .env")
	}

	if !withKeys {
		return cfg
	}

	privateKeyData, err := loadKeyFile(JWTPrivateKeyPath)
	if err != nil {
		log.Panic().Err(err).Msg("failed to load JWT private key")
	}

	publicKeyData, err := loadKeyFile(JWTPublicKeyPath)
	if err != nil {
		log.Panic().Err(err).Msg("failed to load JWT public key")
	}

	cfg.JWTPrivateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
	if err != nil {
		log.Panic().Err(err).Msg("failed to parse JWT private key")
	}

	cfg.JWTPublicKey, err = jwt.ParseRSAPublicKeyFromPEM(publicKeyData)
	if err != nil {
		log.Panic().Err(err).Msg("failed to parse JWT public key")
	}

	return cfg
}

func loadKeyFile(path string) ([]byte, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", path, err)
	}

	data, err := os.ReadFile(absPath) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to read key file %s: %w", absPath, err)
	}

	return data, nil
}
