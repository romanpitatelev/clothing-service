package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

const (
	readHeaderTimeoutValue = 3 * time.Second
	timeoutDuration        = 10 * time.Second
)

type Config struct {
	Port int
}

type usersHandler interface {
	CreateUser(w http.ResponseWriter, r *http.Request)
	LoginUser(w http.ResponseWriter, r *http.Request)
	GetUser(w http.ResponseWriter, r *http.Request)
	UpdateUser(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)
}

type tokenHandler interface {
	ValidateUser(w http.ResponseWriter, r *http.Request)
	RefreshToken(w http.ResponseWriter, r *http.Request)
	JWTAuth(next http.Handler) http.Handler
}

type clothesHandler interface {
	GetClothing(w http.ResponseWriter, r *http.Request)
}

type imagesHandler interface {
	GetImage(w http.ResponseWriter, r *http.Request)
}

type Server struct {
	cfg            Config
	server         *http.Server
	usersHandler   usersHandler
	tokenHandler   tokenHandler
	clothesHandler clothesHandler
	imagesHandler  imagesHandler
}

func New(
	cfg Config,
	userHandler usersHandler,
	tokenHandler tokenHandler,
	clothesHandler clothesHandler,
	imagesHandler imagesHandler,
) *Server { //nolint:whitespace
	router := chi.NewRouter()
	s := &Server{
		server: &http.Server{
			Addr:              fmt.Sprintf(":%d", cfg.Port),
			Handler:           router,
			ReadHeaderTimeout: readHeaderTimeoutValue,
		},
		cfg:            cfg,
		usersHandler:   userHandler,
		tokenHandler:   tokenHandler,
		clothesHandler: clothesHandler,
		imagesHandler:  imagesHandler,
	}

	router.Get("/metrics", promhttp.Handler().ServeHTTP)
	router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Use(middleware.Recoverer)

			r.Post("/users/register", s.usersHandler.CreateUser)
			r.Post("/users/{userId}/login", s.usersHandler.LoginUser)
			r.Post("/users/{userId}/otp", s.tokenHandler.ValidateUser)
			r.Post("/users/{userId}/refresh", s.tokenHandler.RefreshToken)

			r.Group(func(r chi.Router) {
				r.Use(s.tokenHandler.JWTAuth)
				r.Get("/users/{userId}", s.usersHandler.GetUser)
				r.Patch("/users/{userId}", s.usersHandler.UpdateUser)
				r.Delete("/users/{userId}", s.usersHandler.DeleteUser)

				r.Get("/clothes/{clothingId}", s.clothesHandler.GetClothing)

				r.Get("/images/{imageName}", s.imagesHandler.GetImage)
			})
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
