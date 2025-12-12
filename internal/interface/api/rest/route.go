package rest

import (
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/config"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/api/rest/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type RouteModule interface {
	Register(r chi.Router)
}

func NewServeMux(cfg *config.Config, routes []RouteModule) chi.Router {
	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Logger)
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.HTTP.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	}))
	mux.Use(middleware.Recoverer)
	mux.Use(middlewares.Logger())

	mux.Route("/api", func(r chi.Router) {
		for _, route := range routes {
			route.Register(r)
		}
	})

	return mux
}
