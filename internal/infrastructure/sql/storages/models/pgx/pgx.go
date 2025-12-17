package pgx

import (
	"context"
	"errors"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/models"
	models2 "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/models/models"
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
	sq                     = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	_  models.ModelStorage = (*Pgx)(nil)
)

func NewPgx(pool *pgxpool.Pool) *Pgx {
	return &Pgx{
		pool:   pool,
		getter: trmgr.DefaultCtxGetter,
	}
}

func (p *Pgx) Create(ctx context.Context, plan entities.Model) error {
	const op = "storages.models.pgx.Create"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.
		Insert(table).
		Columns(columns...).
		Values(
			plan.ID,
			plan.Name,
			plan.MinLevel,
			plan.CreatedAt,
			plan.UpdatedAt,
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

func (p *Pgx) GetByName(ctx context.Context, n string) (*entities.Model, error) {
	const op = "storages.models.pgx.GetByName"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.
		Select(columns...).
		From(table).
		Where(squirrel.Eq{columns[name]: n}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	row := conn.QueryRow(ctx, query, values...)

	model, err := scanToEntity(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w: %w", op, models.ErrNotFound, err)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &model, nil
}

func (p *Pgx) GetAll(ctx context.Context) ([]entities.Model, error) {
	const op = "storages.m.pgx.GetAll"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, _, err := sq.Select(columns...).From(table).ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	m := make([]entities.Model, 0)

	for rows.Next() {
		model, err := scanToEntity(rows)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		m = append(m, model)
	}

	return m, nil
}

func scanToEntity(row pgx.Row) (entities.Model, error) {
	var model models2.Model

	err := row.Scan(
		&model.ID,
		&model.Name,
		&model.MinLevel,
		&model.CreatedAt,
		&model.UpdatedAt,
	)
	if err != nil {
		return entities.Model{}, err
	}

	modelEntity := entities.Model{
		ID:        model.ID,
		Name:      model.Name,
		MinLevel:  model.MinLevel,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}

	return modelEntity, nil
}
