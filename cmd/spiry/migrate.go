package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	const op = "cmd.migrate"

	steps := flag.Int("steps", 0, "number of steps to migrate, positive to migrate up, negative to migrate down")
	flag.Parse()

	cfg := config.Config{}
	cfg.MustLoad()


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

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		slog.Error(op, slog.String("err", err.Error()))
		os.Exit(1)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		slog.Error(op, slog.String("err", err.Error()))
		os.Exit(1)
	}
	slog.Info(op, slog.Int("steps", *steps))


	err = m.Steps(*steps)
	if err != nil {
		slog.Error(op, slog.String("err", err.Error()))
		os.Exit(1)
	}

	slog.Info("Migration completed successfully")
	os.Exit(0)
}
