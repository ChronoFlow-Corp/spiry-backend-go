package pgx

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/commands"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/commands/models"
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
var _ commands.CommandStorage = (*Pgx)(nil)

func NewPgx(pool *pgxpool.Pool) *Pgx {
	return &Pgx{
		pool:   pool,
		getter: trmgr.DefaultCtxGetter,
	}
}

func (p *Pgx) Create(ctx context.Context, command entities.Command) error {
	const op = "storages.commands.pgx.Create"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.
		Insert(table).
		Columns(columns...).
		Values(
			command.ID,
			command.Prompt,
			command.Settings,
			command.Flags,
			command.ChatID,
			command.ToolID,
			command.ModelID,
			command.Status,
			command.UserID,
			command.CreatedAt,
			command.UpdatedAt,
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

func (p *Pgx) GetByID(ctx context.Context, commandID uuid.UUID) (*entities.Command, error) {
	const op = "storages.commands.pgx.GetByID"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	builder := sq.
		Select(columns...).
		From(table).
		Where(squirrel.Eq{columns[id]: commandID})

	addUserIDWhere(ctx, builder)

	query, values, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	row := conn.QueryRow(ctx, query, values...)

	cmd, err := scanToEntity(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w: %w", op, commands.ErrNotFound, err)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &cmd, nil
}

func (p *Pgx) Update(ctx context.Context, command entities.Command) error {
	const op = "storages.commands.pgx.Update"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	cmdDB, err := p.GetByID(ctx, command.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	update := sq.Update(table).Where(squirrel.Eq{columns[id]: command.ID})

	addUserIDWhere(ctx, update)

	if command.Prompt != cmdDB.Prompt {
		update = update.Set(columns[prompt], command.Prompt)
	}
	if command.Settings != nil {
		update = update.Set(columns[settings], command.Settings)
	}
	if !slices.Equal(command.Flags, cmdDB.Flags) {
		update = update.Set(columns[flags], command.Flags)
	}
	if command.Status != cmdDB.Status {
		update = update.Set(columns[status], command.Status)
	}

	update = update.Set(columns[modelID], command.ModelID)
	update = update.Set(columns[toolID], command.ToolID)
	update = update.Set(columns[updatedAt], command.UpdatedAt)

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

func (p *Pgx) Delete(ctx context.Context, commandID uuid.UUID) error {
	const op = "storages.commands.pgx.Delete"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	builder := sq.Delete(table).Where(squirrel.Eq{columns[id]: commandID})

	addUserIDWhere(ctx, builder)

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

func scanToEntity(row pgx.Row) (entities.Command, error) {
	var command models.Command
	err := row.Scan(
		&command.ID,
		&command.Prompt,
		&command.Settings,
		&command.Flags,
		&command.ChatID,
		&command.ToolID,
		&command.ModelID,
		&command.Status,
		&command.UserID,
		&command.CreatedAt,
		&command.UpdatedAt,
	)
	if err != nil {
		return entities.Command{}, err
	}

	cmd := entities.Command{
		ID:        command.ID,
		Prompt:    command.Prompt,
		Settings:  command.Settings,
		Flags:     command.Flags,
		ChatID:    command.ChatID,
		ToolID:    command.ToolID,
		ModelID:   command.ModelID,
		Status:    command.Status,
		UserID:    command.UserID,
		CreatedAt: command.CreatedAt,
		UpdatedAt: command.UpdatedAt,
	}

	return cmd, nil
}
