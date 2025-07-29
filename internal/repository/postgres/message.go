package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func (p *Postgres) AddMessage(ctx context.Context, chatID uuid.UUID, msg repository.Message) error {
	const op = "repository.postgres.AddMessage"

	q := `INSERT INTO messages(id, question, answer, chat_id, user_id) VALUES ($1, $2, $3, $4, $5)`

	_, err := p.db.ExecContext(ctx, q, msg.ID, msg.Question, msg.Answer, chatID, msg.UserID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Constraint == "messages_chat_id_fkey" {
				return repository.NewNotFound(err, "chat not found", "chatID", msg.ChatID.String())
			}

			if pqErr.Constraint == "messages_user_id_fkey" {
				return repository.NewNotFound(err, "user not found", "userID", msg.UserID.String())
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
