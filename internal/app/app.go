package app

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/romanpitatelev/clothing-service/internal/configs"
	"github.com/romanpitatelev/clothing-service/internal/controller/rest"
	iamhandler "github.com/romanpitatelev/clothing-service/internal/controller/rest/iam-handler"
	usershandler "github.com/romanpitatelev/clothing-service/internal/controller/rest/users-handler"
	smsregistrationrepo "github.com/romanpitatelev/clothing-service/internal/repository/sms-registration-repo"
	"github.com/romanpitatelev/clothing-service/internal/repository/store"
	usersrepo "github.com/romanpitatelev/clothing-service/internal/repository/users-repo"
	"github.com/romanpitatelev/clothing-service/internal/usecase/token-service"
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
	smsClient := smsregistrationrepo.New(smsregistrationrepo.Config{
		Schema:   cfg.SMSAPISchema,
		Host:     cfg.SMSAPIHost,
		Email:    cfg.SMSEmail,
		ApiKey:   cfg.SMSAPIKey,
		Sender:   cfg.SMSSenderName,
		TestMode: cfg.SMSSenderTestMode,
	})

	tokenService := tokenservice.New(tokenservice.Config{
		OTPLifetime: cfg.OTPLifetime,
		PublicKey:   cfg.JWTPublicKey,
		PrivateKey:  cfg.JWTPrivateKey,
	}, usersRepo)
	usersService := usersservice.New(usersservice.Config{
		OTPMaxValue: cfg.OTPMaxValue,
	}, usersRepo, smsClient)

	usersHandler := usershandler.New(usersService)
	iamHandler := iamhandler.New(tokenService)

	server := rest.New(
		rest.Config{Port: cfg.AppPort},
		usersHandler,
		iamHandler,
	)

	if err := server.Run(ctx); err != nil {
		return fmt.Errorf("failed to run the server: %w", err)
	}

	return nil
}
