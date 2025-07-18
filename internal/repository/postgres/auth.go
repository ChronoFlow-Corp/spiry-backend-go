package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"strings"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository"
)

func (p *Postgres) SaveUser(ctx context.Context, u repository.User) error {
	const op = "repository.postgres.SaveUser"

	const q = `insert into "users" (id, email, access_token_google, refresh_token_google, refresh_token) 
	values ($1, $2, $3, $4, $5)`

	_, err := p.db.ExecContext(ctx, q,
		u.ID(),
		u.Email(),
		u.AccessTokenGoogle(),
		u.RefreshTokenGoogle(),
		u.RefreshToken())
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			return handleAuthError(pgErr, u)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func handleAuthError(err *pq.Error, u repository.User) error {
	fmt.Println(err.Code.Name())
	switch err.Code.Name() {
	case uniqueViolation:
		switch {
		case strings.Contains(err.Message, "email"):
			return repository.NewUniqueViolation(err, "email already taken", "email", u.Email())
		}
	}
	return err
}
