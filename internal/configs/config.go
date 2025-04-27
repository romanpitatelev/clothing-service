package configs

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog/log"
	"os"
)

type Config struct {
	AppPort int `env:"APP_PORT" env-default:"8081" env-description:"Application port"`

	PostgresDSN string `env:"POSTGRES_PORT" env-default:"postgresql://postgres:my_pass@localhost:5432/clothing_db" env-description:"PostgreSQL DSN"`

	SMSToken      string `env:"SMS_AUTH_TOKEN"  env-default:"QWxhZGRpbjpvcGVuIHNlc2FtZQ=="`
	SMSSenderName string `env:"SMS_SENDER_NAME" env-default:"sms_promo"`

	S3Address string `env:"S3_ADDRESS" env-default:"http://localhost:9000"`
	S3Bucket  string `env:"S3_BUCKET" env-default:"test.bucket"`
	S3Access  string `env:"S3_ACCESS_KEY" env-default:"access"`
	S3Secret  string `env:"S3_SECRET_KEY" env-default:"truesecret"`
	S3Region  string `env:"S3_REGION" env-default:"us-east-1"`
}

func (e *Config) getHelpString() (string, error) {
	baseHeader := "Environment variables that can be set with env: "

	helpString, err := cleanenv.GetDescription(e, &baseHeader)
	if err != nil {
		return "", fmt.Errorf("failed to get help string: %w", err)
	}

	return helpString, nil
}

func New() *Config {
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

	return cfg
}
