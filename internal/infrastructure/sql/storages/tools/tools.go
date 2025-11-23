package tools

import (
	"context"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
)

type ToolStorage interface {
	Create(ctx context.Context, t entities.Tool) error
	GetByName(ctx context.Context, toolName string) (*entities.Tool, error)
	GetAll(ctx context.Context) ([]*entities.Tool, error)
}
