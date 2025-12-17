package entities

import (
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/google/uuid"
)

type Period string

// Time to expires subscription.
const (
	Day   Period = "day"
	Month Period = "month"
	Year  Period = "year"
)

const (
	maxLenSubscriptionName = 64
)

// Subscription is existing plan configuration.
type Subscription struct {
	ID        uuid.UUID
	Name      string
	Quote     Quote
	Period    Period
	Price     string
	Level     uint
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewSubscription(
	name string,
	period Period,
	quote Quote,
	price string,
	level uint,
) *Subscription {
	return &Subscription{
		ID:     uuid.New(),
		Name:   name,
		Quote:  quote,
		Period: period,
		Price:  price,
		Level:  level,
	}
}

func (s *Subscription) Validate() error {
	if s.Name == "" {
		return domain.NewValidationError(nil, "name", "name is required")
	}

	if len(s.Name) > maxLenSubscriptionName {
		return domain.NewValidationError(nil, "name", "name is too long")
	}

	if s.Period != Day && s.Period != Month && s.Period != Year {
		return domain.NewValidationError(nil, "Period", "invalid Period")
	}

	if s.CreatedAt.After(s.UpdatedAt) {
		return domain.NewValidationError(nil, "created_at", "createdAt is in the future")
	}

	return nil
}
