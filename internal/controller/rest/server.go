package rest

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

const (
	ReadHeaderTimeoutValue = 3
	timeoutDuration        = 10 * time.Second
)

type Config struct {
	Port int
}

type Server struct {
	server       *http.Server
	usersHandler usersHandler
	port         int
	key          *rsa.PublicKey
}

type usersHandler interface {
	CreateUser(w http.ResponseWriter, t *http.Request)
	GetUser(w http.ResponseWriter, t *http.Request)
	UpdateUser(w http.ResponseWriter, t *http.Request)
	DeleteUser(w http.ResponseWriter, t *http.Request)
	GetUsers(w http.ResponseWriter, t *http.Request)
}

func New(cfg Config, userusersHandler usersHandler, key *rsa.PublicKey) *Server {
	router := chi.NewRouter()
	s := &Server{
		server: &http.Server{
			Addr:              fmt.Sprintf(":%d", cfg.Port),
			Handler:           router,
			ReadHeaderTimeout: ReadHeaderTimeoutValue * time.Second,
		},
		usersHandler: userusersHandler,
		port:         cfg.Port,
		key:          key,
	}

	router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Use(middleware.Recoverer)

			r.Post("/users", s.usersHandler.CreateUser)
			r.Get("/users/{userId}", s.usersHandler.GetUser)
			r.Patch("users/{userId}", s.usersHandler.UpdateUser)
			r.Delete("users/{userId}", s.usersHandler.DeleteUser)
			r.Get("/users", s.usersHandler.GetUsers)
		})
	})

	return s
}

func (s *Server) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()

		gracefulCtx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
		defer cancel()

		//nolint:contextcheck
		if err := s.server.Shutdown(gracefulCtx); err != nil {
			log.Warn().Err(err).Msg("failed to shutdown server")
		}
	}()

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start a server: %w", err)
	}

	return nil
}
