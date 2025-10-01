package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/config"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository/postgres"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/service"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/service/chatting"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/transport/server"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/jwt"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/llm"
	"github.com/golang-migrate/migrate/v4"
	ps "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)


func main() {
	cfg := config.Config{}
	cfg.MustLoad()

	steps := flag.Int("steps", 0, "number of steps to migrate, positive to migrate up, negative to migrate down")
	flag.Parse()

	if steps != nil && *steps != 0 {
		migrating(cfg, *steps)
	}

	setLogger(cfg.Env)

	db, err := postgres.New(cfg.Database.PostgresHost,
		cfg.Database.PostgresPort,
		cfg.Database.PostgresUser,
		cfg.Database.PostgresPassword,
		cfg.Database.PostgresDatabase)
	if err != nil {
		panic(err)
	}

	j := jwt.New([]byte(cfg.JWT.AccessSecretPrivate),
		[]byte(cfg.JWT.AccessSecretPublic),
		[]byte(cfg.JWT.RefreshSecret),
		cfg.JWT.AccessExpire,
		cfg.JWT.RefreshExpire)

	auth := service.New(cfg.GoogleAuth.ClientID, cfg.GoogleAuth.ClientSecret,
		"http://localhost:1337/api/connect/google/callback", db, j)

	llmUrl, err := url.Parse(cfg.LLM.URL)
	if err != nil {
		panic("invalid LLM URL")
	}

	chat := chatting.NewChat(llm.NewClient(cfg.LLM.Key, llmUrl), db, db, db, db, db)

	srv := server.New(
		cfg.HTTP.Addr,
		cfg.HTTP.CertFile,
		cfg.HTTP.KeyFile,
		cfg.HTTP.FrontendURL,
		cfg.HTTP.Port,
		cfg.HTTP.Timeout,
		auth,
		j,
		chat)

	go func() {
		fmt.Println("Starting server...")
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

func setLogger(level string) {
	switch level {
	case "development":
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
		return
	case "production":
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
		return
	}
}

func migrating(cfg config.Config, steps int) {
	const op = "migrating"

	db, err := sql.Open("postgres",
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			cfg.Database.PostgresUser,
			cfg.Database.PostgresPassword,
			cfg.Database.PostgresHost,
			cfg.Database.PostgresPort,
			cfg.Database.PostgresDatabase,
			"disable"))
	if err != nil {
		slog.Error(op, slog.String("err", err.Error()))
		os.Exit(1)
	}

	driver, err := ps.WithInstance(db, &ps.Config{})
	if err != nil {
		slog.Error(op, slog.String("err", err.Error()))
		os.Exit(1)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		slog.Error(op, slog.String("err", err.Error()))
		os.Exit(1)
	}
	slog.Info(op, slog.Int("steps", steps))

	err = m.Steps(steps)
	if err != nil {
		slog.Error(op, slog.String("err", err.Error()))
	} else {
		slog.Info("Migration completed successfully")
	}
}

