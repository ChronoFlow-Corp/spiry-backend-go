package pgx

import (
	"context"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	unloggedusers "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/unlogged_users"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/unlogged_users/models"
	"github.com/Masterminds/squirrel"
	trmgr "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pgx struct {
	pool   *pgxpool.Pool
	getter *trmgr.CtxGetter
}

var (
	sq                                   = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	_  unloggedusers.UnloggedUserStorage = (*Pgx)(nil)
)

func NewPgx(pool *pgxpool.Pool) *Pgx {
	return &Pgx{
		pool:   pool,
		getter: trmgr.DefaultCtxGetter,
	}
}

func (p *Pgx) Create(ctx context.Context, u *entities.UnloggedUser) error {
	const op = "storages.unlogged_users.pgx.Create"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.Insert(table).
		Columns(columns...).
		Values(u.ID, u.IP, u.PlanID, u.CreatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = conn.Exec(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *Pgx) GetByIP(ctx context.Context, ip string) (*entities.UnloggedUser, error) {
	const op = "storages.unlogged_users.pgx.GetByIP"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.Select(columns...).
		From(table).
		Where(squirrel.Eq{columns[ip_address]: ip}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	row := conn.QueryRow(ctx, query, values...)

	mod, err := scanToEntity(row)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &mod, nil
}

func scanToEntity(row pgx.Row) (entities.UnloggedUser, error) {
	var mod models.UnloggedUser

	err := row.Scan(&mod.ID, &mod.IP, &mod.PlanID, &mod.CreatedAt)
	if err != nil {
		return entities.UnloggedUser{}, err
	}

	return entities.UnloggedUser{
		ID:        mod.ID,
		IP:        mod.IP,
		PlanID:    mod.PlanID,
		CreatedAt: mod.CreatedAt,
	}, nil
}
