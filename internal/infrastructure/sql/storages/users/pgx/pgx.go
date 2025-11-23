package pgx

import (
	"context"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/users"
	"github.com/Masterminds/squirrel"
	trmgr "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pgx struct {
	pool   *pgxpool.Pool
	getter *trmgr.CtxGetter
}

var sq = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
var _ users.UserStorage = (*Pgx)(nil)

func NewPgx(pool *pgxpool.Pool) *Pgx {
	return &Pgx{
		pool:   pool,
		getter: trmgr.DefaultCtxGetter,
	}
}

func (p *Pgx) Create(ctx context.Context, user entities.User) error {
	const op = "storages.users.pgx.NewUser"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.Insert("users").Columns(columns...).Values(
		user.ID,
		user.Email,
		user.AvatarURL,
		user.Name,
		user.LastName,
		user.Theme,
		user.GoogleAccessToken,
		user.GoogleRefreshToken,
		user.CreatedAt,
		user.UpdatedAt,
	).ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = conn.Exec(ctx, query, values...)
	if err != nil {
		return handlePgxError(op, err)
	}

	return nil
}

func (p *Pgx) GetByID(ctx context.Context, userId uuid.UUID) (*entities.User, error) {
	const op = "storages.users.pgx.GetByID"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.
		Select(columns...).
		From(table).
		Where(squirrel.Eq{columns[id]: userId}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	row := conn.QueryRow(ctx, query, values...)

	user := &entities.User{}

	err = scanToEntity(row, user)
	if err != nil {
		return nil, handlePgxError(op, err)
	}

	return user, nil
}

func (p *Pgx) GetByEmail(ctx context.Context, e string) (*entities.User, error) {
	const op = "storages.users.pgx.GetByEmail"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.
		Select(columns...).
		From(table).
		Where(squirrel.Eq{columns[email]: e}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	row := conn.QueryRow(ctx, query, values...)

	user := &entities.User{}

	err = scanToEntity(row, user)
	if err != nil {
		return nil, handlePgxError(op, err)
	}

	return user, nil
}

func (p *Pgx) Update(ctx context.Context, user entities.User) error {
	const op = "storages.users.pgx.Update"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	userDB, err := p.GetByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	update := sq.Update(table).Where(squirrel.Eq{columns[id]: user.ID})

	if userDB.Name != user.Name {
		update = update.Set(columns[name], user.Name)
	}
	if userDB.Email != user.Email {
		update = update.Set(columns[email], user.Email)
	}
	if userDB.LastName != user.LastName {
		update = update.Set(columns[lastName], user.LastName)
	}
	if userDB.AvatarURL != user.AvatarURL {
		update = update.Set(columns[avatarURL], user.AvatarURL)
	}
	if userDB.Theme != user.Theme {
		update = update.Set(columns[theme], user.Theme)
	}
	if userDB.GoogleAccessToken != user.GoogleAccessToken {
		update = update.Set(columns[googleAccessToken], user.GoogleAccessToken)
	}
	if userDB.GoogleRefreshToken != user.GoogleRefreshToken {
		update = update.Set(columns[googleRefreshToken], user.GoogleRefreshToken)
	}

	update = update.Set(columns[updatedAt], user.UpdatedAt)

	query, values, err := update.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = conn.Exec(ctx, query, values...)
	if err != nil {
		return handlePgxError(op, err)
	}

	return nil
}

func scanToEntity(row pgx.Row, user *entities.User) error {
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.AvatarURL,
		&user.Name,
		&user.LastName,
		&user.Theme,
		&user.GoogleAccessToken,
		&user.GoogleRefreshToken,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}
