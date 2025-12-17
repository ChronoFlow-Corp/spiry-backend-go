package pgx

import (
	"errors"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/plans"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	errCodeUniqueViolation = "23505"
)

func handlePgxError(op string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: %w: %w", op, plans.ErrNotFound, err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == errCodeUniqueViolation {
			return fmt.Errorf("%s: %w: %s", op, plans.ErrAlreadyExists, pgErr.Message)
		}

		return fmt.Errorf("%s: database error: %w", op, err)
	}

	return fmt.Errorf("%s: %w", op, err)
}
