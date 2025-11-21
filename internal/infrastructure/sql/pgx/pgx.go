package pgx

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/config"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

func NewPool(cfg *config.Config) *pgxpool.Pool {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		cfg.Database.PostgresHost,
		cfg.Database.PostgresPort,
		cfg.Database.PostgresUser,
		cfg.Database.PostgresDatabase,
		cfg.Database.PostgresPassword,
	)

	pool, err := newPoolWithTracing(context.Background(), dsn)
	if err != nil {
		panic("failed to connect to database: " + dsn)
	}

	if err := pool.Ping(context.Background()); err != nil {
		panic(fmt.Sprintf("failed to ping database: %s; err: %v", dsn, err))
	}

	return pool
}

func NewManager(pool *pgxpool.Pool) *manager.Manager {
	return manager.Must(trmpgx.NewDefaultFactory(pool))
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
		Logger: tracelog.LoggerFunc(func(
			ctx context.Context,
			level tracelog.LogLevel,
			msg string,
			data map[string]any,
		) {
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
