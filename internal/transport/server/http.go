package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/service"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/service/chatting"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/transport/server/handlers/deleteChat"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/transport/server/handlers/getChats"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/transport/server/handlers/getUserInfo"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/transport/server/handlers/google"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/transport/server/handlers/patchChat"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/transport/server/handlers/refresh"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/transport/server/handlers/ws"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/transport/server/middlewares"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/jwt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

const readHeaderTimeout = time.Second * 5

// Server implement http transport.
type Server struct {
	s           *http.Server
	addr        string
	certFile    string
	keyFile     string
	frontendURL string
	devOrigin   string
	stageOrigin string
	prodOrigin  string
	auth        service.Auth
	ll          chatting.Service
	j           jwt.JWT
}

// New creates new instance server struct.
func New(addr, certFile, keyFile, frontendURL string,
	port int,
	timeout time.Duration,
	auth service.Auth,
	j jwt.JWT,
	ll chatting.Service,
	devOrigin,
	stageOrigin,
	prodOrigin string) Server {
	s := &http.Server{
		Addr:              addr + ":" + strconv.Itoa(port),
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      timeout,
		ReadTimeout:       timeout,
		IdleTimeout:       timeout,
	}

	return Server{
		s:           s,
		certFile:    certFile,
		frontendURL: frontendURL,
		addr:        addr,
		keyFile:     keyFile,
		auth:        auth,
		prodOrigin:  prodOrigin,
		devOrigin:   devOrigin,
		stageOrigin: stageOrigin,
		ll:          ll,
		j:           j,
	}
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
		slog.Default().Debug("HTTPS server listening on " + s.addr)
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
	router.Use(middlewares.Logger())
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowCredentials:   false,
		OptionsPassthrough: false,
		AllowedOrigins:     []string{s.stageOrigin, s.devOrigin, s.prodOrigin},
		AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:     []string{"Content-Type", "Authorization"}}))

	router.Route("/api", func(r chi.Router) {
		r.Route("/connect", func(r chi.Router) {
			r.Get("/google", google.NewRedirect(s.auth))
			r.Get("/google/callback", google.NewCallback(frontendURL, s.addr, s.auth))
		})
		r.Route("/", func(r chi.Router) {
			r.Use(middlewares.AuthJwt(s.j))
			r.Route("/user", func(r chi.Router) {
				r.Get("/", getUserInfo.New(s.auth))
				r.Get("/refresh", refresh.New(s.auth))
			})
			r.Get("/chats", getChats.New(s.ll))
			r.Patch("/chats", patchChat.New(s.ll))
			r.Delete("/chats", deleteChat.New(s.ll))
		})
		r.Get("/ws", ws.New(s.ll, s.j, s.stageOrigin, s.devOrigin, s.prodOrigin))
	})

	s.s.Handler = router
}
