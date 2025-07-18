package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/service"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/transport/server/handlers/google"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const readHeaderTimeout = time.Second * 5

// Server implement http transport.
type Server struct {
	s        *http.Server
	addr     string
	certFile string
	keyFile  string
	frontendURL string
	auth     service.Auth
}

// New creates new instance server struct.
func New(addr, certFile, keyFile, frontendURL string, port int, timeout time.Duration, auth service.Auth) Server {
	s := &http.Server{
		Addr:              addr + ":" + strconv.Itoa(port),
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      timeout,
		ReadTimeout:       timeout,
		IdleTimeout:       timeout,
	}

	return Server{s: s, certFile: certFile, frontendURL: frontendURL, addr: addr, keyFile: keyFile, auth: auth}
}

// ListenAndServe start listening port, if ssl credentials not provide listen on http.
func (s Server) ListenAndServe() error {
	const op = "server.ListenAndServe"

	frontendURL, err := url.Parse(s.frontendURL)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	s.setRoutes(frontendURL)

	if s.certFile != "" && s.keyFile != "" {
		err := s.s.ListenAndServeTLS(s.certFile, s.keyFile)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("%s: %w", op, err)
		}

		return nil
	}

	err = s.s.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// Shutdown stop http.Server.
func (s Server) Shutdown(ctx context.Context) error {
	const op = "server.Shutdown"

	err := s.s.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s Server) setRoutes(frontendURL *url.URL) {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowCredentials: true,
	}))
	router.Get("/api/connect/google", google.NewRedirect(s.auth))
	router.Get("/api/connect/google/callback", google.NewCallback(frontendURL, s.addr, s.auth))

	s.s.Handler = router
}
