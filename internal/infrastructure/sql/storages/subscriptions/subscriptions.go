package subscriptions

import (
	"context"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/google/uuid"
)

type SubscriptionStorage interface {
	Create(ctx context.Context, subscription entities.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Subscription, error)
	GetAll(ctx context.Context) ([]entities.Subscription, error)
}
