package commands

import (
	"context"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/google/uuid"
)

type CommandStorage interface {
	Create(ctx context.Context, command entities.Command) error
	GetByID(ctx context.Context, commandID uuid.UUID) (*entities.Command, error)
	Update(ctx context.Context, command entities.Command) error
	Delete(ctx context.Context, commandID uuid.UUID) error
}
