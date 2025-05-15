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

	brnadshandler "github.com/romanpitatelev/clothing-service/internal/controller/rest/brands-handler"
	clotheshandler "github.com/romanpitatelev/clothing-service/internal/controller/rest/clothes-handler"
	fileshandler "github.com/romanpitatelev/clothing-service/internal/controller/rest/files-handler"
	iamhandler "github.com/romanpitatelev/clothing-service/internal/controller/rest/iam-handler"
	usershandler "github.com/romanpitatelev/clothing-service/internal/controller/rest/users-handler"
)

const (
	readHeaderTimeoutValue = 3 * time.Second
	timeoutDuration        = 10 * time.Second
)

type Config struct {
	Port int
}

type Server struct {
	cfg            Config
	server         *http.Server
	usersHandler   *usershandler.Handler
	tokenHandler   *iamhandler.Handler
	clothesHandler *clotheshandler.Handler
	imagesHandler  *fileshandler.Handler
	brandsHandler  *brnadshandler.Handler
}

func New(
	cfg Config,
	usersHandler *usershandler.Handler,
	tokenHandler *iamhandler.Handler,
	clothesHandler *clotheshandler.Handler,
	imagesHandler *fileshandler.Handler,
	brandsHandler *brnadshandler.Handler,
) *Server { //nolint:whitespace
	router := chi.NewRouter()
	s := &Server{
		server: &http.Server{
			Addr:              fmt.Sprintf(":%d", cfg.Port),
			Handler:           router,
			ReadHeaderTimeout: readHeaderTimeoutValue,
		},
		cfg:            cfg,
		usersHandler:   usersHandler,
		tokenHandler:   tokenHandler,
		clothesHandler: clothesHandler,
		imagesHandler:  imagesHandler,
		brandsHandler:  brandsHandler,
	}

	router.Use(middleware.Recoverer)
	router.Get("/metrics", promhttp.Handler().ServeHTTP)
	router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
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

				r.Get("/brands", s.brandsHandler.ListBrands)
				r.Get("/users/{userId}/brands", s.brandsHandler.ListPreferredBrands)
				r.Post("/user/{userId}/brands", s.brandsHandler.SetPreferredBrands)
				r.Put("/user/{userId}/brands", s.brandsHandler.RequestBrand)
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
