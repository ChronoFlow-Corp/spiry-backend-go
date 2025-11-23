package models

import (
	"context"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
)

type ModelStorage interface {
	Create(ctx context.Context, plan entities.Model) error
	GetByName(ctx context.Context, n string) (*entities.Model, error)
	GetAll(ctx context.Context) ([]entities.Model, error)
}
