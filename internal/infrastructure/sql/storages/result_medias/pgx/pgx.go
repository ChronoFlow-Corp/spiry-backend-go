package pgx

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/models"
	resultmedias "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/result_medias"
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
var _ resultmedias.ResultMediaStorage = (*Pgx)(nil)

func NewPgx(pool *pgxpool.Pool) *Pgx {
	return &Pgx{
		pool:   pool,
		getter: trmgr.DefaultCtxGetter,
	}
}

func (p *Pgx) Create(ctx context.Context, result entities.ResultMedia) error {
	const op = "storages.result_medias.Create"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.Insert(table).
		Columns(columns...).
		Values(
			result.ID,
			result.Name,
			result.Type,
			result.URL,
			result.Size,
			result.ResultID,
			result.UserID,
			result.CreatedAt,
			result.UpdatedAt,
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

func (p *Pgx) GetByID(ctx context.Context, cmID uuid.UUID) (*entities.ResultMedia, error) {
	const op = "storages.result_medias.GetByID"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	builder := sq.Select(columns...).From(table).Where(squirrel.Eq{columns[id]: cmID})

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

func (p *Pgx) GetByURL(ctx context.Context, url string) (*entities.ResultMedia, error) {
	const op = "storages.result_medias.GetByURL"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	builder := sq.Select(columns...).From(table).Where(squirrel.Eq{columns[columnURL]: url})

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

func (p *Pgx) GetByResultID(ctx context.Context, i uuid.UUID) ([]*entities.ResultMedia, error) {
	const op = "storages.result_medias.GetByResultID"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	builder := sq.Select(columns...).From(table).Where(squirrel.Eq{columns[resultID]: i})

	addUserIDWhere(ctx, builder)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	entityList := make([]*entities.ResultMedia, 0)

	for rows.Next() {
		entity, err := scanToEntity(rows)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		entityList = append(entityList, &entity)
	}

	return entityList, nil
}

func (p *Pgx) Update(ctx context.Context, rm entities.ResultMedia) error {
	const op = "storages.result_medias.Update"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	rmDB, err := p.GetByID(ctx, rm.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	builder := sq.Update(table).Where(squirrel.Eq{columns[id]: rmDB.ID})

	addUserIDWhere(ctx, builder)

	if rmDB.Name != rm.Name {
		builder.Set(columns[name], rm.Name)
	}
	if rmDB.URL != rm.URL {
		builder.Set(columns[columnURL], rm.URL)
	}
	if rmDB.Size != rm.Size {
		builder.Set(columns[size], rm.Size)
	}
	if rmDB.ResultID != rm.ResultID {
		builder.Set(columns[resultID], rm.ResultID)
	}

	builder.Set(columns[updatedAt], rm.UpdatedAt)

	query, values, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = conn.Exec(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func scanToEntity(row pgx.Row) (entities.ResultMedia, error) {
	var result models.ResultMedia
	err := row.Scan(
		&result.ID,
		&result.Name,
		&result.Type,
		&result.URL,
		&result.Size,
		&result.ResultID,
		&result.UserID,
		&result.CreatedAt,
		&result.UpdatedAt)
	if err != nil {
		return entities.ResultMedia{}, err
	}

	u, err := url.Parse(result.URL)
	if err != nil {
		return entities.ResultMedia{}, err
	}

	entity := entities.ResultMedia{
		ID:        result.ID,
		Name:      result.Name,
		Type:      result.Type,
		URL:       u,
		Size:      int64(result.Size),
		ResultID:  result.ResultID,
		UserID:    result.UserID,
		CreatedAt: result.CreatedAt,
		UpdatedAt: result.UpdatedAt,
	}

	return entity, nil
}
