package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository/entities"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func (p *Postgres) CreateChat(ctx context.Context, chat entities.Chat) error {
	const op = "repository.postgres.CreateChat"

	q := `INSERT INTO chats (id, title, user_id, tool_id) VALUES ($1, $2, $3, $4)`

	_, err := p.db.ExecContext(ctx, q, chat.ID, chat.Title, chat.UserID, chat.ToolID)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			return handleChatError(pgErr, chat)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *Postgres) UpdateChat(ctx context.Context, id uuid.UUID, title string) error {
	const op = "repository.postgres.UpdateChat"

	q := `UPDATE chats SET title = $1 WHERE id = $2`

	_, err := p.db.ExecContext(ctx, q, title, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *Postgres) DeleteChat(ctx context.Context, id uuid.UUID) error {
	const op = "repository.postgres.DeleteChat"

	q := `DELETE FROM chats WHERE id = $1`

	_, err := p.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *Postgres) GetChatByUserID(ctx context.Context) ([]entities.Chat, error) {
	const op = "repository.postgres.GetChatByID"

	id, ok := ctx.Value(entities.UserIDCtxKey{}).(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("%s: %w", op, errors.New("missing user ID in context"))
	}

	q := `SELECT * FROM chats WHERE user_id = $1`

	res, err := p.db.QueryContext(ctx, q, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer res.Close()

	chats := make([]entities.Chat, 0)

	for res.Next() {
		var chat dbChat

		err := res.Scan(&chat.ID, &chat.Title, &chat.UserID, &chat.CreatedAt, &chat.UpdatedAt, &chat.ToolID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		chat.Title = strings.TrimSpace(chat.Title)

		chats = append(chats, dbChatToRepositoryChat(chat))
	}

	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return chats, nil
}

func (p *Postgres) GetChatByID(ctx context.Context, id uuid.UUID) (entities.Chat, error) {
	const op = "repository.postgres.GetChatByID"

	q := `SELECT 
		chats.id as id,
		chats.title as title,
		chats.user_id as user_id,
		chats.created_at as chat_created_at,
		chats.updated_at as chat_updated_at,
		chats.tool_id as tool_id,
		messages.id as msg_id,
		messages.content as content,
		messages.role as role,
		messages.chat_id as chat_id,
		messages.user_id as msg_user_id,
		messages.updated_at as msg_updated_at,
        messages.created_at as msg_created_at
	FROM chats JOIN messages ON chats.id = messages.chat_id and chats.user_id = messages.user_id 
	WHERE chats.id = $1 and chats.user_id = $2`

	var rows []dbChatWithMessage

	err := p.db.SelectContext(ctx, &rows, q, id, ctx.Value(entities.UserIDCtxKey{}).(uuid.UUID))
	if err != nil {
		return entities.Chat{}, fmt.Errorf("%s: %w", op, err)
	}

	if len(rows) == 0 {
		return entities.Chat{}, nil
	}

	mpChat := make(map[uuid.UUID]*entities.Chat)
	for _, row := range rows {
		ch, ok := mpChat[row.ChatID]
		if !ok {
			ch = &entities.Chat{
				ID:        row.ChatID,
				Title:     strings.TrimSpace(row.Title),
				UserID:    row.UserID,
				CreatedAt: row.CreatedAt,
				UpdatedAt: row.UpdatedAt,
			}
			mpChat[row.ChatID] = ch
		}

		if row.MessageID != nil {
			ch.Messages = append(ch.Messages, entities.Message{
				ID:        *row.MessageID,
				Text:      strings.TrimSpace(row.Content),
				Role:      strings.TrimSpace(row.Role),
				ChatID:    ch.ID,
				UserID:    row.UserID,
				CreatedAt: row.MsgCreatedAt,
				UpdatedAt: row.MsgUpdatedAt,
			})
		}
	}

	res := entities.Chat{}
	for _, ch := range mpChat {
		res.ID = ch.ID
		res.Title = ch.Title
		res.UserID = ch.UserID
		res.ToolID = ch.ToolID
		res.CreatedAt = ch.CreatedAt
		res.UpdatedAt = ch.UpdatedAt
		res.Messages = append(res.Messages, ch.Messages...)
	}

	return res, nil
}

func handleChatError(err *pq.Error, chat entities.Chat) error {
	switch err.Code.Name() {
	case "unique_violation":
		return repository.NewUniqueViolation(err, "chat already exist", "id", chat.ID.String())
	case "foreign_key_violation":
		if err.Constraint == "chats_user_id_fkey" {
			return repository.NewNotFound(err, "user not found", "id", chat.UserID.String())
		}

		if err.Constraint == "chats_tool_id_fkey" {
			return repository.NewNotFound(err, "tool not found", "id", chat.ToolID.String())
		}
	}

	return err
}