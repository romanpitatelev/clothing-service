package app

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/romanpitatelev/clothing-service/internal/configs"
	"github.com/romanpitatelev/clothing-service/internal/controller/rest"
	usershandler "github.com/romanpitatelev/clothing-service/internal/controller/rest/users-handler"
	smsregistrationrepo "github.com/romanpitatelev/clothing-service/internal/repository/sms-registration-repo"
	"github.com/romanpitatelev/clothing-service/internal/repository/store"
	usersrepo "github.com/romanpitatelev/clothing-service/internal/repository/users-repo"
	tokenservice "github.com/romanpitatelev/clothing-service/internal/token-service"
	usersservice "github.com/romanpitatelev/clothing-service/internal/usecase/users-service"
	"github.com/rs/zerolog/log"
	migrate "github.com/rubenv/sql-migrate"
)

func Run(cfg *configs.Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	db, err := store.New(ctx, store.Config{Dsn: cfg.PostgresDSN})
	if err != nil {
		log.Panic().Err(err).Msg("failed to connect to database")
	}

	if err := db.Migrate(migrate.Up); err != nil {
		log.Panic().Err(err).Msg("failed to migrate")
	}

	log.Info().Msg("successful migration")

	usersRepo := usersrepo.New(db)
	smsService := smsregistrationrepo.New(smsregistrationrepo.Config{
		Email:                cfg.SMSEmail,
		ApiKey:               cfg.SMSAPIKey,
		Sender:               cfg.SMSSenderName,
		CodeLength:           cfg.SMSCodeLength,
		CodeValidityDuration: cfg.SMSCodeValidityDuration,
	})

	tokenGenerator := tokenservice.New(cfg.JWTPrivateKey, cfg.JWTPublicKey)

	usersService := usersservice.New(
		usersRepo,
		smsService,
		tokenGenerator,
	)

	usersHandler := usershandler.New(usersService)

	server := rest.New(
		rest.Config{Port: cfg.AppPort},
		usersHandler,
		rest.GetPublicKey(),
	)

	if err := server.Run(ctx); err != nil {
		return fmt.Errorf("failed to run the server: %w", err)
	}

	return nil
}
