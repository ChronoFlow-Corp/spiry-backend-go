package rest

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strconv"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/config"
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
)

type Server struct {
	HTTPServer *http.Server
}

func NewHTTPServer(
	cfg *config.Config,
	r chi.Router,
	log *slog.Logger,
	lc fx.Lifecycle,
) *http.Server {
	srv := &http.Server{
		Addr:         cfg.HTTP.Addr + ":" + strconv.Itoa(cfg.HTTP.Port),
		IdleTimeout:  cfg.HTTP.Timeout,
		ReadTimeout:  cfg.HTTP.Timeout,
		WriteTimeout: cfg.HTTP.Timeout,
	}

	srv.Handler = r

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}

			log.Info("Listening on " + srv.Addr)

			go srv.Serve(ln)

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})

	return srv
}

func (s *Server) Start() error {
	return s.HTTPServer.ListenAndServe()
}
