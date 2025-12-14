package entities

import (
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/google/uuid"
)

// Plan is current user subscription.
type Plan struct {
	ID uuid.UUID

	// Quote will compute of plan created at and subscription Period.
	Quote          Quote
	SubscriptionID uuid.UUID
	Level          uint

	UserID    uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewPlan(quote Quote, level uint, subscriptionID, userID uuid.UUID) *Plan {
	return &Plan{
		ID:             uuid.New(),
		Quote:          quote,
		SubscriptionID: subscriptionID,
		Level:          level,
		UserID:         userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now().Add(time.Millisecond),
	}
}

func (p *Plan) Validate() error {
	if p.SubscriptionID == uuid.Nil {
		return domain.NewValidationError(nil, "subscription_id", "subscription_id is required")
	}

	if p.UserID == uuid.Nil {
		return domain.NewValidationError(nil, "user_id", "user_id is required")
	}

	if p.CreatedAt.After(p.UpdatedAt) {
		return domain.NewValidationError(nil, "created_at", "createdAt is in the future")
	}

	return nil
}
