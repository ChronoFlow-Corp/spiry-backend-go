package main

import (
	"context"
	"fmt"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/config"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository/postgres"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/service"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/transport/server"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/jwt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.Config{}
	cfg.MustLoad()

	db, err := postgres.New(cfg.Database.PostgresHost,
		cfg.Database.PostgresPort,
		cfg.Database.PostgresUser,
		cfg.Database.PostgresPassword,
		cfg.Database.PostgresDatabase)
	if err != nil {
		panic(err)
	}

	j := jwt.New([]byte(cfg.JWT.AccessSecretPrivate),
		[]byte(cfg.JWT.AccessSecretPrivate),
		[]byte(cfg.JWT.RefreshSecret),
		cfg.JWT.AccessExpire,
		cfg.JWT.RefreshExpire)

	auth := service.New(cfg.GoogleAuth.ClientID, cfg.GoogleAuth.ClientSecret,
		"http://localhost:1337/api/connect/google/callback", db, j)

	srv := server.New(cfg.HTTP.Addr, cfg.HTTP.CertFile, cfg.HTTP.KeyFile, cfg.HTTP.FrontendURL, cfg.HTTP.Port, cfg.HTTP.Timeout, auth)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("ListenAndServe(): %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v", err)
	}
}
