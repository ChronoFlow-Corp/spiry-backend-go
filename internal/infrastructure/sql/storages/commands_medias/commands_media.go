package commandsmedia

import (
	"context"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/google/uuid"
)

type CommandMediaStorage interface {
	Create(ctx context.Context, m entities.CommandMedia) error
	GetByID(ctx context.Context, mediaID uuid.UUID) (*entities.CommandMedia, error)
	Update(ctx context.Context, m entities.CommandMedia) error
	Delete(ctx context.Context, mediaID uuid.UUID) error
}
