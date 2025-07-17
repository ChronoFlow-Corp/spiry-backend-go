package postgres

import (
	"context"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository"
)

func (p *Postgres) SaveUser(ctx context.Context, u repository.User) error {
	const op = "repository.postgres.SaveUser"

	const q = `insert into "users" (id, email, access_token_google, refresh_token_google, refresh_token) $1, $2, $3, $4, $5`

	_, err := p.db.ExecContext(ctx, q,
		u.ID(),
		u.Email(),
		u.AccessTokenGoogle(),
		u.RefreshTokenGoogle(),
		u.RefreshTokenGoogle())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
