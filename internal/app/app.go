package app

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/romanpitatelev/clothing-service/internal/configs"
	"github.com/romanpitatelev/clothing-service/internal/controller/rest"
	usershandler "github.com/romanpitatelev/clothing-service/internal/controller/rest/users-handler"
	"github.com/romanpitatelev/clothing-service/internal/repository/sms-registration"
	"github.com/romanpitatelev/clothing-service/internal/repository/store"
	usersrepo "github.com/romanpitatelev/clothing-service/internal/repository/users-repo"
	usersservice "github.com/romanpitatelev/clothing-service/internal/usecase/users-service"
	"github.com/rs/zerolog/log"
	migrate "github.com/rubenv/sql-migrate"
)

func Run(cfg *configs.Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	db, err := store.New(ctx, store.Config{Dsn: cfg.GetPostgresDSN()})
	if err != nil {
		log.Panic().Err(err).Msg("failed to connect to database")
	}

	if err := db.Migrate(migrate.Up); err != nil {
		log.Panic().Err(err).Msg("failed to migrate")
	}

	log.Info().Msg("successful migration")

	usersRepo := usersrepo.New(db)
	smsService := smsregistration.New(cfg.GetSMSConfig())
	smsService.StartCleanup()

	usersService := usersservice.New(
		usersRepo,
		smsService,
	)

	usersHandler := usershandler.New(usersService)

	server := rest.New(
		rest.Config{Port: cfg.GetAppPort()},
		usersHandler,
		rest.GetPublicKey(),
	)

	if err := server.Run(ctx); err != nil {
		return fmt.Errorf("failed to run the server: %w", err)
	}

	return nil
}
