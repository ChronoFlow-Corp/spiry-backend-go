package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository/entities"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func (p *Postgres) AddMessage(ctx context.Context, msg entities.Message) error {
	const op = "repository.postgres.AddMessage"

	q := `INSERT INTO messages(id, content, role, chat_id, user_id) VALUES ($1, $2, $3, $4, $5)`

	_, err := p.db.ExecContext(ctx, q, msg.ID, msg.Text, msg.Role, msg.ChatID, msg.UserID)
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

func (p *Postgres) DeleteMessage(ctx context.Context, id uuid.UUID) error {
	const op = "repository.postgres.DeleteMessage"

	const q = `DELETE FROM messages WHERE id = $1 and user_id = $2`

	_, err := p.db.ExecContext(ctx, q, id, ctx.Value(entities.UserIDCtxKey{}).(uuid.UUID))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
