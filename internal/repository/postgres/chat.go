package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository"
	"github.com/lib/pq"
)

func (p *Postgres) CreateChat(ctx context.Context, chat repository.Chat) error {
	const op = "repository.postgres.CreateChat"

	q := `INSERT INTO chats (id, title, user_id) VALUES ($1, $2, $3)`

	_, err := p.db.ExecContext(ctx, q, chat.ID, chat.Title, chat.UserID)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			return handleChatError(pgErr, chat)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func handleChatError(err *pq.Error, chat repository.Chat) error {
	switch err.Code.Name() {
	case "unique_violation":
		return repository.NewUniqueViolation(err, "chat already exist", "id", chat.ID.String())
	case "foreign_key_violation":
		if err.Constraint == "chats_user_id_fkey" {
			return repository.NewNotFound(err, "user not found", "id", chat.UserID.String())
		}
	}

	return err
}
