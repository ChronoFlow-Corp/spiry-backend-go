package postgres

import (
	"context"
	"database/sql"
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

func (p *Postgres) GetUserByID(ctx context.Context, id string) (repository.User, error) {
	const op = "repository.postgres.GetUserByID"

	const q = `select * from users where id = $1`
	row := p.db.QueryRowContext(ctx, q, id)
	var u user
	err := row.Scan(&u)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.User{}, repository.NewNotFound(err, op, "ID", id)
		}
		return repository.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return toUser(u), nil
}

func handleAuthError(err *pq.Error, u repository.User) error {
	switch err.Code.Name() {
	case uniqueViolation:
		switch {
		case strings.Contains(err.Message, "email"):
			return repository.NewUniqueViolation(err, "email already taken", "email", u.Email())
		}
	}
	return err
}
