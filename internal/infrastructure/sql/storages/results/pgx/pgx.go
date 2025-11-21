package pgx

import (
	"context"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/results"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/results/models"
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
var _ results.ResultStorage = (*Pgx)(nil)

func NewPgx(pool *pgxpool.Pool) *Pgx {
	return &Pgx{
		pool:   pool,
		getter: trmgr.DefaultCtxGetter,
	}
}

func (p *Pgx) Create(ctx context.Context, res entities.Result) error {
	const op = "storages.results.pgx.Create"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.Insert(table).
		Columns(columns...).
		Values(
			res.ID,
			res.OpenRouterID,
			res.Text,
			res.CommandID,
			res.ToolID,
			res.ChatID,
			res.UserID,
			res.ModelID,
			res.CreatedAt,
			res.UpdatedAt,
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

func (p *Pgx) GetByCommandID(ctx context.Context, i uuid.UUID) (*entities.Result, error) {
	const op = "storages.results.pgx.GetByCommandID"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	builder := sq.Select(columns...).From(table).Where(squirrel.Eq{columns[commandID]: i})

	addUserIDWhere(ctx, builder)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	row := conn.QueryRow(ctx, query, args...)

	entity, err := scanToEntity(row)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &entity, nil
}

func scanToEntity(row pgx.Row) (entities.Result, error) {
	var res models.Result
	err := row.Scan(
		&res.ID,
		&res.OpenRouterID,
		&res.Text,
		&res.CommandID,
		&res.ToolID,
		&res.ChatID,
		&res.UserID,
		&res.ModelID,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err != nil {
		return entities.Result{}, err
	}

	entity := entities.Result{
		ID:           res.ID,
		OpenRouterID: res.OpenRouterID,
		Text:         res.Text,
		CommandID:    res.CommandID,
		ToolID:       res.ToolID,
		ChatID:       res.ChatID,
		UserID:       res.UserID,
		ModelID:      res.ModelID,
		CreatedAt:    res.CreatedAt,
		UpdatedAt:    res.UpdatedAt,
	}

	return entity, nil
}
