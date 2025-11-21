package resultmedias

import (
	"context"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/google/uuid"
)

type ResultMediaStorage interface {
	Create(ctx context.Context, result entities.ResultMedia) error
	Update(ctx context.Context, rm entities.ResultMedia) error
	GetByID(ctx context.Context, cmID uuid.UUID) (*entities.ResultMedia, error)
	GetByURL(ctx context.Context, url string) (*entities.ResultMedia, error)
	GetByResultID(ctx context.Context, i uuid.UUID) ([]*entities.ResultMedia, error)
}
