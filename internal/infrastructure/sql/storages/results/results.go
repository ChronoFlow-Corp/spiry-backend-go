package results

import (
	"context"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/google/uuid"
)

type ResultStorage interface {
	Create(ctx context.Context, res entities.Result) error
	GetByCommandID(ctx context.Context, i uuid.UUID) (*entities.Result, error)
}
