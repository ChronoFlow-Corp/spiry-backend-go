package chats

import (
	"context"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/aggregates"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/google/uuid"
)

type ChatStorage interface {
	Create(ctx context.Context, chat entities.Chat) error
	GetByID(ctx context.Context, chatID uuid.UUID) (*entities.Chat, error)
	Update(ctx context.Context, chat entities.Chat) error
	Delete(ctx context.Context, chatID uuid.UUID) error
	GetAll(ctx context.Context) ([]*entities.Chat, error)
	GetByIDWithCommandsResults(
		ctx context.Context,
		chatID uuid.UUID,
	) (*aggregates.Chat, error)
}
