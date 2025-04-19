package configs

import (
	"fmt"
	"os"
	"reflect"
	"regexp"

	"github.com/davecgh/go-spew/spew"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog/log"
)

const envFileName = ".env"

type Config struct {
	env *EnvSetting
}

type EnvSetting struct {
	AppPort     int    `env:"APP_PORT" env-default:"8081" env-description:"Application port"`
	PostgresDSN string `env:"POSTGRES_PORT" env-default:"postgresql://postgres:my_pass@localhost:5432/wallets_db" env-description:"PostgreSQL DSN"`
}

func findConfigFile() bool {
	_, err := os.Stat(envFileName)

	return err == nil
}

func (e *EnvSetting) GetHelpString() (string, error) {
	baseHeader := "Environment variables that can be set with env: "

	helpString, err := cleanenv.GetDescription(e, &baseHeader)
	if err != nil {
		return "", fmt.Errorf("failed to get help string: %w", err)
	}

	return helpString, nil
}

func New() *Config {
	envSetting := &EnvSetting{}

	helpString, err := envSetting.GetHelpString()
	if err != nil {
		log.Panic().Err(err).Msg("failed to get help string")
	}

	log.Info().Msg(helpString)

	if findConfigFile() {
		if err := cleanenv.ReadConfig(envFileName, envSetting); err != nil {
			log.Panic().Err(err).Msg("failed to read env config")
		}
	} else {
		log.Panic().Err(err).Msg("error reading env config")
	}

	return &Config{env: envSetting}
}

func (c *Config) PrintDebug() {
	envReflect := reflect.Indirect(reflect.ValueOf(c.env))
	envReflectType := envReflect.Type()

	exp := regexp.MustCompile("([Tt]oken[Pp]assword)")

	for i := range envReflect.NumField() {
		key := envReflectType.Field(i).Name

		if exp.MatchString(key) {
			val, _ := envReflect.Field(i).Interface().(string)
			log.Debug().Msgf("%s: len %d", key, len(val))

			continue
		}

		log.Debug().Msgf("%s: %v", key, spew.Sprintf("%#v", envReflect.Field(i).Interface()))
	}
}

func (c *Config) GetAppPort() int {
	return c.env.AppPort
}

func (c *Config) GetPostgresDSN() string {
	return c.env.PostgresDSN
}
