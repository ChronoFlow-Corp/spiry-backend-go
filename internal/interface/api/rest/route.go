package rest

import (
	"fmt"
	"net/url"

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
	u, err := url.Parse(cfg.HTTP.FrontendUrl)
	if err != nil {
		panic("invalid frontend URL " + err.Error())
	}

	var origin string

	if cfg.Env == "development" {
		origin = fmt.Sprintf("http://%s", u.Host)
	}

	if cfg.Env == "production" {
		origin = fmt.Sprintf("https://%s", u.Host)
	}

	if origin == "" {
		panic("invalid frontend URL " + cfg.HTTP.FrontendUrl)
	}

	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Logger)
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{origin},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"Access-Control-Allow-Origin",
		},
		AllowCredentials: true,
		Debug:            true,
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
