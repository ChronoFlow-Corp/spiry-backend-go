package plans

import (
	"context"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/google/uuid"
)

type PlanStorage interface {
	Create(ctx context.Context, plan entities.Plan) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entities.Plan, error)
	GetByID(ctx context.Context, planID uuid.UUID) (*entities.Plan, error)
	Update(ctx context.Context, plan entities.Plan) error
}
