package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository/entities"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository"
)

func (p *Postgres) SaveUser(ctx context.Context, u entities.User) error {
	const op = "repository.postgres.SaveUser"

	var err error

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: %s", err, op)
	}

	defer func() {
		if err != nil {
			txErr := tx.Rollback()
			if txErr != nil {
				err = fmt.Errorf("tx err: %v, rollback err: %w", err, txErr)

				return
			}
		}

		txErr := tx.Commit()
		if txErr != nil {
			err = fmt.Errorf("commit err: %w", txErr)
		}
	}()

	q := `insert into users(id, email, access_token_google, refresh_token_google, refresh_token) 
	values ($1, $2, $3, $4, $5)`

	_, err = tx.ExecContext(ctx, q,
		u.ID,
		u.Email,
		u.AccessTokenGoogle,
		u.RefreshTokenGoogle,
		u.RefreshToken)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			return handleAuthError(pgErr, u)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	q = `insert into plans(id, name, prompt_limit, user_id) values ($1, $2, $3, $4)`

	planID := uuid.New()

	_, err = tx.ExecContext(ctx, q, planID, "free", 5, u.ID)
	if err != nil {
		return fmt.Errorf("%w: %s", err, op)
	}

	return nil
}

func (p *Postgres) GetUserByID(ctx context.Context) (entities.User, error) {
	const op = "repository.postgres.GetUserByID"

	const q = `select * from users where id = $1`
	var u user
	err := p.db.GetContext(ctx, &u, q, ctx.Value(entities.UserIDCtxKey{}).(uuid.UUID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entities.User{}, repository.NewNotFound(
				err,
				op,
				"ID",
				ctx.Value(entities.UserIDCtxKey{}).(uuid.UUID).String(),
			)
		}
		return entities.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return toUser(u), nil
}

func (p *Postgres) UpdateUser(ctx context.Context, u entities.User) error {
	const op = "repository.postgres.UpdateUser"

	const q = `update "users" 
	set access_token_google=$1, 
	refresh_token_google=$2, 
	refresh_token=$3 
	where email=$4`

	_, err := p.db.ExecContext(ctx, q, u.AccessTokenGoogle, u.RefreshTokenGoogle, u.RefreshToken, u.Email)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *Postgres) GetUserByEmail(ctx context.Context, email string) (entities.User, error) {
	const op = "repository.postgres.GetUserByEmail"

	q := `select * from users where email = $1`

	var u user

	err := p.db.GetContext(ctx, &u, q, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entities.User{}, repository.NewNotFound(err, op, "email", email)
		}

		return entities.User{}, fmt.Errorf("%s: %w", op, err)
	}

	q = `select * from plans where user_id = $1`

	err = p.db.GetContext(ctx, &u.Plan, q, u.ID)
	if err != nil {
		return entities.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return toUser(u), nil
}

func handleAuthError(err *pq.Error, u entities.User) error {
	switch err.Code.Name() {
	case uniqueViolation:
		switch {
		case strings.Contains(err.Message, "email"):
			return repository.NewUniqueViolation(err, "email already taken", "email", u.Email)
		}
	}
	return err
}
