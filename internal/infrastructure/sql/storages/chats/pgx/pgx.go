package pgx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/aggregates"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	domainModels "github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/models"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/chats"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/infrastructure/sql/storages/chats/models"
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
	sq                   = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	_  chats.ChatStorage = (*Pgx)(nil)
)

func NewPgx(pool *pgxpool.Pool) *Pgx {
	return &Pgx{
		pool:   pool,
		getter: trmgr.DefaultCtxGetter,
	}
}

func (p *Pgx) Create(ctx context.Context, chat entities.Chat) error {
	const op = "storages.chats.Create"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	query, values, err := sq.Insert(table).
		Columns(columns...).
		Values(chat.ID, chat.Title, chat.CreatedAt, chat.UpdatedAt, chat.UserID).
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

func (p *Pgx) GetByID(ctx context.Context, chatID uuid.UUID) (*entities.Chat, error) {
	const op = "storages.chats.GetByID"

	builder := sq.Select(columns...).From(table).Where(squirrel.Eq{columns[id]: chatID})

	addUserIDWhere(ctx, builder)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	row := conn.QueryRow(ctx, query, args...)

	chat, err := scanToEntity(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w: %w", op, chats.ErrNotFound, err)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &chat, nil
}

func (p *Pgx) Update(ctx context.Context, chat entities.Chat) error {
	const op = "storages.chats.Update"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	chatDB, err := p.GetByID(ctx, chat.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	builder := sq.Update(table).Where(squirrel.Eq{columns[id]: chat.ID})

	addUserIDWhere(ctx, builder)

	if chatDB.Title != chat.Title {
		builder = builder.Set(columns[title], chat.Title)
	}

	builder = builder.Set(columns[updatedAt], chat.UpdatedAt)

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

func (p *Pgx) Delete(ctx context.Context, chatID uuid.UUID) error {
	const op = "storages.chats.Delete"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	builder := sq.Delete(table).Where(squirrel.Eq{columns[id]: chatID})

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

func (p *Pgx) GetAll(ctx context.Context) ([]*entities.Chat, error) {
	const op = "storages.chats.GetAll"

	if _, ok := getUserId(ctx); !ok {
		return nil, chats.ErrNotFound
	}

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	builder := sq.Select(columns...).From(table)

	addUserIDWhere(ctx, &builder)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	fmt.Println(query, args, "ONE!!!!!!!!!!!!!!!1")

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w: %w", op, chats.ErrNotFound, err)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	ch := make([]*entities.Chat, 0)
	for rows.Next() {
		chat, err := scanToEntity(rows)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		ch = append(ch, &chat)
	}

	return ch, nil
}

func (p *Pgx) GetByIDWithCommandsResults(
	ctx context.Context,
	chatID uuid.UUID,
) (*aggregates.Chat, error) {
	const op = "storages.chats.GetByIDWithCommandsResults"

	conn := p.getter.DefaultTrOrDB(ctx, p.pool)

	uID, _ := ctx.Value(domainModels.UserIDCtxKey{}).(uuid.UUID)

	row := conn.QueryRow(ctx, chatWithMediasQuery, chatID, uID)

	agg, err := scanToAggregate(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, chats.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &agg, nil
}

func scanToEntity(row pgx.Row) (entities.Chat, error) {
	var chat models.Chat
	err := row.Scan(&chat.ID, &chat.Title, &chat.CreatedAt, &chat.UpdatedAt, &chat.UserID)
	if err != nil {
		return entities.Chat{}, err
	}

	entity := entities.Chat{
		ID:        chat.ID,
		Title:     chat.Title,
		CreatedAt: chat.CreatedAt,
		UpdatedAt: chat.UpdatedAt,
		UserID:    chat.UserID,
	}

	return entity, nil
}

func scanToAggregate(rows pgx.Row) (aggregates.Chat, error) {
	var chat models.ChatWithCouples
	err := rows.Scan(
		&chat.ID,
		&chat.Title,
		&chat.CreatedAt,
		&chat.UpdatedAt,
		&chat.UserID,
		&chat.CommandResultCouple,
	)
	if err != nil {
		return aggregates.Chat{}, err
	}

	var couple []models.Couple
	err = json.Unmarshal(chat.CommandResultCouple, &couple)
	if err != nil {
		return aggregates.Chat{}, err
	}

	agg := aggregates.Chat{
		Chat: &entities.Chat{
			ID:        chat.ID,
			Title:     chat.Title,
			CreatedAt: chat.CreatedAt,
			UpdatedAt: chat.UpdatedAt,
			UserID:    chat.UserID,
		},
		Couples: make([]aggregates.CommandResultCouple, 0),
	}

	for _, c := range couple {
		agg.Couples = append(agg.Couples, aggregates.CommandResultCouple{
			Command:      mapCommand(c.Command),
			CommandMedia: mapCommandMedias(c.Command),
			Result:       mapResult(c.Result[0]),
			ResultMedia:  mapResultMedia(c.Result[0]),
			Model:        mapModel(c.Result[0]),
		})
	}

	return agg, nil
}

func mapCommand(cm models.CommandWithMedias) *entities.Command {
	command := entities.Command{
		ID:        cm.ID,
		Prompt:    cm.Prompt,
		Settings:  cm.Settings,
		Flags:     cm.Flags,
		ChatID:    cm.ChatID,
		ToolID:    cm.ToolID,
		ModelID:   cm.ModelID,
		Status:    cm.Status,
		UserID:    cm.UserID,
		CreatedAt: cm.CreatedAt,
		UpdatedAt: cm.UpdatedAt,
	}

	return &command
}

func mapCommandMedias(cm models.CommandWithMedias) []*entities.CommandMedia {
	commandMedias := make([]*entities.CommandMedia, 0)
	for _, m := range cm.Medias {
		u, err := url.Parse(m.URL)
		if err != nil {
			panic(err)
		}

		commandMedias = append(commandMedias, &entities.CommandMedia{
			ID:        m.ID,
			Name:      m.Name,
			Type:      m.Type,
			URL:       *u,
			Size:      uint64(m.Size),
			CommandID: m.CommandID,
			UserID:    m.UserID,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		})
	}

	return commandMedias
}

func mapResult(cm models.ResultWithMedias) *entities.Result {
	result := &entities.Result{
		ID:           cm.ID,
		Text:         cm.Text,
		OpenRouterID: cm.OpenRouterID,
		CommandID:    cm.CommandID,
		ToolID:       cm.ToolID,
		ChatID:       cm.ChatID,
		UserID:       cm.UserID,
		ModelID:      cm.ModelID,
		CreatedAt:    cm.CreatedAt,
		UpdatedAt:    cm.UpdatedAt,
	}

	return result
}

func mapModel(cm models.ResultWithMedias) *entities.Model {
	model := &entities.Model{
		ID:        cm.Model.ID,
		Name:      cm.Model.Name,
		MinLevel:  cm.Model.MinLevel,
		CreatedAt: cm.Model.CreatedAt,
		UpdatedAt: cm.Model.UpdatedAt,
	}

	return model
}

func mapResultMedia(cm models.ResultWithMedias) []*entities.ResultMedia {
	resultMedia := make([]*entities.ResultMedia, 0)
	for _, m := range cm.Medias {
		u, err := url.Parse(m.URL)
		if err != nil {
			panic(err)
		}
		resultMedia = append(resultMedia, &entities.ResultMedia{
			ID:        m.ID,
			Name:      m.Name,
			Type:      m.Type,
			URL:       u,
			Size:      int64(m.Size),
			ResultID:  m.ResultID,
			UserID:    m.UserID,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		})
	}

	return resultMedia
}
