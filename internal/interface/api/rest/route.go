package rest

import (
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/interface/api/rest/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type RouteModule interface {
	Register(r chi.Router)
}

func NewServeMux(routes []RouteModule) chi.Router {
	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(middlewares.Logger())

	mux.Route("/api", func(r chi.Router) {
		for _, route := range routes {
			route.Register(r)
		}
	})

	return mux
}
