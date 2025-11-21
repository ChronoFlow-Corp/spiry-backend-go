package pgx

import (
	"context"
	"errors"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/tools"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/tools/models"
	"github.com/Masterminds/squirrel"
	trmgr "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pgx struct {
	pool   *pgxpool.Pool
	getter *trmgr.CtxGetter
}

var sq = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
var _ tools.ToolStorage = (*Pgx)(nil)

func NewPgx(pool *pgxpool.Pool) *Pgx {
	return &Pgx{
		pool:   pool,
		getter: trmgr.DefaultCtxGetter,
	}
}

func (p *Pgx) Create(ctx context.Context, t entities.Tool) error {
	const op = "storages.tools.pgx.Create"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.Insert(table).
		Columns(columns...).
		Values(t.ID, t.Name, t.Settings, t.Prompt, t.MinLevel, t.CreatedAt, t.UpdatedAt).
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

func (p *Pgx) GetByName(ctx context.Context, toolName string) (*entities.Tool, error) {
	const op = "storages.tools.pgx.GetByName"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.Select(columns...).
		From(table).
		Where(squirrel.Eq{columns[name]: toolName}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	row := conn.QueryRow(ctx, query, values...)

	t, err := scanToEntity(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w: %w", op, tools.ErrNotFound, err)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &t, nil
}

func (p *Pgx) GetAll(ctx context.Context) ([]*entities.Tool, error) {
	const op = "storages.tools.pgx.GetAll"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.Select(columns...).From(table).ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := conn.Query(ctx, query, values...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	ts := make([]*entities.Tool, 0)
	for rows.Next() {
		t, err := scanToEntity(rows)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		ts = append(ts, &t)
	}

	return ts, nil
}

func scanToEntity(row pgx.Row) (entities.Tool, error) {
	var t models.Tool
	err := row.Scan(
		&t.ID,
		&t.Name,
		&t.Modalities,
		&t.Settings,
		&t.Prompt,
		&t.MinLevel,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		return entities.Tool{}, err
	}

	mod := make([]entities.Modality, len(t.Modalities))

	for i, m := range t.Modalities {
		mod[i] = entities.Modality(m)
	}

	return entities.Tool{
		ID:         t.ID,
		Name:       t.Name,
		Modalities: mod,
		Settings:   t.Settings,
		Prompt:     t.Prompt,
		MinLevel:   uint(t.MinLevel),
		CreatedAt:  t.CreatedAt,
		UpdatedAt:  t.UpdatedAt,
	}, nil
}
