package pgx

import (
	"context"
	"errors"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/sessions"
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

var (
	sq                         = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	_  sessions.SessionStorage = (*Pgx)(nil)
)

func NewPgx(pool *pgxpool.Pool) *Pgx {
	return &Pgx{
		pool:   pool,
		getter: trmgr.DefaultCtxGetter,
	}
}

func (p *Pgx) Create(ctx context.Context, session entities.Session) error {
	const op = "storages.sessions.pgx.Create"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.Insert(table).Columns(columns...).Values(
		session.ID,
		session.Token,
		session.ExpiresAt,
		session.LastLogin,
		session.Device,
		session.UserID,
		session.CreatedAt,
		session.UpdatedAt,
	).ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = conn.Exec(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *Pgx) Update(ctx context.Context, session entities.Session) error {
	const op = "storages.sessions.pgx.Update"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	s, err := p.GetByID(ctx, session.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	update := sq.Update(table).
		Where(squirrel.Eq{columns[userID]: session.UserID, columns[id]: session.ID})

	if session.Token != s.Token {
		update = update.Set(columns[token], session.Token)
	}

	if session.ExpiresAt != s.ExpiresAt {
		update = update.Set(columns[expiresAt], session.ExpiresAt)
	}

	if session.LastLogin != s.LastLogin {
		update = update.Set(columns[lastLogin], session.LastLogin)
	}

	if session.Device != s.Device {
		update = update.Set(columns[device], session.Device)
	}

	update = update.Set(columns[updatedAt], session.UpdatedAt)

	query, values, err := update.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = conn.Exec(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *Pgx) GetByID(ctx context.Context, i uuid.UUID) (*entities.Session, error) {
	const op = "storages.sessions.pgx.GetByID"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.
		Select(columns...).
		From(table).
		Where(squirrel.Eq{columns[id]: i}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	row := conn.QueryRow(ctx, query, values...)

	var session entities.Session

	err = scanToEntity(row, &session)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w: %w", op, sessions.ErrNotFound, err)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &session, nil
}

func (p *Pgx) GetByToken(ctx context.Context, t string) (*entities.Session, error) {
	const op = "storages.sessions.pgx.GetByToken"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.
		Select(columns...).
		From(table).
		Where(squirrel.Eq{columns[token]: t}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	row := conn.QueryRow(ctx, query, values...)

	var session entities.Session

	err = scanToEntity(row, &session)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w: %w", op, sessions.ErrNotFound, err)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &session, nil
}

func (p *Pgx) GetByUserID(ctx context.Context, u uuid.UUID) ([]*entities.Session, error) {
	const op = "storages.s.pgx.GetByUserID"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.
		Select(columns...).
		From(table).
		Where(squirrel.Eq{columns[userID]: u}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := conn.Query(ctx, query, values...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	s := make([]*entities.Session, 0)

	for rows.Next() {
		var session entities.Session

		err := scanToEntity(rows, &session)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		s = append(s, &session)
	}

	return s, nil
}

func scanToEntity(row pgx.Row, session *entities.Session) error {
	err := row.Scan(
		&session.ID,
		&session.Token,
		&session.ExpiresAt,
		&session.LastLogin,
		&session.Device,
		&session.UserID,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}
