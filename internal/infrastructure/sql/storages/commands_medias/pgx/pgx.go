package pgx

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	commandsmedia "github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/commands_medias"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/commands_medias/models"
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
var _ commandsmedia.CommandMediaStorage = (*Pgx)(nil)

func NewPgx(pool *pgxpool.Pool) *Pgx {
	return &Pgx{
		pool:   pool,
		getter: trmgr.DefaultCtxGetter,
	}
}

func (p *Pgx) Create(ctx context.Context, m entities.CommandMedia) error {
	const op = "storage.commands_medias.pgx.create"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.Insert(table).
		Columns(columns...).
		Values(m.ID, m.Name, m.Type, m.URL, m.Size, m.CommandID, m.UserID, m.CreatedAt, m.UpdatedAt).
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

func (p *Pgx) GetByID(ctx context.Context, mediaID uuid.UUID) (*entities.CommandMedia, error) {
	const op = "storage.commands_medias.pgx.GetByID"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	builder := sq.Select(table).Where(squirrel.Eq{columns[id]: mediaID})

	addUserIDWhere(ctx, builder)

	query, values, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	row := conn.QueryRow(ctx, query, values...)

	cm, err := scanToEntity(row)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &cm, nil
}

func (p *Pgx) Update(ctx context.Context, m entities.CommandMedia) error {
	const op = "storage.commands_medias.pgx.update"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	mDB, err := p.GetByID(ctx, m.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	update := sq.Update(table).Where(squirrel.Eq{columns[id]: m.ID})

	addUserIDWhere(ctx, update)

	if mDB.Name != m.Name {
		update.Set(columns[name], m.Name)
	}
	if mDB.Type != m.Type {
		update.Set(columns[mediaType], m.Type)
	}
	if mDB.URL != m.URL {
		update.Set(columns[indexUrl], m.URL)
	}
	if mDB.Size != m.Size {
		update.Set(columns[size], m.Size)
	}
	if mDB.CommandID != m.CommandID {
		update.Set(columns[commandID], m.CommandID)
	}

	update.Set(columns[updatedAt], m.UpdatedAt)

	query, args, err := update.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *Pgx) Delete(ctx context.Context, mediaID uuid.UUID) error {
	const op = "storage.commands_medias.pgx.Delete"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	builder := sq.Delete(table).Where(squirrel.Eq{columns[id]: mediaID})

	addUserIDWhere(ctx, builder)

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func scanToEntity(row pgx.Row) (entities.CommandMedia, error) {
	var cmm models.CommandMedia
	err := row.Scan(
		&cmm.ID,
		&cmm.Name,
		&cmm.Type,
		&cmm.URL,
		&cmm.Size,
		&cmm.CommandID,
		&cmm.UserID,
		&cmm.CreatedAt,
		&cmm.UpdatedAt,
	)
	if err != nil {
		return entities.CommandMedia{}, err
	}

	u, err := url.Parse(cmm.URL)
	if err != nil {
		return entities.CommandMedia{}, err
	}

	res := entities.CommandMedia{
		ID:        cmm.ID,
		Name:      cmm.Name,
		Type:      cmm.Type,
		URL:       *u,
		Size:      uint64(cmm.Size),
		CommandID: cmm.CommandID,
		UserID:    cmm.UserID,
		CreatedAt: cmm.CreatedAt,
		UpdatedAt: cmm.UpdatedAt,
	}

	return res, nil
}
