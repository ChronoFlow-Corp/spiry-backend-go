package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

type Database struct {
	pool *pgxpool.Pool
	env  string
}

var sq = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

// New create new database instance with pgxpool pool.
func New(ctx context.Context, host, port, user, dbname, password, env string) *Database {
	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		host, port, user, dbname, password)

	pool, err := newPoolWithTracing(ctx, dsn)
	if err != nil {
		panic("failed to connect to database: " + dsn)
	}

	return &Database{
		pool: pool,
		env:  env,
	}
}

func newPoolWithTracing(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	cfg.MinConns = 4
	cfg.MaxConns = 64

	// connections policy
	cfg.MaxConnIdleTime = 10 * time.Minute
	cfg.MaxConnLifetime = 2 * time.Hour
	cfg.HealthCheckPeriod = 1 * time.Minute

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger: tracelog.LoggerFunc(func(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
			logger.Info(msg, "level", level, "data", data)
		}),
		LogLevel: tracelog.LogLevelInfo,
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return pool, nil
}
